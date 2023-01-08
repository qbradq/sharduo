package game

import (
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	ObjectFactory.RegisterCtor(func(v any) util.Serializeable { return &BaseWeapon{} })
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

// Serialize implements the util.Serializeable interface.
func (w *BaseWeapon) Serialize(f *util.TagFileWriter) {
	w.BaseWearable.Serialize(f)
	f.WriteNumber("Skill", int(w.skill))
}

// Deserialize implements the util.Serializeable interface.
func (w *BaseWeapon) Deserialize(f *util.TagFileObject) {
	w.BaseWearable.Deserialize(f)
	w.skill = uo.Skill(f.GetNumber("Skill", int(uo.SkillWrestling)))
}
