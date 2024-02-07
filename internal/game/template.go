package game

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/qbradq/sharduo/lib/uo"
)

// Template represents a JSON template in textual form and can handle its own
// inheritance resolution.
type Template struct {
	Name     string         // Name of the template
	Fields   map[string]any // All of the fields of the template
	Resolved bool           // If true this template has been fully resolved
}

// UnmarshalJSON implements the json.Unmarshal interface.
func (t *Template) UnmarshalJSON(in []byte) error {
	t.Fields = map[string]any{}
	if err := json.Unmarshal(in, &t.Fields); err != nil {
		return err
	}
	return nil
}

// Resolve resolves this template's inheritance chain. Calling this function on
// a template that is already fully resolved is a no-op.
func (t *Template) Resolve(templates map[string]*Template) {
	if t.Resolved {
		return
	}
	btn, found := t.Fields["BaseTemplate"].(string)
	if !found {
		// No base template to resolve
		t.Resolved = true
		return
	}
	bt, found := templates[btn]
	if !found {
		panic(fmt.Errorf("nonexistent base template %s", btn))
	}
	bt.Resolve(templates)
	for k, tv := range bt.Fields {
		// If the field appears in the base template and not in the current
		// template then the base template value will be copied
		btv, found := t.Fields[k]
		if !found {
			t.Fields[k] = tv
			continue
		}
		// If the field appears in both the base template and in the current
		// template we need to merge slices and maps
		switch bt := btv.(type) {
		case []any:
			// Prepend the base template elements
			t.Fields[k] = append(bt, t.Fields[k].([]any)...)
		case map[string]any:
			// Copy over base template elements that are not in this template
			for bk, bv := range bt {
				if _, found := t.Fields[k].(map[string]any)[bk]; !found {
					t.Fields[k] = bv
				}
			}
		}
	}
	t.Resolved = true
}

// constructPrototype copies template values to the object.
func (p *Template) constructPrototype(i any) {
	fn := func(s string) int64 {
		v, err := strconv.ParseInt(s, 0, 32)
		if err != nil {
			panic(err)
		}
		return v
	}
	ie := reflect.ValueOf(i).Elem()
	for k, tv := range p.Fields {
		fe := ie.FieldByName(k)
		if !fe.IsValid() {
			panic(fmt.Errorf("template specifies non-existent field %s", k))
		}
		switch v := tv.(type) {
		case string:
			if k == "Bounds" {
				fe.Set(reflect.ValueOf(uo.ParseBounds(v)))
			} else if k == "Layer" {
				fe.Set(reflect.ValueOf(uo.ParseLayer(v)))
			} else if k == "LootType" {
				fe.Set(reflect.ValueOf(uo.ParseLootType(v)))
			} else if k == "UseSkill" {
				fe.Set(reflect.ValueOf(uo.ParseSkill(v)))
			} else if k == "AnimationAction" {
				fe.Set(reflect.ValueOf(uo.ParseAnimationAction(v)))
			} else if k == "Body" {
				fe.SetUint(uint64(uo.FlexNumber([]byte(v))))
			} else if k == "BaseNotoriety" {
				fe.Set(reflect.ValueOf(uo.ParseNotoriety(v)))
			} else if fe.CanInt() {
				fe.SetInt(fn(v))
			} else if fe.CanUint() {
				fe.SetUint(uint64(fn(v)))
			} else {
				fe.Set(reflect.ValueOf(v))
			}
		case float64:
			if fe.CanFloat() {
				fe.SetFloat(v)
			} else if fe.CanInt() {
				fe.SetInt(int64(v))
			} else if fe.CanUint() {
				fe.SetUint(uint64(v))
			} else {
				panic(fmt.Errorf("number given for non-numeric field %s", k))
			}
		case bool:
			fe.SetBool(v)
		case []any:
			// If this is the flags array we need to join all the flags
			if k == "Flags" {
				var f ItemFlags
				for _, av := range v {
					f.AddByName(av.(string))
				}
				fe.SetUint(uint64(f))
				break
			}
			// If this is the post creation events slice we need to parse
			// the events
			if k == "PostCreationEvents" {
				for _, sv := range v {
					fe.Set(reflect.Append(fe, reflect.ValueOf(parsePostCreationEvent(sv.(string)))))
				}
				break
			}
			// If this is the context menu slice we need to parse the entries
			if k == "ContextMenu" {
				for _, sv := range v {
					fe.Set(reflect.Append(fe, reflect.ValueOf(parseCtxMenuEntry(sv.(string)))))
				}
				break
			}
			// Otherwise we just copy over the strings
			for _, sv := range v {
				fe.Set(reflect.Append(fe, reflect.ValueOf(sv)))
			}
		case map[string]any:
			// Copy over all map strings
			for sk, av := range v {
				fe.SetMapIndex(reflect.ValueOf(sk), reflect.ValueOf(av))
			}
		default:
			panic(fmt.Errorf("unhandled value type in field %s", k))
		}
	}
}
