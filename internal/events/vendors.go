package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/internal/gumps"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("VendorBuy", vendorBuy)
	reg("VendorSell", vendorSell)
	reg("StablePet", stablePet)
	reg("ClaimAllPets", claimAllPets)
}

// vendorBuy opens the vendor buy screen for a vendor.
func vendorBuy(receiver, source, v any) bool {
	sm := source.(game.Mobile)
	if sm.NetState == nil {
		return false
	}
	rm := receiver.(game.Mobile)
	if sm.Location.XYDistance(rm.Location) > uo.MaxViewRange {
		return false
	}
	fsc := rm.Equipment[uo.LayerNPCBuyRestockContainer]
	if fsc == nil {
		// Mobile is not a vendor
		return false
	}
	bc := rm.Equipment[uo.LayerNPCBuyNoRestockContainer]
	if bc == nil {
		// Mobile is not a vendor
		return false
	}
	items := make([]serverpacket.ContentsItem, 0, len(fsc.Contents))
	for _, item := range fsc.Contents {
		items = append(items, serverpacket.ContentsItem{
			Serial:      item.Serial,
			Graphic:     item.Graphic,
			Amount:      999,
			Location:    uo.Point{},
			Container:   fsc.Serial,
			Hue:         item.Hue,
			Price:       uint32(item.Value),
			Description: item.DisplayName(),
		})
	}
	p := &serverpacket.VendorBuySequence{
		Vendor:       rm.Serial,
		ForSale:      fsc.Serial,
		Bought:       bc.Serial,
		ForSaleItems: items,
		BoughtItems:  nil,
	}
	sm.NetState.Send(p)
	return true
}

// vendorSell opens the vendor sell screen for a vendor.
func vendorSell(receiver, source, v any) bool {
	var items []serverpacket.ContentsItem
	var fn func(c *game.Item)
	fn = func(c *game.Item) {
		for _, i := range c.Contents {
			if i.Value < 1 {
				continue
			}
			if i.HasFlags(game.ItemFlagsContainer) {
				fn(i)
				continue
			}
			items = append(items, serverpacket.ContentsItem{
				Serial:      i.Serial,
				Graphic:     i.Graphic,
				Amount:      i.Amount,
				Location:    uo.Point{},
				Container:   c.Serial,
				Hue:         i.Hue,
				Price:       uint32(i.Value),
				Description: i.DisplayName(),
			})
		}
	}
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	rm := receiver.(*game.Mobile)
	bp := sm.Equipment[uo.LayerBackpack]
	items = make([]serverpacket.ContentsItem, 0, bp.ItemCount)
	fn(bp)
	if len(items) == 0 {
		game.World.Map().SendCliloc(rm, 1080012) // You have nothing I would be interested in.
		return false
	}
	sm.NetState.Send(&serverpacket.SellWindow{
		Vendor: rm.Serial,
		Items:  items,
	})
	return true
}

// stablePet presents a targeting cursor for stabling a pet
func stablePet(receiver, source, v any) bool {
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	sm.NetState.TargetSendCursor(uo.TargetTypeObject, func(tr *clientpacket.TargetResponse) {
		pm := game.World.FindMobile(tr.TargetObject)
		if pm == nil {
			sm.NetState.Cliloc(receiver, 1048053) // You can't stable that!
			return
		}
		if pm.Player {
			sm.NetState.Speech(receiver, "I believe there are inns in the area...")
			return
		}
		if pm.ControlMaster == sm {
			sm.NetState.Cliloc(receiver, 1048053) // You can't stable that!
			return
		}
		if err := sm.Stable(pm); err != nil {
			sm.NetState.Send(err.Packet())
			return
		}
		game.World.RemoveMobile(pm)
		sm.NetState.Cliloc(receiver, 1049677) // Your pet has been stabled.
	})
	return true
}

// claimAllPets presents a GUMP for claiming pets
func claimAllPets(receiver, source, v any) bool {
	sm := source.(*game.Mobile)
	if sm.NetState == nil || !sm.Player {
		return false
	}
	if len(sm.PlayerData.StabledPets) == 0 {
		sm.NetState.Cliloc(receiver, 1071150) // But, you have no animals stabled at the moment.
		return false
	}
	sm.NetState.GUMP(gumps.New("claim"), sm.Serial, 0)
	return true
}
