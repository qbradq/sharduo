package game

import (
	"github.com/qbradq/sharduo/lib/marshal"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("BaseWeapon", marshal.ObjectTypeWeapon, func() any { return &BaseWeapon{} })
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

// Deserialize implements the util.Serializeable interface.
func (w *BaseWeapon) Deserialize(t *template.T) {
	w.BaseWearable.Deserialize(t)
	w.skill = uo.Skill(t.GetNumber("Skill", int(uo.SkillWrestling)))
}
