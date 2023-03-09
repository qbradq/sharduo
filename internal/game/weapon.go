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
	Wearable

	// Skill returns the weapon skill this weapon uses
	Skill() uo.Skill
	// Animation returns the animation action code used for an attack
	AnimationAction() uo.AnimationAction
}

// BaseWeapon is the base implementatino of Weapon
type BaseWeapon struct {
	BaseWearable
	// The weapon skill to use
	skill uo.Skill
	// Animation action code to use on attack
	animation uo.AnimationAction
}

// ObjectType implements the Object interface.
func (w *BaseWeapon) ObjectType() marshal.ObjectType {
	return marshal.ObjectTypeWeapon
}

// Deserialize implements the util.Serializeable interface.
func (w *BaseWeapon) Deserialize(t *template.Template, create bool) {
	w.BaseWearable.Deserialize(t, create)
	w.skill = uo.Skill(t.GetNumber("Skill", int(uo.SkillWrestling)))
	w.animation = uo.AnimationAction(t.GetNumber("Animation", int(uo.AnimationActionSlash1H)))
}

// Skill implements the Weapon interface.
func (w *BaseWeapon) Skill() uo.Skill { return w.skill }

// AnimationAction implements the Weapon interface.
func (w *BaseWeapon) AnimationAction() uo.AnimationAction { return w.animation }
