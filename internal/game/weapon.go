package game

import (
	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.Add("BaseWeapon", func() Object { return &BaseWeapon{} })
	marshal.RegisterCtor(marshal.ObjectTypeWeapon, func() interface{} { return &BaseWeapon{} })
}

// Weapon is the interface all weapons implement
type Weapon interface {
	// Skill returns the weapon skill this weapon uses
	Skill() uo.Skill
}

// BaseWeapon is the base implementatino of Weapon
type BaseWeapon struct {
	BaseWearable
	// skill is the weapon skill to use
	skill uo.Skill
}

// ObjectType implements the Object interface.
func (w *BaseWeapon) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeWeapon
}

// Marshal implements the marshal.Marshaler interface.
func (w *BaseWeapon) Marshal(s *marshal.TagFileSegment) {
	w.BaseWearable.Marshal(s)
	s.PutTag(marshal.TagRequiredSkill, marshal.TagValueByte, byte(w.skill))
}

// Deserialize implements the util.Serializeable interface.
func (w *BaseWeapon) Deserialize(f *util.TagFileObject) {
	w.BaseWearable.Deserialize(f)
	w.skill = uo.Skill(f.GetNumber("Skill", int(uo.SkillWrestling)))
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (w *BaseWeapon) Unmarshal(to *marshal.TagObject) {
	w.BaseWearable.Unmarshal(to)
	w.skill = uo.Skill(to.Tags.Byte(marshal.TagRequiredSkill))
}
