package commands

import (
	"github.com/qbradq/sharduo/internal/events"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

// Most GM-level commands live here

func init() {
	reg(&cmDesc{"account", nil, commandAccount, game.RoleGameMaster, "account", "Opens the account management GUMP for the targeted player"})
	reg(&cmDesc{"admin", nil, commandAdmin, game.RoleGameMaster, "admin", "Opens the admin GUMP"})
	reg(&cmDesc{"bank", nil, commandBank, game.RoleGameMaster, "bank", "Opens the bank box of the targeted mobile, if any"})
	reg(&cmDesc{"edit", nil, commandEdit, game.RoleGameMaster, "edit", "Opens the targeted object's edit GUMP if any"})
	reg(&cmDesc{"new", []string{"add"}, commandNew, game.RoleGameMaster, "new template_name [stack_amount]", "Creates a new item with an optional stack amount"})
	reg(&cmDesc{"remove", []string{"rem", "delete", "del"}, commandRemove, game.RoleGameMaster, "remove", "Removes the targeted object and all of its children from the game game.GetWorld()"})
	reg(&cmDesc{"set_hue", nil, commandSetHue, game.RoleGameMaster, "set_hue", "Sets the hue of an object"})
	reg(&cmDesc{"set_z", nil, commandSetZ, game.RoleGameMaster, "set_z", "Adjusts the Z location of the object"})
	reg(&cmDesc{"static", nil, commandStatic, game.RoleGameMaster, "static graphic_number", "Creates a new static object with the given graphic number"})
	reg(&cmDesc{"tame", nil, commandTame, game.RoleGameMaster, "tame", "makes you the control master of the targeted mobile"})
	reg(&cmDesc{"teleport", []string{"t"}, commandTeleport, game.RoleGameMaster, "teleport [x y|x y z|multi]", "Teleports you to the targeted location - optionally multiple times, or to the top Z of the given X/Y location, or to the absolute location"})
}

func commandBank(n game.NetState, args CommandArgs, cl string) {
	if n == nil || n.Mobile() == nil {
		return
	}
	n.TargetSendCursor(uo.TargetTypeObject, func(r *clientpacket.TargetResponse) {
		m, found := game.World.FindMobile(r.TargetObject)
		if !found {
			return
		}
		events.ExecuteEventHandler("OpenBankBox", m, n.Mobile(), nil)
	})
}

func commandNew(n game.NetState, args CommandArgs, cl string) {
	if n == nil {
		return
	}
	if len(args) < 2 || len(args) > 3 {
		n.Speech(n.Mobile(), "new command requires 2 or 3 arguments, got %d", len(args))
	}
	n.TargetSendCursor(uo.TargetTypeLocation, func(r *clientpacket.TargetResponse) {
		var m *game.Mobile
		var item *game.Item
		m = game.NewMobile(args[1])
		if m == nil {
			item = game.NewItem(args[1])
			if item == nil {
				n.Speech(n.Mobile(), "failed to create object with template %s", args[1])
				return
			}
		}
		if len(args) == 3 {
			if item == nil {
				n.Speech(n.Mobile(), "amount specified for non-item %s", args[1])
				return
			}
			v := args.Int(2)
			if v < 1 {
				v = 1
			}
			if item.TemplateName == "Check" {
				item.IArg = v
			} else {
				if !item.HasFlags(game.ItemFlagsStackable) {
					n.Speech(n.Mobile(), "amount specified for non-stackable item %s", args[1])
					return
				}
				item.Amount = v
			}
		}
		if r.TargetObject == uo.SerialZero {
			if item != nil {
				item.Location = r.Location.Bound()
				game.World.Map().AddItem(item, true)
			} else {
				m.Location = r.Location.Bound()
				game.World.Map().AddMobile(m, true)
			}
		} else if mob, found := game.World.FindMobile(r.TargetObject); found {
			if item != nil {
				mob.DropToBackpack(item, true)
			} else {
				n.Speech(n.Mobile(), "mobile targeted for new mobile")
				return
			}
		} else if c, found := game.World.FindItem(r.TargetObject); found {
			if c != nil && c.HasFlags(game.ItemFlagsContainer) {
				if item != nil {
					c.DropInto(item, true)
				} else {
					n.Speech(n.Mobile(), "container targeted for new mobile")
					return
				}
			} else if c != nil {
				if item != nil {
					item.Location = c.Location
					game.World.Map().AddItem(item, true)
				} else {
					m.Location = c.Location
					game.World.Map().AddMobile(m, true)
				}
			}
		}
	})
}

func commandTeleport(n game.NetState, args CommandArgs, cl string) {
	if n.Mobile() == nil {
		return
	}
	targeted := false
	multi := false
	l := uo.Point{}
	l.Z = uo.MapMaxZ
	if len(args) > 4 {
		n.Speech(n.Mobile(), "teleport command expects a maximum of 3 arguments")
		return
	}
	if len(args) == 4 {
		l.Z = args.Int(3)
	}
	if len(args) >= 3 {
		l.Y = args.Int(2)
		l.X = args.Int(1)
	}
	if len(args) == 2 {
		if args[1] == "multi" {
			targeted = true
			multi = true
		} else {
			n.Speech(n.Mobile(), "incorrect usage of teleport command. Use [teleport (multi|X Y|X Y Z)")
			return
		}
	}
	if len(args) == 1 {
		targeted = true
	}
	l = l.Bound()
	if !targeted {
		if l.Z == uo.MapMaxZ {
			floor, _ := game.World.Map().GetFloorAndCeiling(l, false, true)
			if floor == nil {
				n.Speech(n.Mobile(), "location has no floor")
				return
			}
			l.Z = floor.Z()
		}
		if !game.World.Map().TeleportMobile(n.Mobile(), l) {
			n.Speech(n.Mobile(), "something is blocking that location")
		}
		return
	}

	var fn func(*clientpacket.TargetResponse)
	fn = func(r *clientpacket.TargetResponse) {
		if n.Mobile() == nil {
			return
		}
		if !game.World.Map().TeleportMobile(n.Mobile(), r.Location) {
			n.Speech(n.Mobile(), "something is blocking that location")
		}
		if multi {
			n.TargetSendCursor(uo.TargetTypeLocation, fn)
		}
	}
	n.TargetSendCursor(uo.TargetTypeLocation, fn)
}

func commandStatic(n game.NetState, args CommandArgs, cl string) {
	if n == nil {
		return
	}
	if len(args) != 2 {
		n.Speech(nil, "usage: static item_id")
		return
	}
	g := uo.Graphic(args.Int(1))
	if g.IsNoDraw() {
		n.Speech(n.Mobile(), "refusing to create no-draw static 0x%04X", g)
	}
	n.TargetSendCursor(uo.TargetTypeLocation, func(r *clientpacket.TargetResponse) {
		i := game.NewItem("StaticItem")
		l := r.Location
		if i, found := game.World.FindItem(r.TargetObject); found {
			l.Z = i.Highest()
		}
		i.Graphic = g
		i.Location = r.Location
		game.World.Map().AddItem(i, true)
	})
}

func commandRemove(n game.NetState, args CommandArgs, cl string) {
	if n == nil {
		return
	}
	multi := len(args) > 1 && args[1] == "multi"
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		if m, found := game.World.FindMobile(tr.TargetObject); found {
			game.World.RemoveMobile(m)
		} else if i, found := game.World.FindItem(tr.TargetObject); found {
			game.World.RemoveItem(i)
		}
		if multi {
			commandRemove(n, args, cl)
		}
	})
}

func commandEdit(n game.NetState, args CommandArgs, cl string) {
	if n == nil || n.Mobile() == nil {
		return
	}
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		if m, found := game.World.FindMobile(tr.TargetObject); found {
			gumps.Edit(n.Mobile(), m)
		} else if i, found := game.World.FindItem(tr.TargetObject); found {
			gumps.Edit(n.Mobile(), i)
		}
	})
}

func commandSetHue(n game.NetState, args CommandArgs, cl string) {
	var hue uo.Hue
	if n == nil || n.Mobile() == nil {
		return
	}
	if len(args) == 2 {
		hue = uo.Hue(args.Int(1))
	} else if len(args) == 3 && args[1] == "partial" {
		hue = uo.Hue(args.Int(1)).SetPartial()
	} else {
		return
	}
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		if m, found := game.World.FindMobile(tr.TargetObject); found {
			m.Hue = hue
			game.World.UpdateMobile(m)
		} else if item, found := game.World.FindItem(tr.TargetObject); found {
			item.Hue = hue
			game.World.UpdateItem(item)
		}
	})
}

func commandTame(n game.NetState, args CommandArgs, cl string) {
	if n == nil || n.Mobile() == nil {
		return
	}
	if len(args) != 1 {
		return
	}
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		m, found := game.World.FindMobile(tr.TargetObject)
		if !found {
			return
		}
		m.ControlMaster = n.Mobile()
		m.AI = "Follow"
		m.AIGoal = n.Mobile()
	})
}

func commandSetZ(n game.NetState, args CommandArgs, cl string) {
	z := args.Int(1)
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		if m, found := game.World.FindMobile(tr.TargetObject); found {
			game.World.Map().RemoveMobile(m)
			m.Location.Z = z
			game.World.Map().AddMobile(m, true)
		} else if item, found := game.World.FindItem(tr.TargetObject); found {
			if item.Container != nil || item.Wearer != nil {
				return
			}
			game.World.Map().RemoveItem(item)
			item.Location.Z = z
			game.World.Map().AddItem(item, true)
		}
	})
}

func commandAdmin(n game.NetState, args CommandArgs, cl string) {
	n.GUMP(gumps.New("admin"), 0, 0)
}

func commandAccount(n game.NetState, args CommandArgs, cl string) {
	n.Speech(n.Mobile(), "Select player")
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		m, found := game.World.FindMobile(tr.TargetObject)
		if !found || m.NetState == nil || m.Account == nil {
			n.Speech(n.Mobile(), "Not a player")
			return
		}
		n.GUMP(gumps.New("account"), m.Serial, 0)
	})
}
