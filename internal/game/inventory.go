package game

import "github.com/qbradq/sharduo/lib/util"

// Inventory is a collection of items
type Inventory struct {
	c util.Slice[Item]
}

// Add tries to add an item to the inventory
func (n *Inventory) Add(item Item) bool {
	if n.c.Contains(item) {
		return false
	}
	n.c = n.c.Append(item)
	return true
}

// Remove removes the item from the inventory
func (n *Inventory) Remove(item Item) {
	n.c = n.c.Remove(item)
}

// Contains returns true if the collection contains the given item.
func (n *Inventory) Contains(item Item) bool {
	return n.c.Contains(item)
}
