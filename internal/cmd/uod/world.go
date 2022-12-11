package uod

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// File lock error
var ErrSaveFileLocked = errors.New("the save file is currently locked")

// File truncation error
var ErrSaveFileExists = errors.New("refusing to truncate existing save file")

// World encapsulates all of the data for the world and the goroutine that
// manipulates it.
type World struct {
	// The world map
	m *game.Map
	// The data store of the user accounts
	ads *util.DataStore[*game.Account]
	// Index of usernames ot account serials
	aidx map[string]uo.Serial
	// Accounting access lock
	alock sync.Mutex
	// The object data store for the entire world
	ods *util.DataStore[game.Object]
	// The random number generator for the world
	rng uo.RandomSource
	// Inbound requests
	requestQueue chan WorldRequest
	// Save/Load Mutex
	lock sync.Mutex
	// Save directory string
	savePath string
	// TargetManager for the world
	tm *TargetManager
}

// NewWorld creates a new, empty world
func NewWorld(savePath string, rng uo.RandomSource) *World {
	return &World{
		m:            game.NewMap(),
		ads:          util.NewDataStore[*game.Account]("accounts", rng, game.ObjectFactory),
		aidx:         make(map[string]uo.Serial),
		ods:          util.NewDataStore[game.Object]("objects", rng, game.ObjectFactory),
		rng:          rng,
		requestQueue: make(chan WorldRequest, 1024),
		savePath:     savePath,
		tm:           NewTargetManager(rng),
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
	if len(errs) == 0 {
		return nil
	}
	for _, err := range errs {
		log.Println(err)
	}
	return fmt.Errorf("%d errors reported", len(errs))
}

// Load loads all of the data stores that the world is responsible for.
// ErrSaveFileLocked is returned when another goroutine is trying to save or
// load. os.ErrNotExist is returned when there are no saves available to load.
// nil is returned on success. An error describing how many errors were
// encountered is returned if there are any.
func (w *World) Load() error {
	var errs []error

	if !w.lock.TryLock() {
		return ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	sfr := util.NewDebugSaveFileReader()
	filePath := w.latestSavePath()
	if filePath == "" {
		return os.ErrNotExist
	}
	if err := sfr.Open(filePath); err != nil {
		return err
	}
	log.Println("loading data stores from", filePath)

	// Load account objects
	r, err := sfr.GetReader("accounts.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ads.Read(r)...)
	}
	r.Close()

	// Load objects
	r, err = sfr.GetReader("objects.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ods.Read(r)...)
	}
	r.Close()

	// If there were errors trying to load objects into the data stores we
	// should just stop now. There will be tons of dereferencing errors if we
	// try to continue with deserialization.
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}
		log.Fatalf("found %d errors while allocating data store objects", len(errs))
	}

	// Deserialize all data stores
	errs = append(errs, w.ods.Deserialize()...)
	errs = append(errs, w.ads.Deserialize()...)

	// If there were errors trying to deserialize any of our objects we need to
	// report and bail.

	// Rebuild accounts index
	w.ads.Map(func(a *game.Account) error {
		w.aidx[a.Username()] = a.Serial()
		return nil
	})

	return w.reportErrors(errs)
}

// getFileName returns the file name portion of the save path without extension.
func (w *World) getFileName() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

// Save saves all of the data stores that the world is responsible for
func (w *World) Save() error {
	var errs []error

	if !w.lock.TryLock() {
		return ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	sfw := util.NewDebugSaveFileWriter()
	filePath := path.Join(w.savePath, w.getFileName())
	if err := sfw.Open(filePath); err != nil {
		return err
	}
	log.Println("saving data stores to", filePath)

	// Save accounts
	f, err := sfw.GetWriter("accounts.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ads.Write(f)...)
	}

	// Save objects
	f, err = sfw.GetWriter("objects.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ods.Write(f)...)
	}

	return w.reportErrors(errs)
}

// SendTarget sends a targeting request to the client.
func (w *World) SendTarget(n *NetState, ttype uo.TargetType, ctx interface{}, fn TargetCallback) {
	t := w.tm.New(&Target{
		NetState: n,
		Callback: fn,
		Context:  ctx,
		TTL:      uo.DurationSecond * 30,
	})
	n.Send(&serverpacket.Target{
		Serial:     t.Serial(),
		TargetType: ttype,
		CursorType: uo.CursorTypeNeutral,
	})
}

// ExecuteTarget executes a targeting response. Returns true if the target
// request was still pending and executed.
func (w *World) ExecuteTarget(r *clientpacket.TargetResponse) bool {
	return w.tm.Execute(r)
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

// Random returns the uo.RandomSource the world is using for sync operations
func (w *World) Random() uo.RandomSource {
	return w.rng
}

// New adds a new object to the world. It is assigned a unique serial. The
// object is returned.
func (w *World) New(o game.Object) game.Object {
	w.ods.Add(o, o.SerialType())
	return o
}

// AuthenticateAccount attempts to authenticate an account by username and
// password hash. If no account exists for that username, a new one will be
// created for the user. If an account is found but the password hashes do not
// match nil is returned. Otherwise the account is returned.
func (w *World) AuthenticateAccount(username, passwordHash string) *game.Account {
	a := w.GetOrCreateAccount(username, passwordHash)
	if !a.ComparePasswordHash(passwordHash) {
		return nil
	}
	return a
}

// GetOrCreateAccount adds a new account to the world, or returns the existing
// account for that username.
func (w *World) GetOrCreateAccount(username, passwordHash string) *game.Account {
	w.alock.Lock()
	defer w.alock.Unlock()

	if s, ok := w.aidx[username]; ok {
		return w.ads.Get(s)
	}
	a := game.NewAccount(username, passwordHash)
	w.ads.Add(a, uo.SerialTypeUnbound)
	w.aidx[username] = a.Serial()
	return a
}

// AuthenticateLoginSession attempts to find and authenticate the account
// associated with the username and serial. nil is returned if the account is
// not found or the password hashes do not match.
func (w *World) AuthenticateLoginSession(username, passwordHash string, id uo.Serial) *game.Account {
	w.alock.Lock()
	defer w.alock.Unlock()

	a := w.ads.Get(id)
	if a == nil || a.Username() != username || !a.ComparePasswordHash(passwordHash) {
		return nil
	}
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
