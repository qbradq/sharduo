package game

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/qbradq/sharduo/data"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
	"golang.org/x/image/colornames"
)

// LoadMobilePrototypes loads all mobile prototypes.
func LoadMobilePrototypes() {
	// Load all mobile templates
	errors := false
	templates := map[string]*Template{}
	for _, err := range data.Walk("templates/mobiles", func(s string, b []byte) []error {
		// Load prototypes
		ts := map[string]*Template{}
		if err := json.Unmarshal(b, &ts); err != nil {
			return []error{err}
		}
		// Prototype prep
		for k, t := range ts {
			// Check for duplicates
			if _, duplicate := templates[k]; duplicate {
				return []error{fmt.Errorf("duplicate mobile prototype %s", k)}
			}
			// Initialize non-zero default values
			t.Name = k
			t.Fields["TemplateName"] = k
			templates[k] = t
		}
		return nil
	}) {
		errors = true
		log.Printf("error: during mobile prototype load: %v", err)
	}
	if errors {
		panic("errors during mobile prototype load")
	}
	// Resolve all base templates and construct their item prototypes
	for tn, t := range templates {
		t.Resolve(templates)
		m := &Mobile{
			Object: Object{
				Events: map[string]string{},
			},
		}
		t.constructPrototype(m)
		mobilePrototypes[tn] = m
	}
}

// constructMobile creates a new item from the named template.
func constructMobile(which string) *Mobile {
	p := mobilePrototypes[which]
	if p == nil {
		return nil
	}
	m := &Mobile{}
	*m = *p
	// Attach new player data struct if needed
	if m.Player {
		m.PlayerData = NewPlayerData()
	}
	// Establish initial cache values
	m.Hits = m.MaxHits
	m.Mana = m.MaxMana
	m.Stamina = m.MaxStamina
	m.Strength = m.BaseStrength
	m.Dexterity = m.BaseDexterity
	m.Intelligence = m.BaseIntelligence
	for i := range m.Skills {
		m.Skills[i] = m.BaseSkills[i]
	}
	// Generate equipment set
	for _, tn := range m.EquipmentSet {
		i := NewItem(tn)
		if i == nil {
			panic(fmt.Errorf("failed to create equipment item %s", tn))
		}
		if !m.Equip(i) {
			panic(fmt.Errorf("failed to equip item %s", tn))
		}
	}
	// Generate inventory items
	for _, tn := range m.Inventory {
		i := NewItem(tn)
		if i == nil {
			panic(fmt.Errorf("failed to create inventory item %s", tn))
		}
		m.DropToBackpack(i, true)
	}
	// Execute post creation events
	for _, e := range m.PostCreationEvents {
		if !e.Execute(m) {
			panic(fmt.Errorf("failed to execute post creation event %s", e.EventName))
		}
	}
	return m
}

// NewMobile creates a new mobile and adds it to the world datastores.
func NewMobile(which string) *Mobile {
	m := constructMobile(which)
	if m == nil {
		return nil
	}
	World.Add(m)
	return m
}

// mobilePrototypes contains all mobile prototypes.
var mobilePrototypes = map[string]*Mobile{}

// Mobile describes a thinking actor.
type Mobile struct {
	Object
	// Static values
	Body          uo.Body      // Body to use for this mobile
	BaseNotoriety uo.Notoriety // Base notoriety level for this mobile
	EquipmentSet  []string     // Set of items to generate along with this mobile and equip
	Inventory     []string     // Set of items to add to the mobile's backpack at creation time
	// Persistent values
	Female           bool                 // If true the mobile is female
	Player           bool                 // If true this is a player's character
	PlayerData       *PlayerData          // Player specific data, only valid if Player is true
	BaseStrength     int                  // Base strength value
	BaseDexterity    int                  // Base dexterity value
	BaseIntelligence int                  // Base intelligence value
	MaxHits          int                  // Current max hit points
	MaxMana          int                  // Current max mana
	MaxStamina       int                  // Current max stamina
	BaseSkills       [uo.SkillCount]int   // Base skill values in tenths of a percent
	Equipment        [uo.LayerCount]*Item // Current equipment set
	// Transient values
	Account       *Account                // Connected account if any
	NetState      NetState                // Connected net state if any
	ControlMaster *Mobile                 // Mobile that is currently commanding this mobile
	Cursor        *Item                   // The item in the mobile's cursor if any
	Hits          int                     // Current hit points
	Mana          int                     // Current mana
	Stamina       int                     // Current stamina
	Strength      int                     // Current strength value
	Dexterity     int                     // Current dexterity value
	Intelligence  int                     // Current intelligence value
	Skills        [uo.SkillCount]int      // Current skill values in tenths of a percent
	Running       bool                    // If true the mobile is running
	Weight        float64                 // Current weight of all equipment plus the contents of the backpack
	MaxWeight     float64                 // Max carry weight of the mobile
	ViewRange     int                     // Range at which items are reported to the client, valid values are [5-18]
	StandingOn    uo.CommonObject         // Object the mobile is standing on
	AI            string                  // Name of the AI routine to run during mobile think
	AIGoal        *Mobile                 // What mobile we are paying attention to at the moment
	opl           *serverpacket.OPLPacket // Cached OPLPacket
	oplInfo       *serverpacket.OPLInfo   // Cached OPLInfo packet
	lastStepTime  uo.Time                 // Time of the last step taken
}

// Write writes the persistent data of the item to w.
func (m *Mobile) Write(w io.Writer) {
	util.PutUInt32(w, 0)                          // Version
	util.PutString(w, m.TemplateName)             // Template name
	util.PutUInt32(w, uint32(m.Serial))           // Serial
	util.PutPoint(w, m.Location)                  // Location
	util.PutByte(w, byte(m.Facing))               // Facing
	util.PutUInt16(w, uint16(m.Hue))              // Hue
	util.PutBool(w, m.Female)                     // Female flag
	util.PutBool(w, m.Player)                     // Player flag
	util.PutUInt16(w, uint16(m.BaseStrength))     // Strength
	util.PutUInt16(w, uint16(m.BaseDexterity))    // Dexterity
	util.PutUInt16(w, uint16(m.BaseIntelligence)) // Intelligence
	util.PutUInt16(w, uint16(m.MaxHits))          // Max hits
	util.PutUInt16(w, uint16(m.MaxMana))          // Max mana
	util.PutUInt16(w, uint16(m.MaxStamina))       // Max stamina
	for _, v := range m.BaseSkills {              // Skills
		util.PutUInt16(w, uint16(v))
	}
	for _, e := range m.Equipment { // Equipment
		if e == nil {
			util.PutBool(w, false)
			continue
		}
		util.PutBool(w, true)
		e.Write(w)
	}
	if m.Cursor != nil { // Item in cursor
		util.PutBool(w, true)
		m.Cursor.Write(w)
	} else {
		util.PutBool(w, false)
	}
	if m.Player { // Player data
		m.PlayerData.Write(w)
	}
}

// NewMobileFromReader reads the persistent data of the mobile from r and
// returns the mobile. It also inserts the mobile into the world datastores.
func NewMobileFromReader(r io.Reader) *Mobile {
	_ = util.GetUInt32(r)   // Version
	tn := util.GetString(r) // Template name
	m := constructMobile(tn)
	m.TemplateName = tn
	m.Serial = uo.Serial(util.GetUInt32(r))     // Serial
	m.Location = util.GetPoint(r)               // Location
	m.Facing = uo.Direction(util.GetByte(r))    // Facing
	m.Hue = uo.Hue(util.GetUInt16(r))           // Hue
	m.Female = util.GetBool(r)                  // Female flag
	m.Player = util.GetBool(r)                  // Player flag
	m.BaseStrength = int(util.GetUInt16(r))     // Strength
	m.BaseDexterity = int(util.GetUInt16(r))    // Dexterity
	m.BaseIntelligence = int(util.GetUInt16(r)) // Intelligence
	m.MaxHits = int(util.GetUInt16(r))          // Max hits
	m.MaxMana = int(util.GetUInt16(r))          // Max mana
	m.MaxStamina = int(util.GetUInt16(r))       // Max stamina
	for i := range m.BaseSkills {               // Skills
		m.BaseSkills[i] = int(util.GetUInt16(r))
	}
	for layer := range m.Equipment { // Equipment
		if util.GetBool(r) {
			m.Equipment[layer] = NewItemFromReader(r)
		}
	}
	if util.GetBool(r) { // Item in cursor
		m.DropToFeet(NewItemFromReader(r))
	}
	if m.Player { // Player data
		m.PlayerData = NewPlayerDataFromReader(r)
	}
	// Establish sane defaults for variables that need non-zero default values
	m.ViewRange = uo.MaxViewRange
	return m
}

// RecalculateStats recalculates all internal cache states.
func (m *Mobile) RecalculateStats() {
	m.Strength = m.BaseStrength
	m.Dexterity = m.BaseDexterity
	m.Intelligence = m.BaseIntelligence
	if m.Player {
		m.MaxHits = m.Strength/2 + 50
		m.MaxMana = m.Intelligence
		m.MaxStamina = m.Dexterity
		m.MaxWeight = float64(int(float64(m.Strength)*3.5 + 40))
	}
	m.Hits = m.MaxHits
	m.Mana = m.MaxMana
	m.Stamina = m.MaxStamina
	m.Skills = m.BaseSkills // Note to self, this does an array copy
	m.Weight = 0
	for layer, e := range m.Equipment {
		if layer < int(uo.LayerFirstValid) || layer == int(uo.LayerBankBox) {
			continue
		}
		if layer > int(uo.LayerLastValid) {
			break
		}
		m.Weight += e.Weight
		if e.HasFlags(ItemFlagsContainer) {
			e.RecalculateStats()
			m.Weight += e.ContainedWeight
		}
	}
}

// OPLPackets constructs new OPL packets if needed and returns cached packets.
func (m *Mobile) OPLPackets() (*serverpacket.OPLPacket, *serverpacket.OPLInfo) {
	if m.opl == nil {
		m.opl = &serverpacket.OPLPacket{
			Serial: m.Serial,
		}
		// Base mobile properties
		m.opl.AppendColor(colornames.White, m.DisplayName(), false)
		m.opl.Compile()
		m.oplInfo = &serverpacket.OPLInfo{
			Serial: m.Serial,
			Hash:   m.opl.Hash,
		}
	}
	return m.opl, m.oplInfo
}

// Flags returns the compile [uo.MobileFlags] value for this mobile.
func (m *Mobile) Flags() uo.MobileFlags {
	ret := uo.MobileFlagNone
	if m.Female {
		ret |= uo.MobileFlagFemale
	}
	return ret
}

// NotorietyFor returns the [uo.Notoriety] value for mobile other from the
// perspective of this mobile.
func (m *Mobile) NotorietyFor(other *Mobile) uo.Notoriety {
	if other.Player {
		return uo.NotorietyInnocent
	}
	return m.BaseNotoriety
}

// CanSee returns true if the object can be seen by this mobile. This function
// *does not* test for line of sight.
func (m *Mobile) CanSee(o *Object) bool {
	switch o.Visibility {
	case uo.VisibilityVisible:
		return true
	case uo.VisibilityInvisible:
		return false
	case uo.VisibilityHidden:
		return false
	case uo.VisibilityStaff:
		if m.Account == nil {
			return false
		}
		return m.Account.HasRole(RoleGameMaster)
	case uo.VisibilityNone:
		return false
	}
	return false
}

// AfterUnmarshalOntoMap is called after all of the tiles, statics, items and
// mobiles have been loaded and placed on the map. It updates internal states
// that are required for proper movement and control.
func (m *Mobile) AfterUnmarshalOntoMap() {
	// TODO Stub
}

// ContextMenuPacket returns a new context menu packet.
func (m *Mobile) ContextMenuPacket(p *ContextMenu, mob *Mobile) {
	for _, e := range m.ContextMenu {
		p.Append(e.Event, e.Cliloc)
	}
}

// AfterMove handles things that happen every time a mobile steps such as
// stamina decay.
func (m *Mobile) AfterMove() {
	// Max weight checks
	w := int(m.Weight)
	mw := int(m.MaxWeight)
	if w > mw {
		sc := w - mw
		m.Stamina -= sc
		if m.Stamina < 0 {
			m.Stamina = 0
		}
		World.UpdateMobile(m)
	}
	// Check for containers that we need to close
	if m.NetState != nil {
		m.NetState.ContainerRangeCheck()
	}
}

// CanAccess returns true if this mobile is allowed access to the given object.
func (m *Mobile) CanAccess(obj any) bool {
	// TODO Stub
	return true
}

// DropToFeet drops the item at the mobile's feet forcefully.
func (m *Mobile) DropToFeet(i *Item) {
	i.Location = m.Location
	World.Map().AddItem(i, true)
}

// HasLineOfSight returns true if there is line of sight between this mobile and
// the given object.
func (m *Mobile) HasLineOfSight(target any) bool {
	l := uo.Point{
		X: m.Location.X,
		Y: m.Location.Y,
		Z: m.Location.Z + uo.PlayerHeight, // Use our eye position, not the foot position
	}
	var t uo.Point
	switch o := target.(type) {
	case *Mobile:
		t = o.Location
		t.Z += uo.PlayerHeight // Look other mobiles in the eye
	case *Item:
		t = MapLocation(o)
	}
	return World.Map().LineOfSight(l, t)
}

// InvalidateOPL schedules an OPL update.
func (m *Mobile) InvalidateOPL() {
	m.opl = nil
	m.oplInfo = nil
	World.UpdateMobileOPLInfo(m)
}

// AdjustWeight adjusts the mobile's cache of their current carry weight.
func (m *Mobile) AdjustWeight(n float64) {
	m.Weight += n
	m.InvalidateOPL()
	World.UpdateMobile(m)
}

// ChargeGold consumes gold from the mobile's backpack, then bank until the
// amount has been charged. Returns false if there is not enough gold.
func (m *Mobile) ChargeGold(n int) bool {
	bp := m.Equipment[uo.LayerBackpack]
	bb := m.Equipment[uo.LayerBankBox]
	if bp == nil || bb == nil {
		return false
	}
	if bp.Gold+bb.Gold < n {
		return false
	}
	bg := n - bp.Gold
	if bg < 1 {
		bp.ConsumeGold(n)
	} else {
		bp.ConsumeGold(n - bg)
		bb.ConsumeGold(bg)
	}
	return true
}

// DropToBackpack drops the item into the mobile's backpack, forcefully if
// requested. Returns true on success.
func (m *Mobile) DropToBackpack(i *Item, force bool) bool {
	dp := m.Equipment[uo.LayerBackpack]
	if dp == nil {
		return false
	}
	if err := dp.DropInto(i, force); err != nil {
		return false
	}
	return true
}

// DropToBankBox drops the item into the mobile's bank box, forcefully if
// requested. Returns true on success.
func (m *Mobile) DropToBankBox(i *Item, force bool) bool {
	bb := m.Equipment[uo.LayerBankBox]
	if bb == nil {
		return false
	}
	if err := bb.DropInto(i, force); err != nil {
		return false
	}
	return true
}

// CanBeCommandedBy returns true if this mobile will accept control commands
// from s.
func (m *Mobile) CanBeCommandedBy(s *Mobile) bool {
	return m.ControlMaster == s
}

// Dismount implements the Mobile interface.
func (m *Mobile) Dismount() {
	mi := m.Equipment[uo.LayerMount]
	m.UnEquip(mi)
	mi.MArg.Location = m.Location
	mi.MArg.Facing = m.Facing
	World.Map().AddMobile(mi.MArg, true)
	World.RemoveItem(mi)
}

// Mount mounts the given mobile on m.
func (m *Mobile) Mount(mount *Mobile) {
	if mount == nil || m.Equipment[uo.LayerMount] != nil {
		return
	}
	mi := NewItem("MountItem")
	mi.SetBaseGraphicForBody(mount.Body)
	mi.Hue = mount.Hue
	mi.MArg = mount
	if !m.Equip(mi) {
		log.Println("warning: failed to equip mount item")
		World.RemoveItem(mi)
	}
	World.Map().RemoveMobile(mount)
	World.UpdateMobile(m)
}

// Equip attempts to equip an item, returns true on success.
func (m *Mobile) Equip(i *Item) bool {
	if !i.Layer.Valid() {
		return false
	}
	if m.Equipment[i.Layer] != nil {
		return false
	}
	m.Equipment[i.Layer] = i
	i.Wearer = m
	m.Weight += i.Weight + i.ContainedWeight
	for _, om := range World.Map().NetStatesInRange(m.Location, 0) {
		om.NetState.WornItem(i, m)
	}
	return true
}

// UnEquip attempts to remove the item, returns true on success.
func (m *Mobile) UnEquip(i *Item) bool {
	if m.Equipment[i.Layer] != i {
		return false
	}
	m.Equipment[i.Layer] = nil
	i.Wearer = nil
	m.Weight -= i.Weight + i.ContainedWeight
	for _, om := range World.Map().NetStatesInRange(m.Location, 0) {
		om.NetState.RemoveItem(i)
	}
	return true
}

// InBank returns true if the item is somewhere within the mobile's bank box.
func (m *Mobile) InBank(i *Item) bool {
	bb := m.Equipment[uo.LayerBankBox]
	for {
		if i.Container == nil {
			return false
		}
		if i.Container == bb {
			return true
		}
		i = i.Container
	}
}

// Stable adds the mobile to this player's stables. Calling this on a non-player
// mobile is a no-op.
func (m *Mobile) Stable(om *Mobile) *UOError {
	if !m.Player {
		return &UOError{
			Message: "Mobile.Stable() called for non-player mobile",
		}
	}
	if len(m.PlayerData.StabledPets) >= MaxStabledPets {
		return &UOError{
			Cliloc: 1042565, // You have too many pets in the stables!
		}
	}
	m.PlayerData.StabledPets = append(m.PlayerData.StabledPets, om)
	return nil
}

// Equipped returns true if the given item is currently being worn by this
// mobile.
func (m *Mobile) Equipped(i *Item) bool {
	if !i.Layer.Valid() {
		return false
	}
	return m.Equipment[i.Layer] == i
}

// SkillCheck implements the Mobile interface.
func (m *Mobile) SkillCheck(which uo.Skill, min, max int) bool {
	if which > uo.SkillLast {
		return false
	}
	// Get the skill value and look for corner cases
	v := m.Skills[which]
	if v < min {
		// No chance
		return false
	}
	if v >= max {
		// No challenge
		return true
	}
	// Calculate success
	spread := max - min
	chance := ((v - min) * 1000) / (max - min)
	success := false
	if util.Random(0, spread) < chance {
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
		if !m.Player {
			gc *= 2
		}
		toGain := 0
		if v <= 100 {
			// Always gain when below 10.0 skill, and make gains faster
			toGain = util.Random(0, 4) + 1
		} else if util.Random(0, 1000) < gc {
			// Just a chance of gain
			toGain = 1
		}
		if toGain > 0 {
			// Execute skill gain
			m.Skills[which] = v + toGain
			if m.NetState != nil {
				m.NetState.UpdateSkill(which, uo.SkillLockUp, v+toGain)
			}
		}
	}
	// Determine if we can gain a stat
	if util.Random(0, 100) >= 5 {
		// 5% chance of stat gain on every skill use
		return success
	}
	// TODO Consider total stat cap
	info := uo.SkillInfo[which]
	primaryStat := info.PrimaryStat
	secondaryStat := info.SecondaryStat
	statToConsider := primaryStat
	if util.Random(0, 3) == 0 {
		statToConsider = secondaryStat
	}
	// TODO Consider stat locks
	sv := 0
	switch statToConsider {
	case uo.StatStrength:
		sv = m.BaseStrength
	case uo.StatDexterity:
		sv = m.BaseDexterity
	case uo.StatIntelligence:
		sv = m.BaseIntelligence
	}
	if sv >= 100 {
		// Can't gain any more
		return success
	}
	// Apply stat gain
	switch statToConsider {
	case uo.StatStrength:
		m.BaseStrength++
	case uo.StatDexterity:
		m.BaseDexterity++
	case uo.StatIntelligence:
		m.BaseIntelligence++
	}
	// If we've gotten this far we need to send a status update for the new stat
	World.UpdateMobile(m)
	return success
}

// CanTakeStep returns true if the mobile can take a step.
func (m *Mobile) CanTakeStep() bool {
	var rd uo.Time
	if m.Equipment[uo.LayerMount] == nil {
		if !m.Running {
			rd = uo.WalkDelay
		} else {
			rd = uo.RunDelay
		}
	} else {
		if !m.Running {
			rd = uo.MountedWalkDelay
		} else {
			rd = uo.MountedRunDelay
		}
	}
	d := World.Time() - m.lastStepTime
	return d >= rd
}

// Step steps the mobile in the given direction, returning true on success.
func (m *Mobile) Step(d uo.Direction) bool {
	f := m.Facing
	m.Facing = d
	ret := World.Map().MoveMobile(m, d)
	if !ret {
		m.Facing = f
	} else {
		m.lastStepTime = World.Time()
	}
	return ret
}

// ExecuteEvent executes the named event handler if any is configured. Returns
// true if the handler was found and also returned true.
func (m *Mobile) ExecuteEvent(which string, s, v any) bool {
	hn, ok := m.Events[which]
	if !ok {
		return false
	}
	return ExecuteEventHandler(hn, m, s, v)
}
