package game

// chunk manages the data for a single chunk of the map.
type chunk struct {
	Items   []*Item   // List of all items within the chunk
	Mobiles []*Mobile // List of all mobiles within the chunk
}
