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
func VendorBuy(receiver, source game.Object, v any) {
	if source == nil || receiver == nil {
		return
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return
	}
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	w := rm.EquipmentInSlot(uo.LayerNPCBuyRestockContainer)
	if w == nil {
		// Mobile is not a vendor
		return
	}
	fsc, ok := w.(game.Container)
	if !ok {
		return
	}
	w = rm.EquipmentInSlot(uo.LayerNPCBuyNoRestockContainer)
	if w == nil {
		// Mobile is not a vendor
		return
	}
	bc, ok := w.(game.Container)
	if !ok {
		return
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
}

// VendorSell opens the vendor sell screen for a vendor.
func VendorSell(receiver, source game.Object, v any) {
	var items []serverpacket.ContentsItem
	var fn func(c game.Container)
	fn = func(c game.Container) {
		for _, i := range c.Contents() {
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
		return
	}
	if sm.NetState() == nil {
		return
	}
	rm, ok := receiver.(game.Mobile)
	if !ok {
		return
	}
	w := sm.EquipmentInSlot(uo.LayerBackpack)
	bp, ok := w.(game.Container)
	if !ok {
		return
	}
	items = make([]serverpacket.ContentsItem, 0, bp.ItemCount())
	fn(bp)
	sm.NetState().Send(&serverpacket.SellWindow{
		Vendor: rm.Serial(),
		Items:  items,
	})
}
