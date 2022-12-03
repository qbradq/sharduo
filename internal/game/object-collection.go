package game

// ObjectCollection is a collection of objects and is implemented as a light
// wrapper around a (possibly nil) slice of objects.
type ObjectCollection []Object

// Append returns c with an object added to the end of the collection.
func (c ObjectCollection) Append(o Object) ObjectCollection {
	idx := c.IndexOf(o)
	if idx < 0 {
		return append(c, o)
	}
	return c
}

// Remove returns c with the object removed from it.
func (c ObjectCollection) Remove(o Object) ObjectCollection {
	idx := c.IndexOf(o)
	if idx >= 0 {
		copy(c[idx:], c[idx+1:])
		return c[:len(c)-1]
	}
	return c
}

// IndexOf returns the index of the object within the collection, or -1 if the
// object is not in the collection.
func (c ObjectCollection) IndexOf(o Object) int {
	for idx, this := range c {
		if this == o {
			return idx
		}
	}
	return -1
}

// Contains returns true if the collection contains the given object.
func (c ObjectCollection) Contains(o Object) bool {
	return c.IndexOf(o) >= 0
}
