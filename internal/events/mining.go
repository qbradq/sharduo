package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/clientpacket"
	"github.com/qbradq/sharduo/lib/template"
	"github.com/qbradq/sharduo/lib/uo"
)

func init() {
	reg("BeginMining", BeginMining)
	reg("ContinueMining", ContinueMining)
	reg("FinishMining", FinishMining)
	reg("SmeltOre", SmeltOre)
}

// Registry of active miners
var regMiners = map[uo.Serial]struct{}{}

func startMiningLoop(miner game.Mobile, tool game.Weapon, p *clientpacket.TargetResponse) {
	// Register the miner blindly
	regMiners[miner.Serial()] = struct{}{}
	// Sanity checks
	if !miner.IsEquipped(tool) {
		miner.NetState().Cliloc(nil, 1149764) // You must have that equipped to use it.
		delete(regMiners, miner.Serial())
		return
	}
	if miner.IsMounted() {
		miner.NetState().Cliloc(nil, 501864) // You can't dig while riding or flying.
		delete(regMiners, miner.Serial())
		return
	}
	if miner.Location().XYDistance(p.Location) > 2 {
		miner.NetState().Cliloc(nil, 500251) // That location is too far away.
		delete(regMiners, miner.Serial())
		return
	}
	t := game.GetWorld().Map().GetTile(p.Location.X, p.Location.Y)
	if _, minable := mountainAndCaveTiles[t.BaseGraphic()]; !minable {
		miner.NetState().Cliloc(nil, 501863) // You can't mine that.
		delete(regMiners, miner.Serial())
		return
	}
	if !game.GetWorld().Map().HasOre(p.Location) {
		miner.NetState().Cliloc(nil, 503040) // There is no metal here to mine.
		delete(regMiners, miner.Serial())
		return
	}
	// Animation and sound
	game.GetWorld().Map().PlayAnimation(miner, uo.AnimationTypeAttack, tool.AnimationAction())
	game.NewTimer(12, "ContinueMining", tool, miner, true, p)
}

func BeginMining(receiver, source game.Object, v any) bool {
	if receiver == nil || source == nil {
		return false
	}
	miner, ok := source.(game.Mobile)
	if !ok || miner.NetState() == nil {
		return false
	}
	tool, ok := receiver.(game.Weapon)
	if !ok {
		return false
	}
	// Sanity checks
	if miner.NetState() == nil {
		return false
	}
	if !miner.NetState().TakeAction() {
		miner.NetState().Cliloc(nil, 500119) // You must wait to perform another action.
		return false
	}
	if !miner.IsEquipped(tool) {
		miner.NetState().Cliloc(nil, 1149764) // You must have that equipped to use it.
		return false
	}
	if miner.IsMounted() {
		miner.NetState().Cliloc(nil, 501864) // You can't dig while riding or flying.
		return false
	}
	if _, found := regMiners[miner.Serial()]; found {
		miner.NetState().Speech(nil, "You are already mining.")
		return false
	}
	// Targeting
	miner.NetState().TargetSendCursor(uo.TargetTypeLocation, func(p *clientpacket.TargetResponse) {
		startMiningLoop(miner, tool, p)
	})
	return true
}

func ContinueMining(receiver, source game.Object, v any) bool {
	if receiver == nil || source == nil {
		return false
	}
	p := v.(*clientpacket.TargetResponse)
	miner, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if miner.NetState() == nil {
		delete(regMiners, miner.Serial())
		return false
	}
	tool, ok := receiver.(game.Weapon)
	if !ok {
		delete(regMiners, miner.Serial())
		return false
	}
	// Sanity checks
	if !miner.IsEquipped(tool) {
		miner.NetState().Cliloc(nil, 1149764) // You must have that equipped to use it.
		delete(regMiners, miner.Serial())
		return false
	}
	if miner.IsMounted() {
		miner.NetState().Cliloc(nil, 501864) // You can't dig while riding or flying.
		delete(regMiners, miner.Serial())
		return false
	}
	if miner.Location().XYDistance(p.Location) > 2 {
		miner.NetState().Cliloc(nil, 500251) // That location is too far away.
		delete(regMiners, miner.Serial())
		return false
	}
	// Play the hit sound
	s := uo.Sound(0x125)
	if game.GetWorld().Random().RandomBool() {
		s = 0x126
	}
	game.GetWorld().Map().PlaySound(s, p.Location)
	// Queue up the last hit
	game.GetWorld().Map().PlayAnimation(miner, uo.AnimationTypeAttack, tool.AnimationAction())
	game.NewTimer(12, "FinishMining", receiver, source, true, p)
	return true
}

func FinishMining(receiver, source game.Object, v any) bool {
	if receiver == nil || source == nil {
		return false
	}
	p := v.(*clientpacket.TargetResponse)
	miner, ok := source.(game.Mobile)
	if !ok {
		return false
	}
	if miner.NetState() == nil {
		delete(regMiners, miner.Serial())
		return false
	}
	tool, ok := receiver.(game.Weapon)
	if !ok {
		delete(regMiners, miner.Serial())
		return false
	}
	// Sanity checks
	if !miner.IsEquipped(tool) {
		miner.NetState().Cliloc(nil, 1149764) // You must have that equipped to use it.
		delete(regMiners, miner.Serial())
		return false
	}
	if miner.IsMounted() {
		miner.NetState().Cliloc(nil, 501864) // You can't dig while riding or flying.
		delete(regMiners, miner.Serial())
		return false
	}
	if miner.Location().XYDistance(p.Location) > 2 {
		miner.NetState().Cliloc(nil, 500251) // That location is too far away.
		delete(regMiners, miner.Serial())
		return false
	}
	if !game.GetWorld().Map().HasOre(p.Location) {
		miner.NetState().Cliloc(nil, 503040) // There is no metal here to mine.
		delete(regMiners, miner.Serial())
		return false
	}
	// Play the hit sound
	s := uo.Sound(0x125)
	if game.GetWorld().Random().RandomBool() {
		s = 0x126
	}
	game.GetWorld().Map().PlaySound(s, p.Location)
	// Skill check
	if miner.SkillCheck(uo.SkillMining, 0, 1000) {
		ore := template.Create[game.Item]("IronOre")
		if ore == nil {
			// Something very wrong
			return false
		}
		ore.SetAmount(game.GetWorld().Map().ConsumeOre(p.Location, 2))
		if !miner.DropToBackpack(ore, false) {
			ore.SetLocation(miner.Location())
			game.GetWorld().Map().SetNewParent(ore, nil)
			miner.NetState().Cliloc(nil, 503045) // You dig some ore and but you have no place to put it.
		} else {
			miner.NetState().Cliloc(nil, 503044) // You dig some ore and put it in your backpack.
		}
	} else {
		miner.NetState().Cliloc(nil, 503043) // You loosen some rocks but fail to find any useable ore.
	}
	// TODO Item durability
	// Continue mining the spot if the player is still logged in
	if miner.NetState() != nil {
		startMiningLoop(miner, tool, p)
	} else {
		delete(regMiners, miner.Serial())
	}
	return true
}

func SmeltOre(receiver, source game.Object, v any) bool {
	smelter, ok := source.(game.Mobile)
	if !ok || smelter.NetState() == nil {
		// Something is very wrong
		return false
	}
	// The only scenario in which we would reject the request is if the ore
	// belongs to a mobile other than the smelter.
	root := game.RootParent(receiver)
	if root.Serial().IsMobile() && root.Serial() != smelter.Serial() {
		smelter.NetState().Cliloc(nil, 500685) // You can't use that, it belongs to someone else.
		return false
	}
	if !game.GetWorld().Map().Query(source.Location(), 3, forgeItemSet) {
		smelter.NetState().Cliloc(nil, 500420) // You are not near a forge.
		return false
	}
	ore, ok := receiver.(game.Item)
	if !ok {
		// Something is very wrong
		return false
	}
	if !smelter.SkillCheck(uo.SkillMining, 0, 750) {
		// Skill check failed, burn some ore
		smelter.NetState().Cliloc(nil, 501989) // You burn away the impurities but are left with no useable metal.
		ore.Consume(1)
		return true
	}
	ingot := template.Create[game.Item]("IronIngot")
	if ingot == nil {
		// Something very bad
		return false
	}
	ingot.SetAmount(ore.Amount() * 2)
	game.Remove(ore)
	if !smelter.DropToBackpack(ingot, false) {
		smelter.DropToFeet(ingot)
	}
	smelter.NetState().Cliloc(nil, 501988)
	return true
}

var forgeItemSet = map[uo.Graphic]struct{}{
	0x0FB1: {},
	0x197A: {},
	0x197E: {},
	0x1982: {},
	0x1986: {},
	0x198A: {},
	0x198E: {},
	0x1992: {},
	0x1996: {},
	0x199A: {},
	0x199E: {},
	0x19A2: {},
	0x19A6: {},
}

var mountainAndCaveTiles = map[uo.Graphic]struct{}{
	220:    {},
	221:    {},
	222:    {},
	223:    {},
	224:    {},
	225:    {},
	226:    {},
	227:    {},
	228:    {},
	229:    {},
	230:    {},
	231:    {},
	236:    {},
	237:    {},
	238:    {},
	239:    {},
	240:    {},
	241:    {},
	242:    {},
	243:    {},
	244:    {},
	245:    {},
	246:    {},
	247:    {},
	252:    {},
	253:    {},
	254:    {},
	255:    {},
	256:    {},
	257:    {},
	258:    {},
	259:    {},
	260:    {},
	261:    {},
	262:    {},
	263:    {},
	268:    {},
	269:    {},
	270:    {},
	271:    {},
	272:    {},
	273:    {},
	274:    {},
	275:    {},
	276:    {},
	277:    {},
	278:    {},
	279:    {},
	286:    {},
	287:    {},
	288:    {},
	289:    {},
	290:    {},
	291:    {},
	292:    {},
	293:    {},
	294:    {},
	296:    {},
	297:    {},
	321:    {},
	322:    {},
	323:    {},
	324:    {},
	467:    {},
	468:    {},
	469:    {},
	470:    {},
	471:    {},
	472:    {},
	473:    {},
	474:    {},
	476:    {},
	477:    {},
	478:    {},
	479:    {},
	480:    {},
	481:    {},
	482:    {},
	483:    {},
	484:    {},
	485:    {},
	486:    {},
	487:    {},
	492:    {},
	493:    {},
	494:    {},
	495:    {},
	543:    {},
	544:    {},
	545:    {},
	546:    {},
	547:    {},
	548:    {},
	549:    {},
	550:    {},
	551:    {},
	552:    {},
	553:    {},
	554:    {},
	555:    {},
	556:    {},
	557:    {},
	558:    {},
	559:    {},
	560:    {},
	561:    {},
	562:    {},
	563:    {},
	564:    {},
	565:    {},
	566:    {},
	567:    {},
	568:    {},
	569:    {},
	570:    {},
	571:    {},
	572:    {},
	573:    {},
	574:    {},
	575:    {},
	576:    {},
	577:    {},
	578:    {},
	579:    {},
	581:    {},
	582:    {},
	583:    {},
	584:    {},
	585:    {},
	586:    {},
	587:    {},
	588:    {},
	589:    {},
	590:    {},
	591:    {},
	592:    {},
	593:    {},
	594:    {},
	595:    {},
	596:    {},
	597:    {},
	598:    {},
	599:    {},
	600:    {},
	601:    {},
	610:    {},
	611:    {},
	612:    {},
	613:    {},
	1010:   {},
	1741:   {},
	1742:   {},
	1743:   {},
	1744:   {},
	1745:   {},
	1746:   {},
	1747:   {},
	1748:   {},
	1749:   {},
	1750:   {},
	1751:   {},
	1752:   {},
	1753:   {},
	1754:   {},
	1755:   {},
	1756:   {},
	1757:   {},
	1771:   {},
	1772:   {},
	1773:   {},
	1774:   {},
	1775:   {},
	1776:   {},
	1777:   {},
	1778:   {},
	1779:   {},
	1780:   {},
	1781:   {},
	1782:   {},
	1783:   {},
	1784:   {},
	1785:   {},
	1786:   {},
	1787:   {},
	1788:   {},
	1789:   {},
	1790:   {},
	1801:   {},
	1802:   {},
	1803:   {},
	1804:   {},
	1805:   {},
	1806:   {},
	1807:   {},
	1808:   {},
	1809:   {},
	1811:   {},
	1812:   {},
	1813:   {},
	1814:   {},
	1815:   {},
	1816:   {},
	1817:   {},
	1818:   {},
	1819:   {},
	1820:   {},
	1821:   {},
	1822:   {},
	1823:   {},
	1824:   {},
	1831:   {},
	1832:   {},
	1833:   {},
	1834:   {},
	1835:   {},
	1836:   {},
	1837:   {},
	1838:   {},
	1839:   {},
	1840:   {},
	1841:   {},
	1842:   {},
	1843:   {},
	1844:   {},
	1845:   {},
	1846:   {},
	1847:   {},
	1848:   {},
	1849:   {},
	1850:   {},
	1851:   {},
	1852:   {},
	1853:   {},
	1854:   {},
	1861:   {},
	1862:   {},
	1863:   {},
	1864:   {},
	1865:   {},
	1866:   {},
	1867:   {},
	1868:   {},
	1869:   {},
	1870:   {},
	1871:   {},
	1872:   {},
	1873:   {},
	1874:   {},
	1875:   {},
	1876:   {},
	1877:   {},
	1878:   {},
	1879:   {},
	1880:   {},
	1881:   {},
	1882:   {},
	1883:   {},
	1884:   {},
	1981:   {},
	1982:   {},
	1983:   {},
	1984:   {},
	1985:   {},
	1986:   {},
	1987:   {},
	1988:   {},
	1989:   {},
	1990:   {},
	1991:   {},
	1992:   {},
	1993:   {},
	1994:   {},
	1995:   {},
	1996:   {},
	1997:   {},
	1998:   {},
	1999:   {},
	2000:   {},
	2001:   {},
	2002:   {},
	2003:   {},
	2004:   {},
	2028:   {},
	2029:   {},
	2030:   {},
	2031:   {},
	2032:   {},
	2033:   {},
	2100:   {},
	2101:   {},
	2102:   {},
	2103:   {},
	2104:   {},
	2105:   {},
	0x453B: {},
	0x453C: {},
	0x453D: {},
	0x453E: {},
	0x453F: {},
	0x4540: {},
	0x4541: {},
	0x4542: {},
	0x4543: {},
	0x4544: {},
	0x4545: {},
	0x4546: {},
	0x4547: {},
	0x4548: {},
	0x4549: {},
	0x454A: {},
	0x454B: {},
	0x454C: {},
	0x454D: {},
	0x454E: {},
	0x454F: {},
}
