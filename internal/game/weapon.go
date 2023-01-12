package game

import (
	"github.com/qbradq/sharduo/internal/marshal"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseWeapon{} })
	objectCtors[marshal.ObjectTypeWeapon] = func() Object { return &BaseWeapon{} }
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

// TypeName implements the util.Serializeable interface.
func (w *BaseWeapon) TypeName() string {
	return "BaseWeapon"
}

// ObjectType implements the Object interface.
func (w *BaseWeapon) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeWeapon
}

// Serialize implements the util.Serializeable interface.
func (w *BaseWeapon) Serialize(f *util.TagFileWriter) {
	w.BaseWearable.Serialize(f)
	f.WriteNumber("Skill", int(w.skill))
}

// Marshal implements the marshal.Marshaler interface.
func (w *BaseWeapon) Marshal(s *marshal.TagFileSegment) {
	w.BaseWearable.Marshal(s)
	s.PutTag(marshal.TagRequiredSkill, byte(w.skill))
}

// Deserialize implements the util.Serializeable interface.
func (w *BaseWeapon) Deserialize(f *util.TagFileObject) {
	w.BaseWearable.Deserialize(f)
	w.skill = uo.Skill(f.GetNumber("Skill", int(uo.SkillWrestling)))
}

// Unmarshal implements the marshal.Unmarshaler interface.
func (w *BaseWeapon) Unmarshal(to *marshal.TagObject) {
	w.BaseWearable.Unmarshal(to)
	w.skill = uo.Skill(to.Tags.Byte(marshal.TagRequiredSkill, byte(uo.SkillSwordsmanship)))
}
