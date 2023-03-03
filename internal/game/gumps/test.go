package gumps

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

// Test is a test GUMP that uses all of the GUMP features I am aware of and
// support.
type Test struct {
	BaseGUMP
}

// Layout implements the GUMP interface.
func (g *Test) Layout(target, param game.Object) {
	g.Background(-17, -17, 768+34, 640+34, 9250)
	// g.ReplyButton(768-63, 640-23, 1151, 1152, 1)
	// g.PageButton(0, 0, 1147, 1148, 1)
	// g.PageButton(0, 48, 1144, 1145, 0)
	// g.Page(1)
	// g.Checkbox(96, 0, 210, 211, 101, false)
	g.Label(128, 0, uo.HueDefault, "Unchecked Check Box")
	// g.Checkbox(96, 32, 210, 211, 101, true)
	g.Label(128, 32, uo.HueDefault, "Checked Check Box")
}
