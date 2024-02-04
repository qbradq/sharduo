package events

import (
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
)

// BankBalance says the source's account balance.
func BankBalance(receiver, source, v any) bool {
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	bb := sm.Equipment[uo.LayerBankBox]
	if bb == nil {
		return false
	}
	sm.NetState.Cliloc(receiver, 1042759, strconv.FormatInt(int64(bb.Gold), 10)) // Thy current bank balance is ~1_AMOUNT~ gold.
	return true
}

// BankCheck creates a bank check for the given amount.
func BankCheck(receiver, source, v any) bool {
	n, err := strconv.ParseInt(v.(string), 0, 32)
	if err != nil || n < 1000 {
		return false
	}
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	if !sm.ChargeGold(int(n)) {
		return false
	}
	check := game.NewItem("Check")
	check.IArg = int(n)
	if !sm.DropToBackpack(check, false) {
		if !sm.DropToBankBox(check, false) {
			sm.DropToFeet(check)
		}
		sm.NetState.Cliloc(receiver, 1042764, strconv.FormatInt(int64(n), 10)) // A check worth ~1_AMOUNT~ in gold has been placed in your bank box.
	} else {
		sm.NetState.Cliloc(receiver, 1042765, strconv.FormatInt(int64(n), 10)) // A check worth ~1_AMOUNT~ in gold has been placed in your backpack.
	}
	return true
}

// BankDeposit deposits the amount of gold into the bank.
func BankDeposit(receiver, source, v any) bool {
	n, err := strconv.ParseInt(v.(string), 0, 32)
	if err != nil || n < 1 {
		return false
	}
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	bp := sm.Equipment[uo.LayerBackpack]
	if !bp.ConsumeGold(int(n)) {
		return false
	}
	gc := game.NewItem("GoldCoin")
	gc.Amount = int(n)
	if sm.DropToBankBox(gc, true) {
		sm.NetState.Cliloc(receiver, 1042760, strconv.FormatInt(n, 10)) // Thou hast deposited ~1_AMOUNT~ gold.
		sm.NetState.Sound(0x02E6, sm.Location)
		return true
	}
	return false
}

// BankWithdraw withdraws the amount of gold from the bank.
func BankWithdraw(receiver, source, v any) bool {
	n, err := strconv.ParseInt(v.(string), 0, 32)
	if err != nil || n < 1 {
		return false
	}
	sm := source.(*game.Mobile)
	if sm.NetState == nil {
		return false
	}
	bb := sm.Equipment[uo.LayerBankBox]
	if !bb.ConsumeGold(int(n)) {
		return false
	}
	var gc *game.Item
	if n > 1000 {
		gc = game.NewItem("Check")
		gc.IArg = int(n)
	} else {
		gc = game.NewItem("GoldCoin")
		gc.Amount = int(n)
	}
	if !sm.DropToBackpack(gc, false) {
		sm.DropToFeet(gc)
	}
	sm.NetState.Sound(0x02E6, sm.Location)
	return true
}
