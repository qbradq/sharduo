package uod

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/marshal"
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
	// Index of usernames to account serials
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
	// Collection of all objects that need to be updated
	updateList map[uo.Serial]game.Object
	// Current time in the Sossarian universe in seconds
	time uo.Time
	// Current time on the server
	wallClockTime time.Time
}

// NewWorld creates a new, empty world
func NewWorld(savePath string, rng uo.RandomSource) *World {
	return &World{
		m:             game.NewMap(),
		ads:           util.NewDataStore[*game.Account]("accounts", rng, game.ObjectFactory),
		aidx:          make(map[string]uo.Serial),
		ods:           util.NewDataStore[game.Object]("objects", rng, game.ObjectFactory),
		rng:           rng,
		requestQueue:  make(chan WorldRequest, 1024*16),
		savePath:      savePath,
		updateList:    make(map[uo.Serial]game.Object),
		time:          uo.TimeEpoch,
		wallClockTime: time.Now(),
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

// loadGlobalData loads all global data for the world.
func (w *World) loadGlobalData(r io.Reader) []error {
	tfr := &util.TagFileReader{}
	tfr.StartReading(r)
	tfo := tfr.ReadObject()
	if tfo == nil {
		return append(tfr.Errors(), errors.New("unable to load world object"))
	}
	w.time = uo.Time(tfo.GetULong("Time", uint64(uo.TimeEpoch)))
	return tfo.Errors()
}

// Load loads all of the data stores that the world is responsible for.
// ErrSaveFileLocked is returned when another goroutine is trying to save or
// load. os.ErrNotExist is returned when there are no saves available to load.
// nil is returned on success. An error describing how many errors were
// encountered is returned if there are any.
func (w *World) Load() error {
	var errs []error
	start := time.Now()

	pf, err := os.Create("load.cpu.pprof")
	if err != nil {
		return err
	}
	if err := pprof.StartCPUProfile(pf); err != nil {
		return err
	}
	defer pprof.StopCPUProfile()

	if !w.lock.TryLock() {
		return ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	filePath := w.latestSavePath()
	sfr, err := util.NewSaveFileReaderByPath(filePath)
	if err != nil {
		return err
	}
	if err := sfr.Open(filePath); err != nil {
		return err
	}
	log.Println("loading data stores from", filePath)

	// Load global data
	r, err := sfr.GetReader("world.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.loadGlobalData(r)...)
	}
	r.Close()

	// Load account objects
	r, err = sfr.GetReader("accounts.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ads.Read(r)...)
	}
	r.Close()

	// Load game objects
	r, err = sfr.GetReader("objects.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ods.Read(r)...)
	}
	r.Close()

	// Load timers
	r, err = sfr.GetReader("timers.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, game.ReadTimers(r)...)
	}

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
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}
		log.Fatalf("found %d errors while deserializing data store objects", len(errs))
	}

	// Place all objects on the map
	r, err = sfr.GetReader("map.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.m.Read(r)...)
	}
	r.Close()

	// Call the deserialize hooks for all data stores
	errs = append(errs, w.ods.OnAfterDeserialize()...)
	errs = append(errs, w.ads.OnAfterDeserialize()...)

	// If there were errors during the deserialize hooks we need to report and
	// bail.
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}
		log.Fatalf("found %d errors while running deserialization hooks", len(errs))
	}

	// Call RecalculateStats for every object in the object data store
	w.ods.Map(func(o game.Object) error {
		o.RecalculateStats()
		return nil
	})

	// Rebuild accounts index
	w.ads.Map(func(a *game.Account) error {
		w.aidx[a.Username()] = a.Serial()
		return nil
	})

	if len(errs) == 0 {
		end := time.Now()
		elapsed := end.Sub(start).Round(time.Millisecond)
		log.Printf("load operation complete in %ds%dms",
			elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	}

	return w.reportErrors(errs)
}

// Unmarshal reads in all of the data stores we are responsible for.
func (w *World) Unmarshal() error {
	start := time.Now()
	filePath := w.latestSavePath()
	d, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("read save file into memory in %ds%03d", elapsed.Microseconds()/1000, elapsed.Microseconds()%1000)

	// Global data
	start = time.Now()
	tf := marshal.NewTagFile(d)
	s := tf.Segment(marshal.SegmentWorld)
	w.time = uo.Time(s.Long())
	// Timers
	game.UnmarshalTimers(tf.Segment(marshal.SegmentTimers))
	// Create objects in data stores
	w.ads.LoadMarshalData(tf.Segment(marshal.SegmentAccounts))
	seg := marshal.SegmentObjectsStart
	for {
		s := tf.Segment(seg)
		if s.IsEmpty() {
			break
		}
		w.ods.LoadMarshalData(s)
	}
	// Unmarshal objects in the datastores
	w.ads.UnmarshalObjects()
	w.ods.UnmarshalObjects()
	// Call the AfterUnmarshal hook on all objects in the datastores
	w.ads.AfterUnmarshalObjects()
	w.ods.AfterUnmarshalObjects()
	// Let the map accumulate all of it's child objects
	w.m.Unmarshal(tf.Segment(marshal.SegmentMap))
	// Done
	end = time.Now()
	elapsed = end.Sub(start)
	log.Printf("save unmarshaled in %ds%03d", elapsed.Microseconds()/1000, elapsed.Microseconds()%1000)

	return nil
}

// getFileName returns the file name portion of the save path without extension.
func (w *World) getFileName() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

// saveGlobalData saves all global data in tag file format.
func (w *World) saveGlobalData(writer io.WriteCloser) []error {
	tfw := util.NewTagFileWriter(writer)
	tfw.WriteComment("global world data")
	tfw.WriteBlankLine()
	tfw.WriteSegmentHeader("Globals")
	tfw.WriteULong("Time", uint64(w.time))
	tfw.WriteBlankLine()
	tfw.WriteComment("END OF FILE")
	return nil
}

// Save saves all of the data stores that the world is responsible for
func (w *World) Save() error {
	var errs []error
	start := time.Now()

	pf, err := os.Create("save.cpu.pprof")
	if err != nil {
		return err
	}
	if err := pprof.StartCPUProfile(pf); err != nil {
		return err
	}
	defer pprof.StopCPUProfile()

	if !w.lock.TryLock() {
		return ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	sfw, err := util.NewSaveFileWriter(configuration.GameSaveType)
	if err != nil {
		return err
	}
	filePath := path.Join(w.savePath, w.getFileName())
	if err := sfw.Open(filePath); err != nil {
		return err
	}
	log.Println("saving data stores to", filePath)

	// Save global data
	f, err := sfw.GetWriter("world.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.saveGlobalData(f)...)
	}
	f.Close()

	// Save accounts
	f, err = sfw.GetWriter("accounts.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ads.Write(f)...)
	}
	f.Close()

	// Save objects
	f, err = sfw.GetWriter("objects.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.ods.Write(f)...)
	}
	f.Close()

	// Save timers
	f, err = sfw.GetWriter("timers.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, game.WriteTimers(f)...)
	}
	f.Close()

	// Save the map
	f, err = sfw.GetWriter("map.ini")
	if err != nil {
		errs = append(errs, err)
	} else {
		errs = append(errs, w.m.Write(f)...)
	}
	f.Close()

	if len(errs) == 0 {
		end := time.Now()
		elapsed := end.Sub(start).Round(time.Millisecond)
		log.Printf("save operation complete in %ds%dms",
			elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	}

	sfw.Close()

	if len(errs) == 0 {
		end := time.Now()
		elapsed := end.Sub(start).Round(time.Millisecond)
		log.Printf("save operation with flush to disk complete in %ds%dms",
			elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	}

	return w.reportErrors(errs)
}

// Marshal writes all of the data stores that the world is responsible for.
func (w *World) Marshal() error {
	if !w.lock.TryLock() {
		return ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	filePath := path.Join(w.savePath, w.getFileName()+".sav")
	filePath = path.Clean(filePath)
	os.MkdirAll(path.Dir(filePath), 0777)
	log.Printf("saving data stores to %s", filePath)

	pf, err := os.Create("marshal.cpu.pprof")
	if err != nil {
		return err
	}
	if err := pprof.StartCPUProfile(pf); err != nil {
		return err
	}

	start := time.Now()
	wg := &sync.WaitGroup{}
	tf := marshal.NewTagFile(nil)
	s := tf.Segment(marshal.SegmentWorld)
	wg.Add(1)
	go func(s *marshal.TagFileSegment) {
		s.PutLong(uint64(w.time))
		wg.Done()
	}(s)
	wg.Add(1)
	go func() {
		game.MarshalTimers(tf.Segment(marshal.SegmentTimers))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		game.MarshalAccounts(tf.Segment(marshal.SegmentAccounts), w.ads.Data())
		wg.Done()
	}()
	saveGoroutines := 4
	for i := 0; i < saveGoroutines; i++ {
		s := tf.Segment(marshal.SegmentObjectsStart + marshal.Segment(i))
		d := w.ods.Data()
		pool := i
		wg.Add(1)
		go func() {
			game.MarshalObjects(s, d, saveGoroutines, pool)
			wg.Done()
		}()
	}
	wg.Add(1)
	go func() {
		w.m.Marshal(tf.Segment(marshal.SegmentMap))
		wg.Done()
	}()
	wg.Wait()
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("generated save data in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)

	pprof.StopCPUProfile()

	start = time.Now()
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	tf.Output(f)
	f.Close()
	tf.Close()
	end = time.Now()
	elapsed = end.Sub(start)
	log.Printf("saved file to disk in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)

	return nil
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

// addNewObjectToDataStores adds a new object to the world data stores. It is
// assigned a unique serial. The object is returned. As a special case this
// function refuses to add a nil value to the game data store.
func (w *World) addNewObjectToDataStores(o game.Object) game.Object {
	if o != nil {
		w.ods.Add(o, o.SerialType())
	}
	return o
}

// New implements the game.World interface.
func (w *World) New(templateName string) game.Object {
	o := templateManager.newObject(templateName)
	if o != nil {
		w.addNewObjectToDataStores(o)
		o.SetParent(game.TheVoid)
	}
	return o
}

// Find returns the object with the given serial, or nil if it is not found in
// the game data store.
func (w *World) Find(id uo.Serial) game.Object {
	return w.ods.Get(id)
}

// Remove implements the game.World interface.
func (w *World) Remove(o game.Object) {
	p := o.Parent()
	if p == nil {
		for _, m := range w.m.GetNetStatesInRange(o.Location(), uo.MaxViewRange) {
			m.NetState().RemoveObject(o)
		}
		w.m.ForceRemoveObject(o)
	} else {
		p.ForceRemoveObject(o)
	}
	o.SetParent(game.TheVoid)
	w.ods.Remove(o)
}

// AuthenticateAccount attempts to authenticate an account by username and
// password hash. If no account exists for that username, a new one will be
// created for the user. If an account is found but the password hashes do not
// match nil is returned. Otherwise the account is returned.
func (w *World) AuthenticateAccount(username, passwordHash string) *game.Account {
	a := w.getOrCreateAccount(username, passwordHash)
	if !a.ComparePasswordHash(passwordHash) {
		return nil
	}
	return a
}

// getOrCreateAccount adds a new account to the world, or returns the existing
// account for that username.
func (w *World) getOrCreateAccount(username, passwordHash string) *game.Account {
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

// Time implements the game.World interface.
func (w *World) Time() uo.Time { return w.time }

// ServerTime implements the game.World interface.
func (w *World) ServerTime() time.Time { return w.wallClockTime }

// Stop attempts to gracefully shut down the world process.
func (w *World) Stop() {
	close(w.requestQueue)
}

// Main is the goroutine that services the command queue and is the only
// goroutine allowed to interact with the contents of the world.
func (w *World) Main(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	var done bool
	ticker := time.NewTicker(time.Second / time.Duration(uo.DurationSecond))
	for !done {
		select {
		case t := <-ticker.C:
			// The ticker has a higher priority than packets. This should ensure
			// that the game service cannot be overwhelmed with packets and not
			// be able to do cleanup tasks.
			// TODO detect dropped ticks
			// Time handling
			w.time++
			w.wallClockTime = t
			// Update timers
			game.UpdateTimers(w.time)
			// Interleave net state updates
			UpdateNetStates(int(w.time % uo.DurationSecond))
			// Update objects
			for _, o := range w.updateList {
				for _, m := range w.m.GetNetStatesInRange(o.Location(), uo.MaxViewRange) {
					if o.Location().XYDistance(m.Location()) <= m.ViewRange() {
						m.NetState().UpdateObject(o)
					}
				}
			}
			w.updateList = make(map[uo.Serial]game.Object)
		case r := <-w.requestQueue:
			// TODO Graceful shutdown signal (outside this struct)
			// Handle graceful shutdown
			if r == nil {
				ticker.Stop()
				done = true
				break
			}
			// If we are not trying to handle a tick we process packets.
			if err := r.Execute(); err != nil {
				if r.GetNetState() != nil {
					r.GetNetState().Speech(nil, err.Error())
				}
				log.Println(err)
			}
		}
	}
}

// Map returns the map the world is using.
func (w *World) Map() *game.Map {
	return w.m
}

// GetItemDefinition returns the uo.StaticDefinition that holds the static data
// for a given item graphic.
func (w *World) GetItemDefinition(g uo.Graphic) *uo.StaticDefinition {
	return tiledatamul.GetStaticDefinition(int(g))
}

// Update implements the game.World interface.
func (w *World) Update(o game.Object) {
	w.updateList[o.Serial()] = o
}

// BroadcastPacket implements the game.World interface.
func (w *World) BroadcastPacket(p serverpacket.Packet) {
	BroadcastPacket(p)
}

// BroadcastMessage implements the game.World interface.
func (w *World) BroadcastMessage(speaker game.Object, fmtstr string, args ...interface{}) {
	sid := uo.SerialSystem
	body := uo.BodySystem
	font := uo.FontNormal
	hue := uo.Hue(1153)
	name := ""
	text := fmt.Sprintf(fmtstr, args...)
	stype := uo.SpeechTypeSystem
	if speaker != nil {
		sid = speaker.Serial()
		stype = uo.SpeechTypeNormal
		name = speaker.DisplayName()
		if item, ok := speaker.(game.Item); ok {
			body = uo.Body(item.BaseGraphic())
		} else if mob, ok := speaker.(game.Mobile); ok {
			body = mob.Body()
		}
	}
	w.BroadcastPacket(&serverpacket.Speech{
		Speaker: sid,
		Body:    body,
		Font:    font,
		Hue:     hue,
		Name:    name,
		Text:    text,
		Type:    stype,
	})
}
