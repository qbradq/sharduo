package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/serverpacket"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("VendorBuy", VendorBuy)
	reg("VendorSell", VendorSell)
}

// VendorBuy opens the vendor buy screen for a vendor.
func VendorBuy(receiver, source game.Object, v any) bool {
	if source == nil || receiver == nil {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	if game.RootParent(sm).Location().XYDistance(rm.Location()) > uo.MaxViewRange {
		return false
	}
	w := rm.EquipmentInSlot(uo.LayerNPCBuyRestockContainer)
	if w == nil {
		// Mobile is not a vendor
		return false
	}
	fsc, ok := w.(game.Container)
	if !ok {
		return false
	}
	w = rm.EquipmentInSlot(uo.LayerNPCBuyNoRestockContainer)
	if w == nil {
		// Mobile is not a vendor
		return false
	}
	bc, ok := w.(game.Container)
	if !ok {
		return false
	}
	items := make([]serverpacket.ContentsItem, 0, len(fsc.Contents()))
	for _, item := range fsc.Contents() {
		items = append(items, serverpacket.ContentsItem{
			Serial:        item.Serial(),
			Graphic:       item.BaseGraphic(),
			GraphicOffset: item.GraphicOffset(),
			Amount:        999,
			Location:      uo.Location{},
			Container:     fsc.Serial(),
			Hue:           item.Hue(),
			Price:         uint32(item.Value()),
			Description:   item.DisplayName(),
		})
	}
	p := &serverpacket.VendorBuySequence{
		Vendor:       receiver.Serial(),
		ForSale:      fsc.Serial(),
		Bought:       bc.Serial(),
		ForSaleItems: items,
		BoughtItems:  nil,
	}
	sm.NetState().Send(p)
	return true
}

// VendorSell opens the vendor sell screen for a vendor.
func VendorSell(receiver, source game.Object, v any) bool {
	var items []serverpacket.ContentsItem
	var fn func(c game.Container)
	fn = func(c game.Container) {
		for _, i := range c.Contents() {
			if i.Value() < 1 {
				continue
			}
			if oc, ok := i.(game.Container); ok {
				fn(oc)
				continue
			}
			items = append(items, serverpacket.ContentsItem{
				Serial:        i.Serial(),
				Graphic:       i.BaseGraphic(),
				GraphicOffset: i.GraphicOffset(),
				Amount:        i.Amount(),
				Location:      uo.Location{},
				Container:     c.Serial(),
				Hue:           i.Hue(),
				Price:         uint32(i.Value()),
				Description:   i.DisplayName(),
			})
		}
	}
	sm, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if sm.NetState() == nil {
		return false
	}
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return false
	}
	w := sm.EquipmentInSlot(uo.LayerBackpack)
	bp, ok := w.(game.Container)
	if !ok {
		return false
	}
	items = make([]serverpacket.ContentsItem, 0, bp.ItemCount())
	fn(bp)
	if len(items) == 0 {
		game.GetWorld().Map().SendCliloc(rm, uo.SpeechNormalRange, 1080012) // You have nothing I would be interested in.
		return false
	}
	sm.NetState().Send(&serverpacket.SellWindow{
		Vendor: rm.Serial(),
		Items:  items,
	})
	return true
}
