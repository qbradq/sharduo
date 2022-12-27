package game

import (
	"log"

	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseMobile{} })
}

// Mobile is the interface all mobiles implement
type Mobile interface {
	Object
	ContainerObserver

	//
	// NetState
	//

	// NetState returns the NetState implementation currently bound to this
	// mobile, or nil if there is none.
	NetState() NetState
	// SetNetState sets the currently bound NetState. Use SetNetState(nil) to
	// disconnect the mobile.
	SetNetState(NetState)

	//
	// Stats, attributes, and skills
	//

	// MobileFlags returns the MobileFlags value for this mobile
	MobileFlags() uo.MobileFlags
	// Strength returns the current effective strength
	Strength() int
	// Dexterity returns the current effective dexterity
	Dexterity() int
	// Intelligence returns the current effective intelligence
	Intelligence() int
	// HitPoints returns the current hit points
	HitPoints() int
	// MaxHitPoints returns the current effective max hit points
	MaxHitPoints() int
	// Mana returns the current mana
	Mana() int
	// MaxMana returns the current effective max mana
	MaxMana() int
	// Stamina returns the current stamina
	Stamina() int
	// MaxStamina returns the current effective max stamina
	MaxStamina() int

	//
	// AI-related values
	//

	// ViewRange returns the number of tiles this mobile can see and visually
	// observe objects in the world. If this mobile has an attached NetState,
	// this value can change at any time at the request of the player.
	ViewRange() int
	// SetViewRange sets the view range of the mobile, bounding it to sane
	// values.
	SetViewRange(int)
	// IsRunning returns true if the mobile is running.
	IsRunning() bool
	// Facing returns the current facing of the mobile.
	Facing() uo.Direction
	// SetFacing sets the current facing of the mobile.
	SetFacing(uo.Direction)
	// SetRunning sets the running flag of the mobile.
	SetRunning(bool)

	//
	// Graphics and display
	//

	// GetBody returns the animation body of the mobile.
	Body() uo.Body
	// IsPlayerCharacter returns true if this mobile is attached to a player's
	// account.
	IsPlayerCharacter() bool
	// IsFemale returns true if the mobile is female.
	IsFemale() bool
	// IsHumanBody returns true if the body value is humanoid.
	IsHumanBody() bool

	//
	// Equipment and inventory
	//

	// ItemInHand returns a pointer to the item held in the mobile's cursor.
	// This is usually only used by mobiles being controlled by a client.
	ItemInCursor() Item
	// Returns true if the mobile's cursor has an item on it.
	IsItemOnCursor() bool
	// SetItemInCursor sets the item held in the mobile's cursor. Returns true
	// if successful.
	SetItemInCursor(Item) bool
	// DropItemInCursor drops the item in the cursor to the ground at the
	// mobile's feet.
	DropItemInCursor()
	// Equip equips the given item in the item's layer, returns false if the
	// equip operation failed for any reason.
	Equip(Wearable) bool
	// Unequip unequips the given item from the item's layer. It returns false
	// if the unequip operation failed for any reason.
	Unequip(Wearable) bool
	// MapEquipment executes the function for every item this mobile has
	// equipped and returns any errors. Be careful, as this will also map over
	// inventory backpacks and player bank boxes. Filter them by checking the
	// wearable's layer.
	MapEquipment(func(Wearable) error) []error
	// DropToBackpack is a helper function that places items within a mobile's
	// backpack. If the second argument is true, the item will be placed without
	// regard to weight and item caps. Returns true if successful.
	DropToBackpack(Object, bool) bool

	//
	// Notoriety system
	//

	// GetNotorietyFor returns the notoriety value of the given mobile as
	// observed from this mobile.
	GetNotorietyFor(Mobile) uo.Notoriety

	//
	// Callbacks
	//

	// AfterMove is called by Map.MoveMobile after the mobile has arrived at its
	// new location.
	AfterMove()
}

// BaseMobile provides the base implementation for Mobile
type BaseMobile struct {
	BaseObject

	//
	// Network
	//

	// Attached NetState implementation
	n NetState
	// Collection of all the containers we are currently observing
	observedContainers util.Slice[Container]

	//
	// AI values and such
	//

	// Current view range of the mobile. Please note that the zero value IS NOT
	// SANE for this variable!
	viewRange int
	// isPlayerCharacter is true if the mobile is attached to a player's account
	isPlayerCharacter bool
	// isFemale is true if the mobile is female
	isFemale bool
	// Animation body of the object
	body uo.Body
	// Running flag
	isRunning bool
	// Notoriety of the mobile
	notoriety uo.Notoriety

	//
	// User interface stuff
	//

	cursor *Cursor

	//
	// Equipment and inventory
	//

	// The collection of equipment this mobile is wearing, if any
	equipment *EquipmentCollection

	//
	// Stats and attributes
	//

	// Base strength
	baseStrength int
	// Base dexterity
	baseDexterity int
	// Base intelligence
	baseIntelligence int
	// Current HP
	hitPoints int
	// Current mana
	mana int
	// Current stamina
	stamina int
}

// GetTypeName implements the util.Serializeable interface.
func (m *BaseMobile) TypeName() string {
	return "BaseMobile"
}

// SerialType implements the util.Serializeable interface.
func (o *BaseMobile) SerialType() uo.SerialType {
	return uo.SerialTypeMobile
}

// Serialize implements the util.Serializeable interface.
func (m *BaseMobile) Serialize(f *util.TagFileWriter) {
	m.BaseObject.Serialize(f)
	f.WriteNumber("ViewRange", m.viewRange)
	f.WriteBool("IsPlayerCharacter", m.isPlayerCharacter)
	f.WriteBool("IsFemale", m.isFemale)
	f.WriteNumber("Body", int(m.body))
	f.WriteNumber("Notoriety", int(m.notoriety))
	f.WriteNumber("Strength", m.baseStrength)
	f.WriteNumber("Dexterity", m.baseDexterity)
	f.WriteNumber("Intelligence", m.baseIntelligence)
	f.WriteNumber("HitPoints", m.hitPoints)
	f.WriteNumber("Stamina", m.stamina)
	f.WriteNumber("Mana", m.mana)
	if m.equipment != nil {
		m.equipment.Write("Equipment", f)
	}
	if m.cursor.Occupied() {
		f.WriteHex("ItemInCursor", uint32(m.cursor.Item().Serial()))
	}
}

// Deserialize implements the util.Serializeable interface.
func (m *BaseMobile) Deserialize(f *util.TagFileObject) {
	m.cursor = &Cursor{}
	m.BaseObject.Deserialize(f)
	m.viewRange = f.GetNumber("ViewRange", uo.MaxViewRange)
	m.isPlayerCharacter = f.GetBool("IsPlayerCharacter", false)
	m.isFemale = f.GetBool("IsFemale", false)
	m.body = uo.Body(f.GetNumber("Body", int(uo.BodyDefault)))
	// Special case for human bodies to select between male and female models
	if m.body == uo.BodyHuman && m.isFemale {
		m.body += 1
	}
	m.notoriety = uo.Notoriety(f.GetNumber("Notoriety", int(uo.NotorietyAttackable)))
	m.baseStrength = f.GetNumber("Strength", 10)
	m.baseDexterity = f.GetNumber("Dexterity", 10)
	m.baseIntelligence = f.GetNumber("Intelligence", 10)
	m.hitPoints = f.GetNumber("HitPoints", 1)
	m.mana = f.GetNumber("Mana", 1)
	m.stamina = f.GetNumber("Stamina", 1)
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (m *BaseMobile) OnAfterDeserialize(f *util.TagFileObject) {
	m.equipment = NewEquipmentCollectionWith(f.GetObjectReferences("Equipment"))
	for _, w := range m.equipment.equipment {
		w.SetParent(m)
	}
	// Make sure all mobiles have an inventory
	if !m.equipment.IsLayerOccupied(uo.LayerBackpack) {
		var w Wearable
		if m.isPlayerCharacter {
			w = world.New("PlayerBackpack").(Wearable)
		} else {
			w = world.New("NPCBackpack").(Wearable)
		}
		if !m.Equip(w) {
			log.Println("failed to equip auto-generated inventory backpack")
		}
	}
	// Make sure all players have a bank box
	if m.IsPlayerCharacter() && !m.equipment.IsLayerOccupied(uo.LayerBankBox) {
		if !m.Equip(world.New("PlayerBankBox").(Wearable)) {
			log.Println("failed to equip auto-generated player bank box")
		}
	}
	// If we had an item on the cursor at the time of the save we drop it at
	// our feet just so we don't leak it.
	incs := uo.Serial(f.GetHex("ItemInCursor", uint32(uo.SerialItemNil)))
	if incs != uo.SerialItemNil {
		o := world.Find(incs)
		if o != nil {
			if item, ok := o.(Item); ok {
				m.cursor.PickUp(item)
				m.DropItemInCursor()
			}
		}
	}
}

// NetState implements the Mobile interface.
func (m *BaseMobile) NetState() NetState { return m.n }

// SetNetState implements the Mobile interface.
func (m *BaseMobile) SetNetState(n NetState) {
	m.n = n
}

// ViewRange implements the Mobile interface.
func (m *BaseMobile) ViewRange() int { return m.viewRange }

// SetViewRange implements the Mobile interface.
func (m *BaseMobile) SetViewRange(r int) { m.viewRange = uo.BoundViewRange(r) }

// Body implements the Mobile interface.
func (m *BaseMobile) Body() uo.Body { return m.body }

// IsPlayerCharacter implements the Mobile interface.
func (m *BaseMobile) IsPlayerCharacter() bool { return m.isPlayerCharacter }

// IsFemale implements the Mobile interface.
func (m *BaseMobile) IsFemale() bool { return m.isFemale }

// IsHumanBody implements the Mobile interface.
func (m *BaseMobile) IsHumanBody() bool {
	return m.body == uo.BodyHumanMale || m.body == uo.BodyHumanFemale
}

// IsRunning implements the Mobile interface.
func (m *BaseMobile) IsRunning() bool { return m.isRunning }

// SetRunning implements the Mobile interface.
func (m *BaseMobile) SetRunning(v bool) {
	m.isRunning = v
}

// Facing implements the Mobile interface.
func (m *BaseMobile) Facing() uo.Direction { return m.facing }

// SetFacing implements the Mobile interface.
func (m *BaseMobile) SetFacing(f uo.Direction) {
	m.facing = f.StripRunningFlag()
}

// MobileFlags implements the Mobile interface.
func (m *BaseMobile) MobileFlags() uo.MobileFlags {
	ret := uo.MobileFlagNone
	if m.IsFemale() {
		ret |= uo.MobileFlagFemale
	}
	return ret
}

// Strength implements the Mobile interface.
func (m *BaseMobile) Strength() int { return m.baseStrength }

// Dexterity implements the Mobile interface.
func (m *BaseMobile) Dexterity() int { return m.baseDexterity }

// Intelligence implements the Mobile interface.
func (m *BaseMobile) Intelligence() int { return m.baseIntelligence }

// HitPoints implements the Mobile interface.
func (m *BaseMobile) HitPoints() int { return m.hitPoints }

// MaxHitPoints implements the Mobile interface.
func (m *BaseMobile) MaxHitPoints() int { return 50 + m.Strength()/2 }

// Mana implements the Mobile interface.
func (m *BaseMobile) Mana() int { return m.mana }

// MaxMana implements the Mobile interface.
func (m *BaseMobile) MaxMana() int { return m.Intelligence() }

// Stamina implements the Mobile interface.
func (m *BaseMobile) Stamina() int { return m.stamina }

// MaxStamina implements the Mobile interface.
func (m *BaseMobile) MaxStamina() int { return m.Dexterity() }

func (m *BaseMobile) ItemInCursor() Item { return m.cursor.Item() }

// Returns true if the mobile's cursor has an item on it.
func (m *BaseMobile) IsItemOnCursor() bool { return m.cursor.Occupied() }

// SetItemInCursor sets the item held in the mobile's cursor. It returns true
// if successful.
func (m *BaseMobile) SetItemInCursor(item Item) bool {
	return m.cursor.PickUp(item)
	if !world.Map().SetNewParent(item, m) {
		m.cursor.PickUp(nil)
		return false
	}
	return true
}

// DropItemInCursor drops the item in the player's cursor to their feet.
func (m *BaseMobile) DropItemInCursor() {
	item := m.cursor.Item()
	if item == nil {
		return
	}
	m.cursor.PickUp(nil)
	item.SetLocation(m.location)
	item.SetParent(nil)
	world.Map().AddObject(item)
}

// RecalculateStats implements the Object interface.
func (m *BaseMobile) RecalculateStats() {
	m.equipment.recalculateStats()
}

// AddObject adds the object to the mobile. It returns true if successful.
func (m *BaseMobile) AddObject(o Object) bool {
	if o == nil {
		return true
	}
	if m.cursor.State == CursorStatePickUp {
		m.cursor.PickUp(o)
		return true
	}
	if m.toDrop == o && m.itemInCursor == o {
		// This is the item we were trying to drop. This means that whatever
		// we were trying to drop it into refused, so now we need to try to
		// send it back to the previous parent.
		oldParent := o.PreviousParent()
		o.SetParent(oldParent)
		m.toDrop = nil
		if w, ok := o.(Wearable); ok && oldParent == m {
			// Something got sent back to the cursor that was equipped to us.
			m.toWear = w
			// Fall-through to the equipment handling section. Kinda hacky.
		} else if oldParent == nil {
			// Needs to be returned to the map
			world.Map().ForceAddObject(o)
			return true
		} else {
			if container, ok := oldParent.(Container); ok {
				// Needs to be returned to a container
				container.ForceAddObject(o)
				return true
			}
			// We should never reach this
			return false
		}
	}
	if m.toWear == o {
		// This is the item we are trying to wear
		m.toWear = nil
		w, ok := o.(Wearable)
		if !ok {
			return false
		}
		if m.equipment == nil {
			m.equipment = NewEquipmentCollection()
		}
		if !m.equipment.Equip(w) {
			return false
		}
		// Send the WearItem packet to all netstates in range, including our own
		for _, mob := range world.Map().GetNetStatesInRange(m.Location(), uo.MaxViewRange) {
			if mob.Location().XYDistance(m.Location()) <= mob.ViewRange() {
				mob.NetState().WornItem(w, m)
			}
		}
		return true
	}
	if m.itemInCursor == o {
		// This is the item we are trying to put on the cursor
		m.toDrop = o
		return true
	}
	// Don't know what to do with object
	return false
}

// RemoveObject removes the object from the mobile. It returns true if
// successful.
func (m *BaseMobile) RemoveObject(o Object) bool {
	if o == nil {
		return true
	}
	if wearable, ok := o.(Wearable); ok && m.equipment.Contains(wearable) {
		// This item is currently equipped, try to unequip it
		if !m.equipment.Unequip(wearable) {
			m.SetItemInCursor(nil)
			// Send the wear item packet back at ourselves to force the item
			// back into the paper doll
			m.NetState().WornItem(wearable, m)
			return false
		}
		return true
	}
	if m.toDrop == o || m.toWear == o || m.itemInCursor == o {
		// This is the item we are dropping or trying to manipulate
		return true
	}
	// We don't own this object, reject the remove request
	return false
}

// DropObject implements the Object interface
func (m *BaseMobile) DropObject(obj Object, l uo.Location, from Mobile) bool {
	// TODO Access calculations
	// Try to put the object in our backpack
	backpack := m.equipment.GetItemInLayer(uo.LayerBackpack)
	if backpack == nil {
		// No backpack found, something is very wrong
		return false
	}
	if item, ok := obj.(Item); ok {
		item.SetDropLocation(l)
		return backpack.DropObject(item, l, from)
	}
	return false
}

// DropToBackpack implements the Mobile interface.
func (m *BaseMobile) DropToBackpack(o Object, force bool) bool {
	item, ok := o.(Item)
	if !ok {
		// Something is very wrong
		return false
	}
	backpack, ok := m.equipment.GetItemInLayer(uo.LayerBackpack).(Container)
	if !ok {
		// Something is very wrong
		return false
	}
	if !force {
		return backpack.AddObject(item)
	}
	backpack.ForceAddObject(o)
	return true
}

// Equip implements the Mobile interface.
func (m *BaseMobile) Equip(w Wearable) bool {
	if w == nil {
		return false
	}
	m.toWear = w
	return m.AddObject(w)
}

// Unequip implements the Mobile interface.
func (m *BaseMobile) Unequip(w Wearable) bool {
	if m.equipment == nil {
		return false
	}
	if !m.equipment.Unequip(w) {
		return false
	}
	// Send the remove item packet to everyone including ourselves
	for _, mob := range world.Map().GetNetStatesInRange(m.Location(), uo.MaxViewRange) {
		mob.NetState().RemoveObject(w)
	}
	return true
}

// MapEquipment implements the Mobile interface.
func (m *BaseMobile) MapEquipment(fn func(Wearable) error) []error {
	var ret []error
	for _, w := range m.equipment.equipment {
		if err := fn(w); err != nil {
			ret = append(ret, err)
		}
	}
	return ret
}

// GetNotorietyFor implements the Mobile interface.
func (m *BaseMobile) GetNotorietyFor(other Mobile) uo.Notoriety {
	// TODO Guild system
	// TODO If this is a player's mobile return innocent
	return m.notoriety
}

// Weight implements the Object interface
func (m *BaseMobile) Weight() int {
	ret := m.equipment.Weight()
	if w := m.equipment.GetItemInLayer(uo.LayerBackpack); w != nil {
		if backpack, ok := w.(Container); ok {
			ret += backpack.ContentWeight()
		}
	}
	if m.itemInCursor != nil {
		ret += m.itemInCursor.Weight()
		if container, ok := m.itemInCursor.(Container); ok {
			ret += container.ContentWeight()
		}
	}
	return ret
}

// ContainerOpen implements the ContainerObserver interface.
func (m *BaseMobile) ContainerOpen(c Container) {
	if m.n != nil {
		m.n.OpenContainer(c)
		if !m.observedContainers.Contains(c) {
			m.observedContainers = m.observedContainers.Append(c)
		}
	}
}

// ContainerClose implements the ContainerObserver interface.
func (m *BaseMobile) ContainerClose(c Container) {
	if m.NetState() != nil {
		m.NetState().CloseGump(c.Serial())
	}
	m.observedContainers = m.observedContainers.Remove(c)
}

// ContainerOpen implements the ContainerObserver interface.
func (m *BaseMobile) ContainerItemAdded(c Container, item Item) {
	if m.n != nil {
		m.n.AddItemToContainer(c, item)
	}
}

// ContainerOpen implements the ContainerObserver interface.
func (m *BaseMobile) ContainerItemRemoved(c Container, item Item) {
	if m.n != nil {
		m.n.RemoveItemFromContainer(c, item)
	}
}

// AfterMove implements the Mobile interface.
func (m *BaseMobile) AfterMove() {
	// Check for containers that we need to close
	for _, container := range m.observedContainers.Copy() {
		if m.location.XYDistance(container.Location()) > uo.MaxContainerViewRange {
			container.RemoveObserver(m)
			m.n.CloseGump(container.Serial())
		}
	}
}
