package game

import (
	"fmt"
	"log"

	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("BaseContainer", marshal.ObjectTypeContainer, func() any { return &BaseContainer{} })
}

// ContainerObserver is implemented by anything that can be notified of changes
// to the contents of a container.
type ContainerObserver interface {
	// ContainerOpen sends the open container gump packet and all of the
	// contents of the container to the observer. The observer should keep track
	// of all open containers so they can close then when needed.
	ContainerOpen(Container)
	// ContainerClose closes the container on the client side as well as all
	// child containers this observer may be observing.
	ContainerClose(Container)
	// ContainerItemAdded notifies the observer of a new item in the container.
	ContainerItemAdded(Container, Item)
	// ContainerItemRemoved notifies the observer of a item being removed from
	// the container.
	ContainerItemRemoved(Container, Item)
	// ContainerItemOPLChanged notifies the observer of an item's OPL changing.
	ContainerItemOPLChanged(Container, Item)
	// ContainerRangeCheck asks the observer to close all out-of-range
	// containers.
	ContainerRangeCheck()
	// ContainerIsObserving returns true if the given container is being
	// observed by the observer.
	ContainerIsObserving(Object) bool
}

// Container is the interface all objects implement that can contain other
// other objects within an inventory.
type Container interface {
	Item
	// GumpGraphic returns the gump graphic of the item
	GumpGraphic() uo.GUMP
	// Contains returns true if the object is a direct child of this container,
	// or any child containers.
	Contains(Object) bool
	// ContentWeight returns the total weight of all contained items excluding
	// the weight of the container itself. Use game.Item.Weight to get the
	// weight of the container without the contents.
	ContentWeight() float32
	// ItemCount returns the total number of items within the container.
	ItemCount() int
	// Contents returns the contents of the container.
	Contents() []Item
	// Open is called by mobiles to open the container.
	Open(Mobile)
	// RemoveObserver removes an observer from this container's list of
	// containers.
	RemoveObserver(o ContainerObserver)
	// StopAllObservers forces all observers of this container to stop observing
	// it.
	StopAllObservers()
	// AdjustWeightAndCount adds to the cached total weight and item count of
	// the container.
	AdjustWeightAndCount(float32, int)
	// DropInto attempts to drop the item into this container, merging it with
	// any currently existing stacks. Returns true on success.
	DropInto(Item) bool
	// UpdateItem must be called when an item contained in this container
	// changes.
	UpdateItem(Item)
	// UpdateItemOPL must be called when an item contained in this container
	// has an OPL change.
	UpdateItemOPL(Item)
	// DropSound returns the default drop sound for the container.
	DropSound() uo.Sound
	// CountGold returns the amount of gold contained within this container and
	// all sub-containers.
	CountGold() int
	// ConsumeGold attempts to consume the given amount of gold from the
	// container and all sub-containers and will consume gold coins before
	// checks. It returns true if this was successful. ConsumeGold only modifies
	// items if it returns true.
	ConsumeGold(int) bool
}

// BaseContainer implements the base implementation of the Container interface.
type BaseContainer struct {
	BaseItem
	// Contents of the container
	contents util.Slice[Item]
	// Gump image to show for the container background
	gump uo.GUMP
	// Max weight this container can hold
	maxContainerWeight float32
	// Max items this container can hold
	maxContainerItems int
	// Cache of the current weight of all the items in the container
	contentWeight float32
	// Cache of the number of items in the container and all sub containers
	contentItems int
	// All of the observers of the container
	observers map[ContainerObserver]struct{}
	// Bounds of the container GUMP
	bounds uo.Bounds
	// Default sound of dropped objects
	dropSound uo.Sound
}

// ObjectType implements the Object interface.
func (i *BaseContainer) ObjectType() marshal.ObjectType { return marshal.ObjectTypeContainer }

// RemoveChildren implements the Object interface.
func (i *BaseContainer) RemoveChildren() {
	for _, c := range i.contents {
		Remove(c)
	}
}

// Marshal implements the marshal.Marshaler interface.
func (i *BaseContainer) Marshal(s *marshal.TagFileSegment) {
	i.BaseItem.Marshal(s)
	s.PutInt(0) // version
	l := len(i.contents)
	if l > 255 {
		log.Printf("warning: container %s has way too many items", i.serial.String())
		l = 255
	}
	s.PutByte(byte(l))
	for idx, o := range i.contents {
		// Sanity check
		if idx > 255 {
			break
		}
		s.PutObject(o)
	}
}

// Deserialize implements the util.Serializeable interface.
func (c *BaseContainer) Deserialize(t *template.Template, create bool) {
	c.BaseItem.Deserialize(t, create)
	c.gump = uo.GUMP(t.GetHex("Gump", uint32(0x003C)))
	c.maxContainerWeight = t.GetFloat("MaxContainerWeight", float32(uo.DefaultMaxContainerWeight))
	c.maxContainerItems = t.GetNumber("MaxContainerItems", uo.DefaultMaxContainerItems)
	c.bounds = t.GetBounds("Bounds", uo.Bounds{X: 44, Y: 65, W: 142, H: 94})
	c.dropSound = uo.Sound(t.GetNumber("DropSound", int(uo.SoundDefaultDrop)))
	// Contents
	for _, s := range t.GetObjectReferences("Contents") {
		i := Find[Item](s)
		if i == nil {
			continue
		}
		i.SetDropLocation(uo.RandomContainerLocation)
		c.ForceAddObject(i)
	}
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (i *BaseContainer) Unmarshal(s *marshal.TagFileSegment) {
	i.BaseItem.Unmarshal(s)
	_ = s.Int() // version
	l := int(s.Byte())
	for idx := 0; idx < l; idx++ {
		ium := s.Object()
		item, ok := ium.(Item)
		if !ok {
			panic("container object does not implement the Item interface")
		}
		i.contents = append(i.contents, item)
	}
}

// RecalculateStats implements the Object interface
func (c *BaseContainer) RecalculateStats() {
	c.InvalidateOPL()
	c.contentWeight = 0
	c.contentItems = len(c.contents)
	for _, item := range c.contents {
		item.RecalculateStats()
		c.contentWeight += item.Weight()
		c.contentItems++
		if container, ok := item.(Container); ok {
			c.contentItems += container.ItemCount()
			c.contentWeight += container.ContentWeight()
		}
	}
}

// GumpGraphic implements the Container interface.
func (c *BaseContainer) GumpGraphic() uo.GUMP { return c.gump }

// SingleClick implements the Object interface
func (c *BaseContainer) SingleClick(from Mobile) {
	// Default action is to send the name as over-head text
	if from.NetState() != nil {
		str := fmt.Sprintf("%s\n%d/%d items, %d/%d stones", c.DisplayName(),
			c.ItemCount(), c.maxContainerItems,
			int(c.contentWeight), int(c.maxContainerWeight))
		from.NetState().Speech(c, str)
	}
}

// Open implements the object interface.
func (c *BaseContainer) Open(m Mobile) {
	if m.NetState() == nil {
		return
	}
	observer, ok := m.NetState().(ContainerObserver)
	if !ok {
		return
	}
	// TODO access calculations
	if c.observers == nil {
		c.observers = make(map[ContainerObserver]struct{})
	}
	c.observers[observer] = struct{}{}
	observer.ContainerOpen(c)
}

// StopAllObservers forces all observers of this container to stop observing it.
func (c *BaseContainer) StopAllObservers() {
	for o := range c.observers {
		o.ContainerClose(c)
	}
}

// doRemove removes an object forcefully if requested
func (c *BaseContainer) doRemove(o Object, force bool) bool {
	// Only items go into containers
	item, ok := o.(Item)
	if !ok {
		return force
	}
	// This avoids a duplicate call to IndexOf
	oldLength := len(c.contents)
	c.contents = c.contents.Remove(item)
	if len(c.contents) == oldLength {
		return force
	}
	itemsRemoved := 1
	if container, ok := o.(Container); ok {
		itemsRemoved += container.ItemCount()
	}
	c.AdjustWeightAndCount(-item.Weight(), -itemsRemoved)
	// Broadcast the item removal
	for observer := range c.observers {
		observer.ContainerItemRemoved(c, item)
	}
	// Gold calculations
	if c.TemplateName() != "PlayerBankBox" {
		if item.TemplateName() == "GoldCoin" {
			if mobile, ok := RootParent(c).(Mobile); ok {
				mobile.AdjustGold(-item.Amount())
			}
		} else if check, ok := item.(*Check); ok {
			// Check support
			if mobile, ok := RootParent(c).(Mobile); ok {
				mobile.AdjustGold(-check.CheckAmount())
			}
		}
	} else {
		if item.TemplateName() == "GoldCoin" {
			if mobile, ok := RootParent(c).(Mobile); ok {
				mobile.AdjustBankGold(-item.Amount())
			}
		} else if check, ok := item.(*Check); ok {
			// Check support
			if mobile, ok := RootParent(c).(Mobile); ok {
				mobile.AdjustBankGold(-check.CheckAmount())
			}
		}
	}
	return true
}

// RemoveObject implements the Object interface.
func (c *BaseContainer) RemoveObject(o Object) bool {
	return c.doRemove(o, false)
}

// ForceRemoveObject implements the Object interface.
func (c *BaseContainer) ForceRemoveObject(o Object) {
	c.doRemove(o, true)
}

// AddObject implements the Object interface.
func (c *BaseContainer) AddObject(o Object) bool {
	// Only items go into containers
	item, ok := o.(Item)
	if !ok {
		return false
	}
	// Something is very wrong
	if c.contents.Contains(item) {
		item.SetDropLocation(o.Location())
		return false
	}
	addedItems := 1
	addedWeight := item.Weight()
	if container, ok := item.(Container); ok {
		addedItems += container.ItemCount()
	}
	// Container weight check
	if c.maxContainerWeight > 0 && c.contentWeight+addedWeight > c.maxContainerWeight {
		item.SetDropLocation(o.Location())
		return false
	}
	// Max items check
	if c.maxContainerItems > 0 && c.contentItems+addedItems > c.maxContainerItems {
		item.SetDropLocation(o.Location())
		return false
	}
	c.ForceAddObject(o)
	return true
}

// ForceAddObject implements the Container interface.
func (c *BaseContainer) ForceAddObject(o Object) {
	// Sanity check: we can't allow the number of direct children of this
	// container to exceed 255 or the Marshal() call will panic. This could
	// happen if a player finds a reliable way of force-dropping items to their
	// bank box or backpack. This will leak the object but at this point we are
	// dealing with an exploit / system abuse and don't really care.
	if len(c.contents) > 255 {
		Remove(o)
		return
	}
	if o == nil {
		return
	}
	o.SetParent(c)
	item, ok := o.(Item)
	if !ok {
		return
	}
	if c.contents.Contains(item) {
		return
	}
	// Determine if we should try to auto-stack the item
	l := item.DropLocation()
	if item.Stackable() && l.X == uo.RandomDropX && l.Y == uo.RandomDropY {
		for _, i := range c.contents {
			if i.CanCombineWith(item) && i.Combine(item) {
				c.InvalidateOPL()
				return
			}
		}
		// Never successfully combined with a stack, continue to normal
		// placement.
	}
	addedItems := 1
	addedWeight := item.Weight()
	if container, ok := item.(Container); ok {
		addedItems += container.ItemCount()
	}
	// Location bounding
	if l.X == uo.RandomDropX {
		l.X = int16(world.Random().Random(int(c.bounds.X), int(c.bounds.X+c.bounds.W-1)))
	}
	if l.Y == uo.RandomDropY {
		l.Y = int16(world.Random().Random(int(c.bounds.Y), int(c.bounds.Y+c.bounds.H-1)))
	}
	if l.X < c.bounds.X {
		l.X = c.bounds.X
	}
	if l.X >= c.bounds.X+c.bounds.W {
		l.X = c.bounds.X + c.bounds.W - 1
	}
	if l.Y < c.bounds.Y {
		l.Y = c.bounds.Y
	}
	if l.Y >= c.bounds.Y+c.bounds.H {
		l.Y = c.bounds.Y + c.bounds.H - 1
	}
	// Add the item to our contents
	item.SetLocation(l)
	c.contents = c.contents.Append(item)
	c.AdjustWeightAndCount(addedWeight, addedItems)
	// Let all the observers know about the new item
	for observer := range c.observers {
		observer.ContainerItemAdded(c, item)
	}
	// Gold calculations
	if c.TemplateName() != "PlayerBankBox" {
		if item.TemplateName() == "GoldCoin" {
			if mobile, ok := RootParent(c).(Mobile); ok {
				mobile.AdjustGold(item.Amount())
			}
		} else if check, ok := item.(*Check); ok {
			// Check support
			if mobile, ok := RootParent(c).(Mobile); ok {
				mobile.AdjustGold(check.CheckAmount())
			}
		}
	} else {
		if item.TemplateName() == "GoldCoin" {
			if mobile, ok := RootParent(c).(Mobile); ok {
				mobile.AdjustBankGold(item.Amount())
			}
		} else if check, ok := item.(*Check); ok {
			// Check support
			if mobile, ok := RootParent(c).(Mobile); ok {
				mobile.AdjustBankGold(check.CheckAmount())
			}
		}
	}
}

// InsertObject implements the Object interface.
func (c *BaseContainer) InsertObject(obj any) {
	i, ok := obj.(Item)
	if !ok {
		return
	}
	i.SetDropLocation(uo.RandomContainerLocation)
	c.ForceAddObject(i)
}

// Contains implements the Container interface.
func (c *BaseContainer) Contains(o Object) bool {
	// Only items go into containers
	otherItem, ok := o.(Item)
	if !ok {
		return false
	}
	for _, item := range c.contents {
		if item.Serial() == otherItem.Serial() {
			return true
		}
		if container, ok := item.(Container); ok {
			if container.Contains(o) {
				return true
			}
		}
	}
	return false
}

// ContentWeight implements the Container interface.
func (c *BaseContainer) ContentWeight() float32 { return c.contentWeight }

// ItemCount implements the Container interface.
func (c *BaseContainer) ItemCount() int { return c.contentItems }

// Weight implements the Object interface.
func (c *BaseContainer) Weight() float32 {
	return c.BaseItem.Weight() + c.ContentWeight()
}

// AdjustWeightAndCount implements the Container interface.
func (c *BaseContainer) AdjustWeightAndCount(w float32, n int) {
	c.InvalidateOPL()
	c.contentWeight += w
	c.contentItems += n
	if c.templateName == "PlayerBankBox" {
		// We are a mobile's bank box, don't propagate up
		return
	}
	if container, ok := c.parent.(Container); ok {
		// We are a sub-container, propagate the adjustment up
		container.AdjustWeightAndCount(w, n)
	} else if mobile, ok := c.parent.(Mobile); ok {
		mobile.InvalidateOPL()
		if mobile.IsItemOnCursor() && mobile.ItemInCursor().Serial() == c.Serial() {
			// We are being held by a mobile's cursor, don't need to do anything
			return
		}
		// We are a mobile's backpack, send the weight adjustment up
		mobile.AdjustWeight(w)
	}
}

// Contents implements the Container interface.
func (c *BaseContainer) Contents() []Item {
	return c.contents
}

// RemoveObserver implements the Container interface.
func (c *BaseContainer) RemoveObserver(o ContainerObserver) {
	if c.observers != nil {
		delete(c.observers, o)
		if len(c.observers) == 0 {
			c.observers = nil
		}
	}
}

// DropInto implements the Container interface.
func (c *BaseContainer) DropInto(i Item) bool {
	// Over weight limit
	if c.ContentWeight()+i.Weight() > c.maxContainerWeight {
		return false
	}
	// Try to stack
	if i.Stackable() {
		for _, other := range c.contents {
			// Same item type and enough stack capacity to accept
			if other.Stackable() &&
				i.Hue() == other.Hue() &&
				i.TemplateName() == other.TemplateName() &&
				other.Amount()+i.Amount() <= int(uo.MaxStackAmount) {
				c.AdjustWeightAndCount(i.Weight(), 0)
				other.SetAmount(other.Amount() + i.Amount())
				if c.TemplateName() != "PlayerBankBox" && i.TemplateName() == "GoldCoin" {
					if mobile, ok := RootParent(c).(Mobile); ok {
						mobile.AdjustGold(i.Amount())
					}
				}
				Remove(i)
				world.Update(other)
				return true
			}
		}
	}
	// Try to drop to bag
	return c.AddObject(i)
}

// AppendContextMenuEntries implements the Object interface.
func (c *BaseContainer) AppendContextMenuEntries(m *ContextMenu, src Mobile) {
	c.BaseItem.AppendContextMenuEntries(m, src)
	if src.CanAccess(c) {
		m.Append("OpenContainer", 3000362)
	}
}

// UpdateItem implements the Container interface.
func (c *BaseContainer) UpdateItem(i Item) {
	for o := range c.observers {
		o.ContainerItemAdded(c, i)
	}
}

// UpdateItemOPL implements the Container interface.
func (c *BaseContainer) UpdateItemOPL(i Item) {
	for o := range c.observers {
		o.ContainerItemOPLChanged(c, i)
	}
}

// AppendOPLEntries implements the Object interface.
func (c *BaseContainer) AppendOPLEntires(r Object, p *serverpacket.OPLPacket) {
	c.BaseItem.AppendOPLEntires(r, p)
	p.Append(fmt.Sprintf("%d/%d items, %d/%d stones", c.contentItems,
		c.maxContainerItems, int(c.ContentWeight()), int(c.maxContainerWeight)),
		true)
}

// DropSound implements the Container interface.
func (c *BaseContainer) DropSound() uo.Sound { return c.dropSound }

// CountGold implements the Container interface.
func (c *BaseContainer) CountGold() int {
	count := 0
	for _, i := range c.contents {
		if i.TemplateName() == "GoldCoin" {
			count += i.Amount()
		} else if check, ok := i.(*Check); ok {
			count += check.CheckAmount()
		} else if sub, ok := i.(Container); ok {
			count += sub.CountGold()
		}
	}
	return count
}

// ConsumeGold implements the Container interface.
func (c *BaseContainer) ConsumeGold(amount int) bool {
	var total int
	var toRemove []Item
	var collectGold func(Container) bool
	collectGold = func(cont Container) bool {
		for _, i := range cont.Contents() {
			if sub, ok := i.(Container); ok {
				if collectGold(sub) {
					return true
				}
				continue
			}
			if i.TemplateName() != "GoldCoin" {
				continue
			}
			if total+i.Amount() >= amount {
				i.Consume(amount - total)
				if m, ok := RootParent(c).(Mobile); ok {
					if c.TemplateName() != "PlayerBankBox" {
						m.AdjustGold(-(amount - total))
					} else {
						m.AdjustBankGold(-(amount - total))
					}
				}
				for _, i := range toRemove {
					Remove(i)
				}
				return true
			}
			total += i.Amount()
			toRemove = append(toRemove, i)
		}
		return false
	}
	var collectCheck func(Container) bool
	collectCheck = func(cont Container) bool {
		for _, i := range cont.Contents() {
			if sub, ok := i.(Container); ok {
				if collectCheck(sub) {
					return true
				}
				continue
			}
			check, ok := i.(*Check)
			if !ok {
				continue
			}
			if total+check.CheckAmount() >= amount {
				check.ConsumeGold(amount - total)
				if m, ok := RootParent(c).(Mobile); ok {
					if c.TemplateName() != "PlayerBankBox" {
						m.AdjustGold(-(amount - total))
					} else {
						m.AdjustBankGold(-(amount - total))
					}
				}
				for _, i := range toRemove {
					Remove(i)
				}
				return true
			}
			total += check.CheckAmount()
			toRemove = append(toRemove, i)
		}
		return false
	}
	return collectGold(c) || collectCheck(c)
}
