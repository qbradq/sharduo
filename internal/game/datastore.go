package game

import "github.com/qbradq/sharduo/lib/uo"

// Datastore is responsible for managing pointers to all of the objects in the
// game.
type Datastore struct {
	spMobiles *uo.SerialPool        // Serial pool for mobiles
	spItems   *uo.SerialPool        // Serial pool for items
	mobiles   map[uo.Serial]*Mobile // Pointers to all mobiles
	items     map[uo.Serial]*Item   // Pointers to all items
}

// NewDatastore constructs a new, empty datastore for use.
func NewDatastore() *Datastore {
	return &Datastore{
		spMobiles: uo.NewSerialPool(uo.SerialFirstMobile, uo.SerialLastMobile, uo.SerialMobileNil),
		spItems:   uo.NewSerialPool(uo.SerialFirstItem, uo.SerialLastItem, uo.SerialItemNil),
		mobiles:   map[uo.Serial]*Mobile{},
		items:     map[uo.Serial]*Item{},
	}
}

// StoreMobile stores a mobile in the datastore assigning it a new serial.
func (s *Datastore) StoreMobile(m *Mobile) {
	m.Serial = s.spMobiles.Next()
	s.mobiles[m.Serial] = m
}

// InsertMobile stores a mobile in the datastore with its current serial, which
// must be unique.
func (s *Datastore) InsertMobile(m *Mobile) {
	s.mobiles[m.Serial] = m
	s.spMobiles.Insert(m.Serial)
}

// Mobile returns the mobile with the given serial or nil.
func (s *Datastore) Mobile(k uo.Serial) *Mobile {
	return s.mobiles[k]
}

// RemoveMobile removes a mobile from the datastore and releases its serial.
func (s *Datastore) RemoveMobile(m *Mobile) {
	delete(s.mobiles, m.Serial)
	s.spMobiles.Release(m.Serial)
}

// StoreItem stores an item in the datastore assigning it a new serial.
func (s *Datastore) StoreItem(i *Item) {
	i.Serial = s.spItems.Next()
	s.items[i.Serial] = i
}

// InsertItem stores an item in the datastore with its current serial, which
// must be unique.
func (s *Datastore) InsertItem(i *Item) {
	s.items[i.Serial] = i
	s.spItems.Insert(i.Serial)
}

// Item returns the item with the given serial or nil.
func (s *Datastore) Item(k uo.Serial) *Item {
	return s.items[k]
}

// RemoveItem removes an item from the datastore and releases its serial.
func (s *Datastore) RemoveItem(i *Item) {
	delete(s.items, i.Serial)
	s.spItems.Release(i.Serial)
}

// RecalculateStats calls RecalculateStats on all items and mobiles in the
// datastore.
func (s *Datastore) RecalculateStats() {
	for _, m := range s.mobiles {
		m.RecalculateStats()
	}
	for _, i := range s.items {
		i.RecalculateStats()
	}
}
