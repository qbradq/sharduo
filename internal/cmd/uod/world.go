package uod

import (
	"io"
	"log"
	"sync"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// File lock error
var ErrSaveFileLocked = errors.New("the save file is currently locked")
var ErrInitializationInProgress = errors.New("world initialization seems to be in progress")

// World encapsulates all of the data for the world and the goroutine that
// manipulates it.
type World struct {
	// The world map
	m *game.Map
	// The data store of the user accounts
	ads *util.DataStore[*game.Account]
	// Index of usernames ot account serials
	aidx map[string]uo.Serial
	// The object data store for the entire world
	ods *util.DataStore[game.Object]
	// The random number generator for the world
	rng *util.RNG
	// Inbound requests
	requestQueue chan WorldRequest
	// Save/Load Mutex
	lock sync.Mutex
	// Save directory string
	savePath string
}

// NewWorld creates a new, empty world
func NewWorld(savePath string) *World {
	rng := util.NewRNG()
	return &World{
		m:            game.NewMap(),
		ads:          util.NewDataStore[*game.Account]("accounts", rng, game.ObjectFactory),
		aidx:         make(map[string]uo.Serial),
		ods:          util.NewDataStore[game.Object]("objects", rng, game.ObjectFactory),
		rng:          rng,
		requestQueue: make(chan WorldRequest, 1024),
		savePath: savePath,
	}
}

// latestSavePath returns the path to the most recent save file or directory
func (w *World) latestSavePath() string {
	entries, err := os.ReadDir(w.savePath)
	if err != nil {
		return ""
	}
	if len(entries) == 0 {
		return ""
	}
	latestIdx := -1
	var latestModTime time.Time
	for i, e := range entries {
		info, err := e.Info()
		if err != nil {
			return ""
		}
		if info.IsDir() {
			continue
		}
		if info.ModTime().After(latestModTime) {
			latestModTime = info.ModTime()
			latestIdx = i
		}
	}
	e := entries[latestIdx]
	return path.Join(w.savePath, e.Name())
}

// reportErrors logs all errors in the slice, then returns a single error with
// a summary report.
func (w *World) reportErrors(errs []error) error {
	for _, err := range errs {
		log.Println(err)
	}
	return fmt.Errorf("%d errors reported", len(errs))
}

// Load loads all of the data stores that the world is responsible for
func (w *World) Load(r io.Reader) error {
	var errs []error

	if !w.lock.TryLock() {
		return ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	sf := util.NewCompressedSaveFileReader()
	filePath := w.latestSavePath()
	if filePath == "" {
		return os.ErrNotExists
	}
	if err := sf.Open(filePath); err {
		return err
	}

	r, err := sf.GetReader("accounts.ini")	
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ads.Read(r)...)
	}

	r, err = sf.GetReader("objects.ini")	
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ods.Read(r)...)
	}

	return w.reportErrors(errs)
}

// getFileName returns the file name portion of the save path without extension.
func (m *SaveManager) getFileName() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

// Save saves all of the data stores that the world is responsible for
func (w *World) Save() error {
	var errs []error

	if !w.lock.TryLock() {
		return ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	// Initialization in progress
	if w == nil {
		return ErrInitializationInProgress
	}

	// Make sure the save directory is present
	if err := os.MkdirAll(m.savePath, 0777); err != nil {
		log.Println("error creating save directory", err)
		return err
	}

	// Try to create the output file
	filePath := path.Join(w.savePath, w.getFileName()+".zip")
	z, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer z.Close()
	log.Println("saving data stores to", filePath)

	// Create the zip writer
	w := zip.NewWriter(z)
	defer w.Close()

	// Save data sets
	f, err := w.Create("accounts.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ads.Save(w)...)
	}
	f.Close()

	f, err = w.Create("objects.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ads.Save(w)...)
	}
	f.Close()
}

// SendRequest sends a WorldRequest to the world's goroutine. Returns true if
// the command was successfully queued. This never blocks.
func (w *World) SendRequest(cmd WorldRequest) bool {
	select {
	case w.requestQueue <- cmd:
		return true
	default:
		return false
	}
}

// Random returns the *util.RNG the world is using for sync operations
func (w *World) Random() *util.RNG {
	return w.rng
}

// New adds a new object to the world. It is assigned a unique serial. The
// object is returned.
func (w *World) New(o game.Object) game.Object {
	w.ods.Add(o, o.GetSerialType())
	return o
}

// AuthenticateAccount attempts to authenticate an account by username and
// password hash. If no account exists for that username, a new one will be
// created for the user. If an account is found but the password hashes do not
// match nil is returned. Otherwise the account is returned.
func (w *World) AuthenticateAccount(username, passwordHash string) *Account {
	a := w.getOrCreateAccount(username, passwordHash)
	if a.passwordHash != passwordHash {
		return nil
	}
	return a
}

// getOrCreateAccount adds a new account to the world, or returns the existing
// account for that username.
func (w *World) getOrCreateAccount(username, passwordHash) *Account {
	if s, ok := w.aidx[username]; ok {
		return w.ads.Get(s)
	}
	a := &Account{
		Username: username,
		PasswordHash: passwordHash,
	}
	w.ads.Add(a)
	w.aidx[username] = a
	return a
}

// Process is the goroutine that services the command queue and is the only
// goroutine allowed to interact with the contents of the world.
func (w *World) Process() {
	for r := range w.requestQueue {
		if err := r.Execute(); err != nil {
			log.Println(err)
		}
	}
}
