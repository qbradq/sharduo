package uod

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/qbradq/sharduo/internal/configuration"
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

// worldPacket is a value that joins a client packet with the net state it came
// in on.
type worldPacket struct {
	Packet   clientpacket.Packet // The packet to process
	NetState *NetState           // The net state it came in on
}

// World encapsulates all of the data for the world and the goroutine that
// manipulates it.
type World struct {
	m             *game.Map                // The world map
	ods           *game.Datastore          // The object data store for the entire world
	accounts      map[string]*game.Account // Collection of all accounts
	al            sync.Mutex               // Account access mutex
	requestQueue  chan worldPacket         // Inbound requests
	lock          sync.Mutex               // Save/Load Mutex
	savePath      string                   // Save directory string
	updateList    map[uo.Serial]struct{}   // Collection of all objects that need to be updated
	oplUpdateList map[uo.Serial]struct{}   // Collection of all objects that need to have OPLInfo packets sent
	time          uo.Time                  // Current time in the Sossarian universe in seconds
	wallClockTime time.Time                // Current time on the server
	superUser     *game.Account            // Pointer to the super-user account
}

// NewWorld creates a new, empty world
func NewWorld(savePath string) *World {
	return &World{
		m:             game.NewMap(),
		ods:           game.NewDatastore(),
		accounts:      make(map[string]*game.Account),
		requestQueue:  make(chan worldPacket, 1024*16),
		savePath:      savePath,
		updateList:    make(map[uo.Serial]struct{}),
		oplUpdateList: make(map[uo.Serial]struct{}),
		time:          uo.TimeEpoch,
		wallClockTime: time.Now(),
	}
}

// LatestSavePath returns the path to the most recent save file or directory
func (w *World) LatestSavePath() string {
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
	fp := w.LatestSavePath()
	if fp == "" {
		return os.ErrNotExist
	}
	tf, err := util.ZipRead(fp)
	if err != nil {
		log.Fatalf("fatal: failed reading save file %s", fp)
	}
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("info: read save file into memory in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	// Global data
	start = time.Now()
	s := tf["global"]
	nThreads := int(util.GetUInt32(s))
	w.time = uo.Time(util.GetUInt64(s))
	// Timers
	game.ReadTimers(tf["timers"])
	// Accounts
	s = tf["accounts"]
	n := int(util.GetUInt32(s))
	for idx := 0; idx < n; idx++ {
		a := &game.Account{}
		a.Read(s)
		w.accounts[a.Username] = a
		if a.HasRole(game.RoleSuperUser) {
			w.superUser = a
		}
	}
	// Unmarshal map deep storage
	s = tf["deep-storage"]
	w.m.ReadDeepStorage(s, w.ods)
	// Unmarshal objects on the map
	for i := 0; i < nThreads; i++ {
		s := tf[fmt.Sprintf("objects-%03d", i)]
		w.m.ReadObjects(s, w.ods)
	}
	// Call RecalculateStats on all objects in the map
	w.ods.RecalculateStats()
	// Map data
	w.m.Read(tf["map"])
	// Done
	end = time.Now()
	elapsed = end.Sub(start)
	log.Printf("info: save restored in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	return nil
}

// getFileName returns the file name portion of the save path without extension.
func (w *World) getFileName() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

// Marshal writes all of the data stores that the world is responsible for. A
// WaitGroup is returned to wait for the file to be written to disk.
func (w *World) Marshal() (*sync.WaitGroup, error) {
	nThreads := runtime.NumCPU()
	if !w.lock.TryLock() {
		return nil, ErrSaveFileLocked
	}
	defer w.lock.Unlock()
	w.BroadcastMessage(nil, "The world is saving, please wait...")
	fp := filepath.Join(w.savePath, w.getFileName()+".zip")
	fp = filepath.Clean(fp)
	os.MkdirAll(filepath.Dir(fp), 0777)
	log.Printf("info: saving data stores to %s", fp)
	start := time.Now()
	wg := &sync.WaitGroup{}
	tf := map[string]io.Reader{}
	// Global data
	s := bytes.NewBuffer(nil)
	tf["global"] = s
	wg.Add(1)
	go func(s io.Writer) {
		defer wg.Done()
		util.PutUInt32(s, uint32(nThreads))
		util.PutUInt64(s, uint64(w.time))
	}(s)
	// Timers
	s = bytes.NewBuffer(nil)
	tf["timers"] = s
	wg.Add(1)
	go func(s io.Writer) {
		// Raw data for timers, this shouldn't change anymore
		defer wg.Done()
		game.WriteTimers(s)
	}(s)
	// Accounting data
	s = bytes.NewBuffer(nil)
	tf["accounts"] = s
	wg.Add(1)
	go func(s io.Writer) {
		defer wg.Done()
		util.PutUInt32(s, uint32(len(w.accounts)))
		for _, a := range w.accounts {
			a.Write(s)
		}
	}(s)
	// Kick off the object persistance goroutines
	wg.Add(nThreads)
	for i := 0; i < nThreads; i++ {
		s = bytes.NewBuffer(nil)
		tf[fmt.Sprintf("objects-%03d", i)] = s
		go w.m.WriteObjects(wg, s, i, nThreads)
	}
	// Map data
	s = bytes.NewBuffer(nil)
	tf["map"] = s
	wg.Add(1)
	go w.m.Write(wg, s)
	s = bytes.NewBuffer(nil)
	tf["deep-storage"] = s
	wg.Add(1)
	go w.m.WriteDeepStorage(wg, s)
	// The main goroutine is blocked at this point
	wg.Wait()
	// Calculate and report elapsed time
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("info: generated save data in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	w.BroadcastMessage(nil, "World save completed in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	// Kick off another goroutine to persist the save to disk and let the main
	// goroutine continue.
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		err := util.ZipWrite(fp, tf)
		if err != nil {
			log.Printf("error: unable to write save file %s", fp)
		}
		end := time.Now()
		elapsed := end.Sub(start)
		log.Printf("info: saved file to disk in %ds%03dms", elapsed.Milliseconds()/1000, elapsed.Milliseconds()%1000)
	}()
	return wg, nil
}

// SendPacket sends a packet to the world's goroutine. Returns true on success.
// This never blocks.
func (w *World) SendPacket(p clientpacket.Packet, ns *NetState) bool {
	w.requestQueue <- worldPacket{
		Packet:   p,
		NetState: ns,
	}
	return true
}

// Add adds a new object to the world data stores. It is assigned a unique
// serial appropriate for its type. The object is returned. As a special case
// this function refuses to add a nil value to the game data store.
func (w *World) Add(obj any) {
	if obj == nil {
		return
	}
	switch o := obj.(type) {
	case *game.Mobile:
		w.ods.StoreMobile(o)
	case *game.Item:
		w.ods.StoreItem(o)
	}
}

// FindMobile returns the item with the given serial or nil.
func (w *World) FindMobile(id uo.Serial) (*game.Mobile, bool) {
	return w.ods.Mobile(id)
}

// FindItem returns the item with the given serial or nil.
func (w *World) FindItem(id uo.Serial) (*game.Item, bool) {
	return w.ods.Item(id)
}

// DeleteMobile deletes the mobile from the world.
func (w *World) DeleteMobile(m *game.Mobile) {
	w.ods.RemoveMobile(m)
}

// DeleteItem deletes the item from the world.
func (w *World) DeleteItem(i *game.Item) {
	w.ods.RemoveItem(i)
}

// AuthenticateAccount attempts to authenticate an account by username and
// password hash. If no account exists for that username, a new one will be
// created for the user. If an account is found but the password hashes do not
// match nil is returned. Otherwise the account is returned. If no accounts
// exist in the accounts datastore at all, the newly created account will have
// super-user permissions and a message will be logged.
func (w *World) AuthenticateAccount(username, passwordHash string) *game.Account {
	w.al.Lock()
	defer w.al.Unlock()
	// Find account or create new
	a := w.accounts[username]
	newAccount := false
	if a == nil {
		a = game.NewAccount(username, passwordHash, game.RolePlayer)
		newAccount = true
	}
	// Authenticate password
	if a.PasswordHash != passwordHash {
		return nil
	}
	// New account handling
	if newAccount {
		// Auto-create the super-user on first user login
		if len(w.accounts) == 0 {
			a = game.NewAccount(username, passwordHash, game.RoleAll)
			log.Printf("warning: new user %s granted all roles and marked as the super-user", username)
			w.superUser = a
		}
		// TODO DEBUG REMOVE Create the player mobile
		tn := "PlayerMobile"
		if a.HasRole(game.RoleDeveloper) {
			tn = "DeveloperMobile"
		} else if a.HasRole(game.RoleAdministrator) {
			tn = "AdministratorMobile"
		} else if a.HasRole(game.RoleGameMaster) {
			tn = "GameMasterMobile"
		}
		m := game.NewMobile(tn)
		m.Location = configuration.StartingLocation
		m.Name = a.Username
		world.m.StoreObject(m)
		a.Characters = append(a.Characters, m.Serial)
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
	if configuration.LogPanics {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("panic: %v\n%s\n", p, debug.Stack())
				panic(p)
			}
		}()
	}
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
			// Interleaved chunk updates, mobile think, etc
			w.m.Update(w.time)
			// OPLInfo updates
			for s := range w.oplUpdateList {
				if o, found := w.FindItem(s); found {
					if o.Removed {
						// Ignore items slated for removal
						continue
					}
					if o.Container != nil {
						// For items in containers we need the container to
						// distribute the OPL updates to all observers
						o.Container.UpdateItemOPL(o)
					} else if o.Wearer != nil {
						// For items worn by a mobile we need to update every
						// net state within range of that mobile
						_, info := o.OPLPackets()
						for _, m := range w.m.NetStatesInRange(o.Wearer.Location, 0) {
							m.NetState.Send(info)
						}
					} else {
						// For items on the ground we need to update every net
						// state in range of the item itself
						_, info := o.OPLPackets()
						for _, m := range w.m.NetStatesInRange(o.Location, 0) {
							m.NetState.Send(info)
						}
					}
				} else if o, found := world.FindMobile(s); found {
					if o.Removed {
						// Ignore items slated for removal
						continue
					}
					// Distribute the OPL info to every net state in range of
					// the mobile.
					_, info := o.OPLPackets()
					for _, m := range w.m.NetStatesInRange(o.Location, 0) {
						m.NetState.Send(info)
					}
				}
			}
			// Clear the OPL update list
			w.oplUpdateList = make(map[uo.Serial]struct{})
			// Update objects
			for s := range w.updateList {
				if i, found := w.FindItem(s); found {
					if i.Removed {
						// Ignore objects that are slated for removal
						continue
					}
					if i.Container != nil {
						// For items within a container the container has to
						// distribute the update to all container observers
						i.Container.UpdateItem(i)
					} else if i.Wearer != nil {
						// For items being worn by a mobile we need to
						// distribute the update to all net states in range of
						// the mobile wearing the item.
						for _, m := range w.m.NetStatesInRange(i.Wearer.Location, 0) {
							m.NetState.UpdateItem(i)
						}
					} else {
						// For items on the ground we need to distribute the
						// update to all net states in range of the item.
						for _, mob := range w.m.NetStatesInRange(i.Location, 0) {
							mob.NetState.UpdateItem(i)
						}
					}
				} else if m, found := w.FindMobile(s); found {
					if m.Removed {
						// Ignore objects that are slated for removal
						continue
					}
					// Distribute the update to all net states in range
					for _, mob := range w.m.NetStatesInRange(m.Location, 0) {
						mob.NetState.UpdateMobile(m)
					}
				}
			}
			// Clear the update list
			w.updateList = make(map[uo.Serial]struct{})
		case r := <-w.requestQueue:
			// Handle graceful shutdown
			if r.Packet == nil {
				ticker.Stop()
				done = true
				break
			}
			// If we are not trying to handle a tick we process packets
			fn, found := packetHandlers.Get(r.Packet.ID())
			if !found || fn == nil {
				log.Printf("error: unhandled packet 0x%02X", r.Packet.ID())
			} else {
				fn(r.NetState, r.Packet)
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
func (w *World) ItemDefinition(g uo.Graphic) *uo.StaticDefinition {
	return tileDataMul.GetStaticDefinition(int(g))
}

// UpdateItem schedules an update packet for the item.
func (w *World) UpdateItem(i *game.Item) {
	w.updateList[i.Serial] = struct{}{}
}

// UpdateMobile schedules an update packet for the item.
func (w *World) UpdateMobile(m *game.Mobile) {
	w.updateList[m.Serial] = struct{}{}
}

// BroadcastPacket broadcasts an arbitrary server packet to every connected
// client. Use this for things like server-wide system messages or global
// lighting and weather effects.
func (w *World) BroadcastPacket(p serverpacket.Packet) {
	BroadcastPacket(p)
}

// BroadcastMessage broadcasts lower-left system message to every connected
// client from the given speaker. Use nil for speaker for the system.
func (w *World) BroadcastMessage(s *game.Mobile, fs string, args ...any) {
	sid := uo.SerialSystem
	body := uo.BodySystem
	font := uo.FontNormal
	hue := uo.Hue(1153)
	name := ""
	text := fmt.Sprintf(fs, args...)
	sType := uo.SpeechTypeSystem
	if s != nil {
		sid = s.Serial
		sType = uo.SpeechTypeNormal
		name = s.DisplayName()
		body = s.Body
	}
	w.BroadcastPacket(&serverpacket.Speech{
		Speaker: sid,
		Body:    body,
		Font:    font,
		Hue:     hue,
		Name:    name,
		Text:    text,
		Type:    sType,
	})
}

// Insert inserts the object into the world's datastores blindly. *Only* used
// during a restore from backup.
func (w *World) Insert(obj any) {
	switch o := obj.(type) {
	case *game.Item:
		w.ods.InsertItem(o)
	case *game.Mobile:
		w.ods.InsertMobile(o)
	}
}

// UpdateItemOPLInfo adds the item to the list of items that must have their OPL
// data updated client-side.
func (w *World) UpdateItemOPLInfo(i *game.Item) {
	w.oplUpdateList[i.Serial] = struct{}{}
}

// UpdateMobileOPLInfo adds the mobile to the list of mobiles that must have
// their OPL data updated client-side.
func (w *World) UpdateMobileOPLInfo(m *game.Mobile) {
	w.oplUpdateList[m.Serial] = struct{}{}
}

// Accounts returns a slice of pointers to all accounts on the server for admin
// purposes.
func (w *World) Accounts() []*game.Account {
	ret := make([]*game.Account, 0, len(w.accounts))
	for _, a := range w.accounts {
		ret = append(ret, a)
	}
	return ret
}

// RemoveItem removes the item from the world datastores.
func (w *World) RemoveItem(i *game.Item) {
	w.ods.RemoveItem(i)
	i.Removed = true
}

// RemoveMobile removes the item from the world datastores.
func (w *World) RemoveMobile(m *game.Mobile) {
	w.ods.RemoveMobile(m)
	m.Removed = true
}
