package game

import (
	"encoding/json"
	"fmt"

	"github.com/qbradq/sharduo/lib/uo"
)

// Equipment represents the collection of all equipped items on
// a mobile.
type Equipment [uo.LayerCount]*Item

// UnmarshalJSON implements the json.Marshaler interface.
func (e *Equipment) UnmarshalJSON(in []byte) error {
	tns := []string{}
	if err := json.Unmarshal(in, &tns); err != nil {
		return err
	}
	for _, tn := range tns {
		i := NewItem(tn)
		if i == nil {
			panic(fmt.Errorf("invalid template name %s", tn))
		}
		if !i.Layer.Valid() {
			return fmt.Errorf("invalid layer %d", i.Layer)
		}
		if (*e)[i.Layer] != nil {
			return fmt.Errorf("duplicate equipment in slot %d", i.Layer)
		}
		(*e)[i.Layer] = i
	}
	return nil
}
