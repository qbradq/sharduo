package serverpacket

// StartingLocation represents one starting location for new character creation.
type StartingLocation struct {
	// Name of the city.
	City string
	// Name of the building or area.
	Area string
}

// StartingLocations is the list of starting locations in correct order.
var StartingLocations = []StartingLocation{
	{"New Haven", "New Haven Bank"},
	{"Yew", "The Empath Abbey"},
	{"Minoc", "The Barnacle"},
	{"Britain", "The Wayfarer's Inn"},
	{"Moonglow", "The Scholars Inn"},
	{"Trinsic", "The Traveler's Inn"},
	{"Jhelom", "The Mercenary Inn"},
	{"Skara Brae", "The Falconer's Inn"},
	{"Vesper", "The Ironwood Inn"},
}
