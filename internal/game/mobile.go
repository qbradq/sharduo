package game

import (
	"log"

	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseMobile{} })
	objectCtors[marshal.ObjectTypeMobile] = func() Object { return &BaseMobile{} }
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
	// Skill returns the raw skill value (range 0-1000) of the named skill
	Skill(uo.Skill) int16
	// Skills returns a slice of all raw skill values (range 0-1000)
	Skills() []int16
	// SkillCheck returns true if the skill check succeeded. This function will
	// also calculate skill and stat gains and send mobile updates for these.
	SkillCheck(uo.Skill, int, int) bool

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
	// DropToFeet is a helper function that places items at the mobile's feet.
	// The item is forced to the location of the mobile's feet without regard to
	// having enough space. Use this function as a last-ditch method to get an
	// item to a player.
	DropToFeet(Object)
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

	//
	// Queries
	//

	// CanAccess returns true if this mobile has access to the given object.
	// This only considers container accessability and ownership, NOT line of
	// sight or range.
	CanAccess(Object) bool
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
	// Raw skill values
	skills []int16

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

// ObjectType implements the Object interface.
func (m *BaseMobile) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeMobile
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
	for i, v := range m.skills {
		if v != 0 {
			f.WriteNumber("Skill"+uo.SkillNames[i], int(v))
		}
	}
}

// Marshal implements the marshal.Marshaler interface.
func (m *BaseMobile) Marshal(s *marshal.TagFileSegment) {
	cs := uo.SerialZero
	if m.cursor.item != nil {
		cs = m.cursor.item.Serial()
	}
	m.BaseObject.Marshal(s)
	s.PutTag(marshal.TagViewRange, marshal.TagValueByte, byte(m.viewRange))
	s.PutTag(marshal.TagIsPlayerCharacter, marshal.TagValueBool, m.isPlayerCharacter)
	s.PutTag(marshal.TagIsFemale, marshal.TagValueBool, m.isFemale)
	s.PutTag(marshal.TagBody, marshal.TagValueShort, uint16(m.body))
	s.PutTag(marshal.TagNotoriety, marshal.TagValueByte, byte(m.notoriety))
	s.PutTag(marshal.TagStrength, marshal.TagValueShort, uint16(m.baseStrength))
	s.PutTag(marshal.TagDexterity, marshal.TagValueShort, uint16(m.baseDexterity))
	s.PutTag(marshal.TagIntelligence, marshal.TagValueShort, uint16(m.baseIntelligence))
	s.PutTag(marshal.TagHitPoints, marshal.TagValueShort, uint16(m.hitPoints))
	s.PutTag(marshal.TagStamina, marshal.TagValueShort, uint16(m.stamina))
	s.PutTag(marshal.TagMana, marshal.TagValueShort, uint16(m.mana))
	s.PutTag(marshal.TagCursor, marshal.TagValueInt, uint32(cs))
	s.PutTag(marshal.TagSkills, marshal.TagValueShortSlice, m.skills)
}

// Deserialize implements the util.Serializeable interface.
func (m *BaseMobile) Deserialize(f *util.TagFileObject) {
	m.skills = make([]int16, uo.SkillCount)
	m.cursor = &Cursor{}
	m.BaseObject.Deserialize(f)
	m.viewRange = f.GetNumber("ViewRange", uo.MaxViewRange)
	m.isPlayerCharacter = f.GetBool("IsPlayerCharacter", false)
	m.isFemale = f.GetBool("IsFemale", false)
	m.body = uo.Body(f.GetNumber("Body", int(uo.BodyDefault)))
	m.notoriety = uo.Notoriety(f.GetNumber("Notoriety", int(uo.NotorietyAttackable)))
	m.baseStrength = f.GetNumber("Strength", 10)
	m.baseDexterity = f.GetNumber("Dexterity", 10)
	m.baseIntelligence = f.GetNumber("Intelligence", 10)
	m.hitPoints = f.GetNumber("HitPoints", 1)
	m.mana = f.GetNumber("Mana", 1)
	m.stamina = f.GetNumber("Stamina", 1)
	// Load skills
	for s := uo.SkillFirst; s <= uo.SkillLast; s++ {
		m.skills[s] = int16(f.GetNumber("Skill"+uo.SkillNames[s], 0))
	}
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (m *BaseMobile) Unmarshal(to *marshal.TagObject) {
	m.BaseObject.Unmarshal(to)
	m.cursor = &Cursor{}
	m.viewRange = uo.BoundViewRange(int(to.Tags.Byte(marshal.TagViewRange)))
	m.isPlayerCharacter = to.Tags.Bool(marshal.TagIsPlayerCharacter)
	m.isFemale = to.Tags.Bool(marshal.TagIsFemale)
	m.body = uo.Body(to.Tags.Short(marshal.TagBody))
	m.notoriety = uo.Notoriety(to.Tags.Byte(marshal.TagNotoriety))
	m.baseStrength = int(to.Tags.Short(marshal.TagStrength))
	m.baseDexterity = int(to.Tags.Short(marshal.TagDexterity))
	m.baseIntelligence = int(to.Tags.Short(marshal.TagIntelligence))
	m.hitPoints = int(to.Tags.Short(marshal.TagStrength))
	m.stamina = int(to.Tags.Short(marshal.TagStrength))
	m.mana = int(to.Tags.Short(marshal.TagStrength))
	m.skills = to.Tags.ShortSlice(marshal.TagSkills)
	if len(m.skills) < int(uo.SkillCount) {
		s := make([]int16, uo.SkillCount)
		copy(s, m.skills)
	} else if len(m.skills) > int(uo.SkillCount) {
		m.skills = m.skills[0:uo.SkillCount]
	}
}

// OnAfterDeserialize implements the util.Serializeable interface.
func (m *BaseMobile) OnAfterDeserialize(f *util.TagFileObject) {
	m.BaseObject.OnAfterDeserialize(f)
	m.equipment = NewEquipmentCollectionWith(f.GetObjectReferences("Equipment"), m)
	// If we had an item on the cursor at the time of the save we drop it at
	// our feet just so we don't leak it.
	incs := uo.Serial(f.GetHex("ItemInCursor", uint32(uo.SerialItemNil)))
	if incs != uo.SerialItemNil {
		o := world.Find(incs)
		if o != nil {
			m.DropToFeet(o)
		}
	}
	if !m.equipment.IsLayerOccupied(uo.LayerBackpack) {
		log.Printf("error: mobile %s does not have a backpack, removing", m.Serial().String())
		world.Remove(m)
		return
	}
	// Make sure all players have a bank box
	if m.IsPlayerCharacter() && !m.equipment.IsLayerOccupied(uo.LayerBankBox) {
		log.Printf("error: player mobile %s does not have a bank box, removing", m.Serial().String())
		world.Remove(m)
		return
	}
}

// AfterUnmarshal implements the marshal.Unmarshaler interface.
func (m *BaseMobile) AfterUnmarshal(to *marshal.TagObject) {
	m.BaseObject.AfterUnmarshal(to)
	m.equipment = NewEquipmentCollectionWith(to.Tags.ReferenceSlice(marshal.TagEquipment), m)
	// If we had an item on the cursor at the time of the save we drop it at
	// our feet just so we don't leak it.
	incs := uo.Serial(to.Tags.Int(marshal.TagCursor))
	if incs != 0 {
		o := world.Find(incs)
		if o != nil {
			m.DropToFeet(o)
		}
	}
	// Make sure all mobiles have a backpack, this should be covered in the template
	if !m.equipment.IsLayerOccupied(uo.LayerBackpack) {
		log.Printf("error: mobile %s does not have a backpack, removing", m.Serial().String())
		world.Remove(m)
		return
	}
	// Make sure all players have a bank box
	if m.IsPlayerCharacter() && !m.equipment.IsLayerOccupied(uo.LayerBankBox) {
		log.Printf("error: player mobile %s does not have a bank box, removing", m.Serial().String())
		world.Remove(m)
		return
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
	bpo := m.EquipmentInSlot(uo.LayerBackpack)
	if bpo != nil && bpo.TemplateName() == "PackAnimalBackpack" {
		// We are a pack animal, try to put the item in our pack
		// TODO check master
		return m.DropToBackpack(obj, false)
	}
	if from.Serial() == m.Serial() {
		// We are dropping something onto ourselves, try to put it in our
		// backpack
		return m.DropToBackpack(obj, false)
	}
	// TODO try to feed tamed animals
	// This is a regular mobile, dropping things on them makes no sense
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

// DropToFeet implements the Mobile interface.
func (m *BaseMobile) DropToFeet(o Object) {
	o.SetLocation(m.location)
	world.Map().ForceAddObject(o)
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

// CanAccess implements the Mobile interface.
func (m *BaseMobile) CanAccess(o Object) bool {
	if o == nil {
		return false
	}
	for {
		if o.Parent() == nil {
			// Object is directly on the map
			return true
		}
		if c, ok := o.(Container); ok {
			if m.n != nil && m.n.ContainerIsObserving(c) {
				// We are observing the container or one of the object's parent
				// containers, so we can see the object. No need for redundant
				// access rights checking.
				return true
			}
			if c.TemplateName() == "PlayerBankBox" {
				// This is a player's bank box and it's not open, otherwise the
				// above check would have been true. Bail now. Even though we
				// own the bank box we can't access its children if it's closed.
				return false
			}
		}
		if owner, ok := o.(Mobile); ok {
			// This object is directly owned by a mobile
			return owner.Serial() == m.Serial()
		}
		o = o.Parent()
	}
}

// Skill implements the Mobile interface.
func (m *BaseMobile) Skill(which uo.Skill) int16 { return m.skills[which] }

// Skills implements the Mobile interface.
func (m *BaseMobile) Skills() []int16 { return m.skills }

// SkillCheck implements the Mobile interface.
func (m *BaseMobile) SkillCheck(which uo.Skill, min, max int) bool {
	if which > uo.SkillLast {
		return false
	}
	// Get the skill value and look for corner cases
	v := int(m.skills[which])
	if v < min {
		// No chance
		return false
	}
	if v >= max {
		// No callange
		return true
	}
	// Calculate success
	spread := max - min
	chance := ((v - min) * 1000) / (max - min)
	success := false
	if world.Random().Random(0, spread) < chance {
		success = true
	}
	// Calculate skill gain
	tryGainSkill := v < 1000
	// TODO Check skill lock
	if tryGainSkill {
		gc := 1000
		gc += 1000 - v
		gc /= 2
		if success {
			gc *= 1500
			gc /= 1000
		}
		gc /= 2
		if gc < 10 {
			gc = 10
		}
		if gc > 1000 {
			gc = 1000
		}
		// NPCs get double the skill gain chance
		if !m.isPlayerCharacter {
			gc *= 2
		}
		toGain := 0
		if v <= 100 {
			// Always gain when below 10.0 skill, and make gains faster
			toGain = world.Random().Random(0, 4) + 1
		} else if world.Random().Random(0, 1000) < gc {
			// Just a chance of gain
			toGain = 1
		}
		if toGain > 0 {
			// Execute skill gain
			m.skills[which] = int16(v + toGain)
			if m.n != nil {
				m.n.UpdateSkill(which, uo.SkillLockUp, v+toGain)
			}
		}
	}
	// Determine if we can gain a stat
	if world.Random().Random(0, 100) >= 5 {
		// 5% chance of stat gain on every skill use
		return success
	}
	// TODO Consider total stat cap
	info := uo.SkillInfo[which]
	primaryStat := info.PrimaryStat
	secondaryStat := info.SecondaryStat
	statToConsider := primaryStat
	if world.Random().Random(0, 3) == 0 {
		statToConsider = secondaryStat
	}
	// TODO Consider stat locks
	sv := 0
	switch statToConsider {
	case uo.StatStrength:
		sv = m.baseStrength
	case uo.StatDexterity:
		sv = m.baseDexterity
	case uo.StatIntelligence:
		sv = m.baseIntelligence
	}
	if sv >= 100 {
		// Can't gain any more
		return success
	}
	// Apply stat gain
	switch statToConsider {
	case uo.StatStrength:
		m.baseStrength++
	case uo.StatDexterity:
		m.baseDexterity++
	case uo.StatIntelligence:
		m.baseIntelligence++
	}
	// If we've gotten this far we need to send a status update for the new stat
	world.Update(m)

	return success
}
