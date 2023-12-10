package events

import (
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

// BankBalance says the source's account balance.
func BankBalance(receiver, source game.Object, v any) bool {
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	sm.NetState().Cliloc(receiver, 1042759, strconv.FormatInt(int64(sm.BankGold()), 10)) // Thy current bank balance is ~1_AMOUNT~ gold.
	return true
}

// BankCheck creates a bank check for the given amount.
func BankCheck(receiver, source game.Object, v any) bool {
	n, err := strconv.ParseInt(v.(string), 0, 32)
	if err != nil || n < 1000 {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	if !sm.ChargeGold(int(n)) {
		return false
	}
	check := template.Create[*game.Check]("Check")
	check.SetCheckAmount(int(n))
	if !sm.DropToBackpack(check, false) {
		if !sm.DropToBankBox(check, false) {
			sm.DropToFeet(check)
		}
		sm.NetState().Cliloc(receiver, 1042764, strconv.FormatInt(int64(n), 10)) // A check worth ~1_AMOUNT~ in gold has been placed in your bank box.
	} else {
		sm.NetState().Cliloc(receiver, 1042765, strconv.FormatInt(int64(n), 10)) // A check worth ~1_AMOUNT~ in gold has been placed in your backpack.
	}
	return true
}

// BankDeposit deposits the amount of gold into the bank.
func BankDeposit(receiver, source game.Object, v any) bool {
	n, err := strconv.ParseInt(v.(string), 0, 32)
	if err != nil || n < 1 {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	bpo := sm.EquipmentInSlot(uo.LayerBackpack)
	if bpo == nil {
		return false
	}
	bp, ok := bpo.(game.Container)
	if !ok || !bp.ConsumeGold(int(n)) {
		return false
	}
	gc := template.Create[game.Item]("GoldCoin")
	gc.SetAmount(int(n))
	if sm.DropToBankBox(gc, true) {
		sm.NetState().Cliloc(receiver, 1042760, strconv.FormatInt(n, 10)) // Thou hast deposited ~1_AMOUNT~ gold.
		sm.NetState().Sound(0x02E6, sm.Location())
		return true
	}
	return false
}

// BankWithdraw withdraws the amount of gold from the bank.
func BankWithdraw(receiver, source game.Object, v any) bool {
	n, err := strconv.ParseInt(v.(string), 0, 32)
	if err != nil || n < 1 {
		return false
	}
	sm, ok := source.(game.Mobile)
	if !ok || sm.NetState() == nil {
		return false
	}
	bbo := sm.EquipmentInSlot(uo.LayerBankBox)
	if bbo == nil {
		return false
	}
	bb, ok := bbo.(game.Container)
	if !ok || !bb.ConsumeGold(int(n)) {
		return false
	}
	var gc game.Item
	if n > 1000 {
		gc = template.Create[game.Item]("Check")
		gc.(*game.Check).SetCheckAmount(int(n))
	} else {
		gc = template.Create[game.Item]("GoldCoin")
		gc.SetAmount(int(n))
	}
	if !sm.DropToBackpack(gc, false) {
		sm.DropToFeet(gc)
	}
	sm.NetState().Sound(0x02E6, sm.Location())
	return true
}
