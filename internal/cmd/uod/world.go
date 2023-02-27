package uod

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/datastore"
	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
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
	// The object data store for the entire world
	ods *datastore.T[game.Object]
	// Collection of all accounts
	accounts map[string]*game.Account
	// Account access mutex
	alock sync.Mutex
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
		ods:           datastore.NewDataStore[game.Object](rng),
		accounts:      make(map[string]*game.Account),
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

// Unmarshal reads in all of the data stores we are responsible for.
func (w *World) Unmarshal() error {
	if !w.lock.TryLock() {
		return ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	start := time.Now()
	filePath := w.latestSavePath()
	gzf, err := os.Open(filePath)
	if err != nil {
		if strings.Contains(err.Error(), "handle is invalid") {
			return os.ErrNotExist
		}
		return err
	}
	gz, err := gzip.NewReader(gzf)
	if err != nil {
		return err
	}
	d, err := io.ReadAll(gz)
	if d == nil {
		return os.ErrNotExist
	} else if err != nil {
		return err
	}
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("read save file into memory in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)

	// Global data
	start = time.Now()
	tf := marshal.NewTagFile(d)
	s := tf.Segment(marshal.SegmentWorld)
	w.time = uo.Time(s.Long())
	// Timers
	game.UnmarshalTimers(tf.Segment(marshal.SegmentTimers))
	// Create objects in data stores
	s = tf.Segment(marshal.SegmentObjectList)
	for i := uint32(0); i < s.RecordCount(); i++ {
		serial := uo.Serial(s.Int())
		otype := marshal.ObjectType(s.Byte())
		ctor := marshal.Constructor(otype)
		if ctor == nil {
			return fmt.Errorf("error: object %s of type %d constructor not found", serial.String(), otype)
		}
		v := ctor()
		o, ok := v.(game.Object)
		if !ok {
			return fmt.Errorf("error: object %s of type %d does not implement the Object interface", serial.String(), otype)
		}
		o.SetSerial(serial)
		w.ods.Insert(o)
	}
	// Unmarshal objects in the datastore
	seg := marshal.SegmentObjectsStart
	for {
		s := tf.Segment(seg)
		if s.IsEmpty() {
			break
		}
		w.ods.UnmarshalObjects(s)
		seg++
	}
	// Call the AfterUnmarshal hook on all objects in the datastore
	w.ods.AfterUnmarshalObjects()
	// Call RecalculateStats on all objects in the datastore
	for _, o := range w.ods.Data() {
		o.RecalculateStats()
	}
	// Let the map accumulate all of it's child objects
	w.m.Unmarshal(tf.Segment(marshal.SegmentMap))
	// Rebuild accounts dataset
	s = tf.Segment(marshal.SegmentAccounts)
	for i := uint32(0); i < s.RecordCount(); i++ {
		a := &game.Account{}
		a.Unmarshal(s)
		w.accounts[a.Username()] = a
	}
	// Done
	end = time.Now()
	elapsed = end.Sub(start)
	log.Printf("save unmarshaled in %ds%03d", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)

	return nil
}

// getFileName returns the file name portion of the save path without extension.
func (w *World) getFileName() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

// Marshal writes all of the data stores that the world is responsible for. A
// WaitGroup is returned to wait for the file to be written to disk.
func (w *World) Marshal() (*sync.WaitGroup, error) {
	if !w.lock.TryLock() {
		return nil, ErrSaveFileLocked
	}
	defer w.lock.Unlock()

	filePath := path.Join(w.savePath, w.getFileName()+".sav.gz")
	filePath = path.Clean(filePath)
	os.MkdirAll(path.Dir(filePath), 0777)
	log.Printf("saving data stores to %s", filePath)

	start := time.Now()
	// We have to build a slice of all of our objects so we don't have
	// concurrency issues on the data store map during the multi-goroutine save
	objects := make([]game.Object, len(w.ods.Data()))
	idx := 0
	for _, obj := range w.ods.Data() {
		objects[idx] = obj
		idx++
	}
	wg := &sync.WaitGroup{}
	tf := marshal.NewTagFile(nil)
	// Global data
	s := tf.Segment(marshal.SegmentWorld)
	wg.Add(1)
	go func(s *marshal.TagFileSegment) {
		defer wg.Done()
		s.PutLong(uint64(w.time))
	}(s)
	// Timers
	s = tf.Segment(marshal.SegmentTimers)
	wg.Add(1)
	go func(s *marshal.TagFileSegment) {
		// Raw data for timers, this shouldn't change anymore
		defer wg.Done()
		game.MarshalTimers(s)
	}(s)
	// Accounting data
	s = tf.Segment(marshal.SegmentAccounts)
	wg.Add(1)
	go func(s *marshal.TagFileSegment) {
		defer wg.Done()
		for _, a := range w.accounts {
			a.Marshal(s)
			s.IncrementRecordCount()
		}
	}(s)
	// Object list
	s = tf.Segment(marshal.SegmentObjectList)
	wg.Add(1)
	go func(s *marshal.TagFileSegment) {
		defer wg.Done()
		for _, o := range objects {
			s.PutInt(uint32(o.Serial()))
			s.PutByte(byte(o.ObjectType()))
			s.IncrementRecordCount()
		}
	}(s)
	// Kick off the object database persistance goroutines
	w.ods.MarshalObjects(tf, 4, wg)
	// Map object list
	wg.Add(1)
	s = tf.Segment(marshal.SegmentMap)
	go func(s *marshal.TagFileSegment) {
		// Raw data for map object list
		defer wg.Done()
		for _, o := range objects {
			if o.Parent() == nil {
				s.PutInt(uint32(o.Serial()))
				s.IncrementRecordCount()
			}
		}
	}(s)
	// The main goroutine is blocked at this point
	wg.Wait()
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("generated save data in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)

	// Kick off another goroutine to persist the save to disk and let the main
	// goroutine continue.
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		f, err := os.Create(filePath)
		if err != nil {
			log.Printf("error: unable to create save file %s", filePath)
			return
		}
		zw := gzip.NewWriter(f)
		zw.Name = filePath
		zw.ModTime = start
		zw.Comment = "ShardUO save file"
		tf.Output(zw)
		zw.Close()
		tf.Close()
		end := time.Now()
		elapsed := end.Sub(start)
		log.Printf("saved file to disk in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	}()

	return wg, nil
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
	w.alock.Lock()
	defer w.alock.Unlock()

	a := w.accounts[username]
	newAccount := false
	if a == nil {
		a = game.NewAccount(username, passwordHash)
		newAccount = true
	}
	if !a.ComparePasswordHash(passwordHash) {
		return nil
	}
	if newAccount {
		w.accounts[username] = a
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
