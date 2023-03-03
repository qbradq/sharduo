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
	g.ReplyButton(768-63, 640-23, 1151, 1152, 1)
	g.PageButton(0, 0, 1147, 1148, 1)
	g.PageButton(0, 48, 1144, 1145, 2)
	g.Page(1)
	g.Label(128, 0, uo.HueDefault, "PAGE 1")
	g.Checkbox(96, 32, 210, 211, 101, false)
	g.Label(128, 32, uo.HueDefault, "Unchecked Check Box")
	g.Checkbox(96, 64, 210, 211, 102, true)
	g.Label(128, 64, uo.HueDefault, "Checked Check Box")
	g.Group(0)
	g.RadioButton(96, 96, 208, 209, 201, false)
	g.Label(128, 96, uo.HueDefault, "Group 0 Button 1")
	g.RadioButton(96, 128, 208, 209, 202, true)
	g.Label(128, 128, uo.HueDefault, "Group 0 Button 2")
	g.RadioButton(96, 160, 208, 209, 201, false)
	g.Label(128, 160, uo.HueDefault, "Group 0 Button 3")
	g.Group(1)
	g.RadioButton(96, 192, 208, 209, 201, false)
	g.Label(128, 192, uo.HueDefault, "Group 1 Button 1")
	g.RadioButton(96, 224, 208, 209, 202, false)
	g.Label(128, 224, uo.HueDefault, "Group 1 Button 2")
	g.RadioButton(96, 256, 208, 209, 201, true)
	g.Label(128, 256, uo.HueDefault, "Group 1 Button 3")
	g.Page(2)
	g.Label(128, 0, uo.HueDefault, "PAGE 2")

}
