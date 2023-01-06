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
	// Gold returns the amount of gold within the mobile's backpack
	Gold() int
	// AdjustGold adds n to the total amount of gold on the mobile
	AdjustGold(int)

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
	// PickUp attempts to place an object on the mobile's cursor. Returns true
	// if successful.
	PickUp(Object) bool
	// DropItemInCursor drops the item in the cursor to the ground at the
	// mobile's feet.
	DropItemInCursor()
	// RequestCursorState is uses to set the cursor state to either
	// CursorStateDrop or CursorStateEquip. All other values will be ignored.
	RequestCursorState(CursorState)
	// Equip equips the given item in the item's layer, returns false if the
	// equip operation failed for any reason.
	Equip(Wearable) bool
	// ForceEquip forcefully equips the given item. If an existing item is in
	// that slot it will be leaked!
	ForceEquip(Wearable)
	// Unequip unequips the given item from the item's layer. It returns false
	// if the unequip operation failed for any reason.
	Unequip(Wearable) bool
	// ForceUnequip forcefully unequips the given item.
	ForceUnequip(Wearable)
	// EquipmentInSlot returns the item equipped in the named slot or nil.
	EquipmentInSlot(uo.Layer) Wearable
	// MapEquipment executes the function for every item this mobile has
	// equipped and returns any errors. Be careful, as this will also map over
	// inventory backpacks and player bank boxes. Filter them by checking the
	// wearable's layer.
	MapEquipment(func(Wearable) error) []error
	// DropToBackpack is a helper function that places items within a mobile's
	// backpack. If the second argument is true, the item will be placed without
	// regard to weight and item caps. Returns true if successful.
	DropToBackpack(Object, bool) bool
	// AdjustWeight adds n to this mobile's equipment collection's weight cache.
	AdjustWeight(float32)
	// InBank returns true if the given item is within this mobile's bank box.
	InBank(Object) bool
	// InBackpack returns true if the given item is within this mobile's
	// backpack.
	InBackpack(Object) bool
	// BankBoxOpen returns true if the mobile is currently observing its bank
	// box.
	BankBoxOpen() bool

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

	//
	// Cache values
	//

	// Total amount of gold in backpack, excludes bank box and cursor
	gold int
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
				m.cursor.item = item
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

// Gold implements the Mobile interface.
func (m *BaseMobile) Gold() int { return m.gold }

// AdjustGold implements the Mobile interface.
func (m *BaseMobile) AdjustGold(n int) { m.gold += n }

// ItemInCursor implements the Mobile interface.
func (m *BaseMobile) ItemInCursor() Item { return m.cursor.Item() }

// IsItemInCursor implements the Mobile interface.
func (m *BaseMobile) IsItemOnCursor() bool { return m.cursor.Occupied() }

// RequestCursorState implements the Mobile interface.
func (m *BaseMobile) RequestCursorState(s CursorState) {
	if s == CursorStateDrop || s == CursorStateEquip {
		m.cursor.State = s
	}
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
	world.Update(m)
}

// RecalculateStats implements the Object interface.
func (m *BaseMobile) RecalculateStats() {
	m.equipment.recalculateStats()
	// Update gold amount
	backpackObj := m.equipment.GetItemInLayer(uo.LayerBackpack)
	if backpackObj == nil {
		log.Printf("error: mobile %s has no backpack", m.Serial().String())
		return
	}
	backpack, ok := backpackObj.(Container)
	if !ok {
		log.Printf("error: mobile %s backpack was not a container", m.Serial().String())
		return
	}

	m.gold = 0
	var fn func(Container)
	fn = func(c Container) {
		c.MapContents(func(item Item) error {
			if container, ok := item.(Container); ok {
				fn(container)
			} else if item.TemplateName() == "GoldCoin" {
				m.gold += item.Amount()
			}
			return nil
		})
	}
	fn(backpack)
}

// PickUp attempts to pick up the object. Returns true if successful.
func (m *BaseMobile) PickUp(o Object) bool {
	if o == nil {
		if m.cursor.item == nil {
			return true
		}
		if m.cursor.PickUp(o) {
			world.Update(m)
			return true
		}
		return false
	}
	if m.cursor.item == nil || m.cursor.item.Serial() != o.Serial() {
		if !m.cursor.PickUp(o) {
			return false
		}
		if !world.Map().SetNewParent(o, m) {
			m.cursor.PickUp(nil)
			return false
		}
		world.Update(m)
		return true
	}
	return false
}

// doAddObject adds the object to us - forcefully if requested.
func (m *BaseMobile) doAddObject(obj Object, force bool) bool {
	if obj == nil {
		return force
	}
	// Handle items coming in from other sources
	if m.cursor.item == nil || m.cursor.item.Serial() != obj.Serial() {
		// Try to equip the item
		if wearable, ok := obj.(Wearable); ok {
			if m.doEquip(wearable, force) {
				return true
			}
		}
		// Try to force the object into the backpack
		return m.DropToBackpack(obj, force)
	}
	// Handle item on cursor
	if m.cursor.State == CursorStatePickUp {
		// This is the object we are currently picking up, accept it
		obj.SetParent(m)
		return true
	}
	if m.cursor.State == CursorStateReturn {
		// We are trying to get this object back to its original parent
		m.cursor.Return()
	}
	if m.cursor.State == CursorStateEquip {
		w, ok := obj.(Wearable)
		if !ok {
			return false
		}
		return m.Equip(w)
	}
	if m.cursor.State == CursorStateDrop {
		// This is the item we are trying to drop that got sent back to our
		// cursor.
		m.cursor.Return()
		return true
	}
	// Should never get here
	log.Println("SHOULD NOT GET HERE")
	return false
}

// ForceAddObject implements the Object interface.
func (m *BaseMobile) ForceAddObject(obj Object) {
	m.doAddObject(obj, true)
}

// AddObject adds the object to the mobile. It returns true if successful.
func (m *BaseMobile) AddObject(o Object) bool {
	return m.doAddObject(o, false)
}

// doRemove removes the object from the mobile, forcefully if requested.
func (m *BaseMobile) doRemove(o Object, force bool) bool {
	if o == nil {
		return true
	}
	item, ok := o.(Item)
	if !ok {
		// We don't own non-item objects
		return force
	}
	if wearable, ok := o.(Wearable); ok && m.equipment.Contains(wearable) {
		return m.doUnequip(wearable, force)
	}
	// If we are removing the cursor item we return true, otherwise we do not
	// own the object and return false.
	if m.cursor.item != nil && m.cursor.item.Serial() == item.Serial() {
		return true
	}
	return force
}

// RemoveObject removes the object from the mobile. It returns true if
// successful.
func (m *BaseMobile) RemoveObject(o Object) bool {
	return m.doRemove(o, false)
}

// ForceRemoveObject removes the object from the mobile forcefully.
func (m *BaseMobile) ForceRemoveObject(o Object) {
	m.doRemove(o, true)
}

// DropObject implements the Object interface
func (m *BaseMobile) DropObject(obj Object, l uo.Location, from Mobile) bool {
	// TODO Access calculations
	if from.Serial() != m.Serial() {
		return false
	}
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
		if force {
			log.Printf("error: Mobile.DropToBackpack(force) leaked object %s because it was not an item", o.Serial().String())
			world.Remove(o)
		}
		return force
	}
	backpackObj := m.equipment.GetItemInLayer(uo.LayerBackpack)
	if backpackObj == nil {
		// Something is very wrong
		if force {
			log.Printf("error: Mobile.DropToBackpack(force) leaked object %s because the backpack was not found", o.Serial().String())
			world.Remove(o)
		}
		return force
	}
	backpack, ok := backpackObj.(Container)
	if !ok {
		// Something is very wrong
		if force {
			log.Printf("error: Mobile.DropToBackpack(force) leaked object %s because the backpack was not a container", o.Serial().String())
			world.Remove(o)
		}
		return force
	}
	if !force {
		return backpack.AddObject(item)
	}
	backpack.ForceAddObject(o)
	return true
}

// doEquip equips a wearable to the mobile forcefully if requested
func (m *BaseMobile) doEquip(w Wearable, force bool) bool {
	if w == nil {
		return force
	}
	if m.equipment == nil {
		m.equipment = NewEquipmentCollection()
	}
	if !m.equipment.Equip(w) {
		if force {
			log.Printf("error: leaked object %s during force-equip", w.Serial().String())
			w.SetParent(TheVoid)
			world.Remove(w)
		}
	}
	w.SetParent(m)
	// Send the WearItem packet to all netstates in range, including our own
	for _, mob := range world.Map().GetNetStatesInRange(m.Location(), uo.MaxViewRange) {
		if mob.Location().XYDistance(m.Location()) <= mob.ViewRange() {
			mob.NetState().WornItem(w, m)
		}
	}
	world.Update(m)
	return true
}

// Equip implements the Mobile interface.
func (m *BaseMobile) Equip(w Wearable) bool {
	return m.doEquip(w, false)
}

// ForceEquip implements the Mobile interface.
func (m *BaseMobile) ForceEquip(w Wearable) {
	m.doEquip(w, false)
}

// doUnequip unequips the wearable forcefully if requested
func (m *BaseMobile) doUnequip(w Wearable, force bool) bool {
	if m.equipment == nil || w == nil {
		return force
	}
	worn := m.equipment.GetItemInLayer(w.Layer())
	if worn == nil || worn.Serial() != w.Serial() {
		return force
	}
	if !m.equipment.Unequip(w) {
		return force
	}
	// Send the remove item packet to everyone including ourselves
	for _, mob := range world.Map().GetNetStatesInRange(m.Location(), uo.MaxViewRange) {
		mob.NetState().RemoveObject(w)
	}
	world.Update(m)
	return true
}

// Unequip implements the Mobile interface.
func (m *BaseMobile) Unequip(w Wearable) bool {
	return m.doUnequip(w, false)
}

// ForceUnequip implements the Mobile interface.
func (m *BaseMobile) ForceUnequip(w Wearable) {
	m.doUnequip(w, true)
}

// EquipmentInSlot implements the Mobile interface.
func (m *BaseMobile) EquipmentInSlot(l uo.Layer) Wearable {
	if m.equipment == nil {
		return nil
	}
	return m.equipment.GetItemInLayer(l)
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

// AdjustWeight implements the Object interface
func (m *BaseMobile) AdjustWeight(n float32) {
	if m.equipment != nil {
		m.equipment.weight += n
	}
	world.Update(m)
}

// Weight implements the Object interface
func (m *BaseMobile) Weight() float32 {
	ret := m.equipment.Weight()
	if m.cursor.item != nil {
		ret += m.cursor.item.Weight()
	}
	return ret
}

// AfterMove implements the Mobile interface.
func (m *BaseMobile) AfterMove() {
	// Check for containers that we need to close
	if m.NetState() != nil {
		m.NetState().ContainerRangeCheck()
	}
}

// InBank implements the Mobile interface.
func (m *BaseMobile) InBank(o Object) bool {
	if o == nil {
		return false
	}
	if !m.isPlayerCharacter {
		// Non-player-characters do not have bank boxes
		return false
	}
	root := RootParent(o)
	if root == nil || root.Serial() != m.Serial() {
		// Object is a child of the map or another mobile
		return false
	}
	bbobj := m.EquipmentInSlot(uo.LayerBankBox)
	if bbobj == nil {
		// Something is very wrong
		log.Printf("error: player mobile %s does not have a bank box", m.Serial().String())
		return false
	}
	// Inspect the parent chain to see if the bank box is anywhere in the chain
	thiso := o
	thisp := thiso.Parent()
	for {
		if thisp == nil {
			// Hit the top-level object without a match
			return false
		}
		if thisp.Serial() == bbobj.Serial() {
			// Hit our own bank box
			return true
		}
		thiso = thisp
		thisp = thiso.Parent()
	}
}

// InBackpack implements the Mobile interface.
func (m *BaseMobile) InBackpack(o Object) bool {
	root := RootParent(o)
	if root == nil || root.Serial() != m.Serial() {
		// Object is a child of the map or another mobile
		return false
	}
	bpobj := m.EquipmentInSlot(uo.LayerBackpack)
	if bpobj == nil {
		// Something is very wrong
		log.Printf("error: mobile %s does not have a backpack", m.Serial().String())
		return false
	}
	// Inspect the parent chain to see if the backpack is anywhere in the chain
	thiso := o
	thisp := thiso.Parent()
	for {
		if thisp == nil {
			// Hit the top-level object without a match
			return false
		} else if thisp.Serial() == bpobj.Serial() {
			// Hit our own bank box
			return true
		}
		thiso = thisp
		thisp = thiso.Parent()
	}
}

// BankBoxOpen implements the Mobile interface.
func (m *BaseMobile) BankBoxOpen() bool {
	if m.NetState() == nil || !m.isPlayerCharacter {
		// Non-player-characters do not have bank boxes
		// If there is no attached net state there can't be any observed
		// containers
		return false
	}
	bbobj := m.EquipmentInSlot(uo.LayerBankBox)
	if bbobj == nil {
		// Something is very wrong
		log.Printf("error: player mobile %s does not have a bank box", m.Serial().String())
		return false
	}
	return m.NetState().ContainerIsObserving(bbobj)
}
