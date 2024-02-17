package ai

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	game.GetModel = GetModel
}

// ctor is a Thinker constructor function
type ctor func() game.AIModel

// Thinker registrar
var tReg *util.Registry[string, ctor] = util.NewRegistry[string, ctor]("ai")

func reg(name string, fn ctor) {
	tReg.Add(name, fn)
}

// GetModel returns the named Thinker or nil if it does not exist
func GetModel(which string) game.AIModel {
	fn, _ := tReg.Get(which)
	return fn()
}
