package commands

import (
	"github.com/qbradq/sharduo/internal/events"
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

// Most GM-level commands live here

func init() {
	regcmd(&cmdesc{"account", nil, commandAccount, game.RoleGameMaster, "account", "Opens the account management GUMP for the targeted player"})
	regcmd(&cmdesc{"admin", nil, commandAdmin, game.RoleGameMaster, "admin", "Opens the admin GUMP"})
	regcmd(&cmdesc{"bank", nil, commandBank, game.RoleGameMaster, "bank", "Opens the bank box of the targeted mobile, if any"})
	regcmd(&cmdesc{"edit", nil, commandEdit, game.RoleGameMaster, "edit", "Opens the targeted object's edit GUMP if any"})
	regcmd(&cmdesc{"new", []string{"add"}, commandNew, game.RoleGameMaster, "new template_name [stack_amount]", "Creates a new item with an optional stack amount"})
	regcmd(&cmdesc{"remove", []string{"rem", "delete", "del"}, commandRemove, game.RoleGameMaster, "remove", "Removes the targeted object and all of its children from the game game.GetWorld()"})
	regcmd(&cmdesc{"sethue", nil, commandSetHue, game.RoleGameMaster, "sethue", "Sets the hue of an object"})
	regcmd(&cmdesc{"setz", nil, commandSetZ, game.RoleGameMaster, "setz", "Adjusts the Z location of the object"})
	regcmd(&cmdesc{"static", nil, commandStatic, game.RoleGameMaster, "static graphic_number", "Creates a new static object with the given graphic number"})
	regcmd(&cmdesc{"tame", nil, commandTame, game.RoleGameMaster, "tame", "makes you the control master of the targeted mobile"})
	regcmd(&cmdesc{"teleport", []string{"tele"}, commandTeleport, game.RoleGameMaster, "teleport [x y|x y z|multi]", "Teleports you to the targeted location - optionally multiple times, or to the top Z of the given X/Y location, or to the absolute location"})
}

func commandBank(n game.NetState, args CommandArgs, cl string) {
	if n == nil || n.Mobile() == nil {
		return
	}
	n.TargetSendCursor(uo.TargetTypeObject, func(r *clientpacket.TargetResponse) {
		m := game.Find[game.Mobile](r.TargetObject)
		if m == nil {
			return
		}
		events.OpenBankBox(m, n.Mobile(), nil)
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
		o := template.Create[game.Object](args[1])
		if o == nil {
			n.Speech(n.Mobile(), "failed to create object with template %s", args[1])
			return
		}
		if len(args) == 3 {
			v := args.Int(2)
			if v < 1 {
				v = 1
			}
			if check, ok := o.(*game.Check); ok {
				check.SetCheckAmount(v)
			} else {
				item, ok := o.(game.Item)
				if !ok {
					n.Speech(n.Mobile(), "amount specified for non-item %s", args[1])
					return
				}
				if !item.Stackable() {
					n.Speech(n.Mobile(), "amount specified for non-stackable item %s", args[1])
					return
				}
				item.SetAmount(v)
			}
		}
		m := game.Find[game.Mobile](r.TargetObject)
		if m != nil {
			// If we targeted a mobile force the item into the backpack
			if item, ok := o.(game.Item); ok {
				item.SetDropLocation(uo.RandomContainerLocation)
				m.DropToBackpack(o, true)
			}
		} else {
			o.SetLocation(r.Location)
			// Try to add the object to the map legit, but if that fails just
			// force it so we don't leak it.
			if !game.GetWorld().Map().AddObject(o) {
				game.GetWorld().Map().ForceAddObject(o)
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
	l := uo.Location{}
	l.Z = uo.MapMaxZ
	if len(args) > 4 {
		n.Speech(n.Mobile(), "teleport command expects a maximum of 3 arguments")
		return
	}
	if len(args) == 4 {
		l.Z = int8(args.Int(3))
	}
	if len(args) >= 3 {
		l.Y = int16(args.Int(2))
		l.X = int16(args.Int(1))
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
			floor, _ := game.GetWorld().Map().GetFloorAndCeiling(l, false, true)
			if floor == nil {
				n.Speech(n.Mobile(), "location has no floor")
				return
			}
			l.Z = floor.Z()
		}
		if !game.GetWorld().Map().TeleportMobile(n.Mobile(), l) {
			n.Speech(n.Mobile(), "something is blocking that location")
		}
		return
	}

	var fn func(*clientpacket.TargetResponse)
	fn = func(r *clientpacket.TargetResponse) {
		if n.Mobile() == nil {
			return
		}
		if !game.GetWorld().Map().TeleportMobile(n.Mobile(), r.Location) {
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
		i := template.Create[*game.StaticItem]("StaticItem")
		if i == nil {
			n.Speech(n.Mobile(), "StaticItem template not found")
			return
		}
		l := r.Location
		if i := game.Find[game.Item](r.TargetObject); i != nil {
			l.Z = i.Highest()
		}
		i.SetBaseGraphic(g)
		i.SetLocation(r.Location)
		game.GetWorld().Map().ForceAddObject(i)
	})
}

func commandRemove(n game.NetState, args CommandArgs, cl string) {
	if n == nil {
		return
	}
	multi := len(args) > 1 && args[1] == "multi"
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		o := game.GetWorld().Find(tr.TargetObject)
		game.Remove(o)
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
		o := game.GetWorld().Find(tr.TargetObject)
		gumps.Edit(n.Mobile(), o)
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
		o := game.GetWorld().Find(tr.TargetObject)
		if o == nil {
			return
		}
		o.SetHue(hue)
		game.GetWorld().Update(o)
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
		m := game.Find[game.Mobile](tr.TargetObject)
		if m == nil {
			return
		}
		m.SetControlMaster(n.Mobile())
	})
}

func commandSetZ(n game.NetState, args CommandArgs, cl string) {
	z := int8(args.Int(1))
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		o := game.Find[game.Object](tr.TargetObject)
		if o == nil {
			return
		}
		l := o.Location()
		l.Z = z
		if o.Parent() == nil {
			game.GetWorld().Map().ForceRemoveObject(o)
			o.SetLocation(l)
			game.GetWorld().Map().ForceAddObject(o)
		} else {
			o.SetLocation(l)
		}
	})
}

func commandAdmin(n game.NetState, args CommandArgs, cl string) {
	n.GUMP(gumps.New("admin"), nil, nil)
}

func commandAccount(n game.NetState, args CommandArgs, cl string) {
	n.Speech(n.Mobile(), "Select player")
	n.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		m := game.Find[game.Mobile](tr.TargetObject)
		if m == nil || m.NetState() == nil || m.NetState().Account() == nil {
			n.Speech(n.Mobile(), "Not a player")
			return
		}
		n.GUMP(gumps.New("account"), m, nil)
	})
}
