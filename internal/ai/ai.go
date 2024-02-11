package ai

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

// AIModel is the interface all AI models implement
type AIModel interface {
	// Act is called every tick and is responsible for making the mobile take all actions
	Act(*game.Mobile, uo.Time)
	// Target is called every fifteen seconds and is responsible for target and goal selection
	Target(*game.Mobile, uo.Time)
}

// ctor is a Thinker constructor function
type ctor func() AIModel

// Thinker registrar
var treg *util.Registry[string, ctor] = util.NewRegistry[string, ctor]("ai")

func reg(name string, fn ctor) {
	treg.Add(name, fn)
}

// GetModel returns the named Thinker or nil if it does not exist
func GetModel(which string) AIModel {
	fn, _ := treg.Get(which)
	return fn()
}
