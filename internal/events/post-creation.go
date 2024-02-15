package events

import (
	"log"
	"strconv"

	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/uo"
	"github.com/qbradq/sharduo/lib/util"
)

func init() {
	reg("ConfigureHuman", configureHuman)
	reg("DressHuman", dressHuman)
	reg("EquipVendor", equipVendor)
	reg("PartialHue", partialHue)
	reg("RandomHue", randomHue)
	reg("RandomPartialHue", randomPartialHue)
}

func randomHue(r, s, v any) bool {
	ln := v.(string)
	hs := game.ListMember(ln).String()
	iv, err := strconv.ParseInt(hs, 0, 32)
	if err != nil {
		panic(err)
	}
	hue := uo.Hue(iv)
	if m, ok := r.(*game.Mobile); ok {
		m.Hue = hue
	} else {
		r.(*game.Item).Hue = hue
	}
	return true
}

func randomPartialHue(r, s, v any) bool {
	ln := v.(string)
	hs := game.ListMember(ln).String()
	iv, err := strconv.ParseInt(hs, 0, 32)
	if err != nil {
		panic(err)
	}
	hue := uo.Hue(iv).SetPartial()
	if m, ok := r.(*game.Mobile); ok {
		m.Hue = hue
	} else {
		r.(*game.Item).Hue = hue
	}
	return true
}

func partialHue(r, s, v any) bool {
	iv, err := strconv.ParseInt(v.(string), 0, 32)
	if err != nil {
		panic(err)
	}
	hue := uo.Hue(iv).SetPartial()
	if m, ok := r.(*game.Mobile); ok {
		m.Hue = hue
	} else {
		r.(*game.Item).Hue = hue
	}
	return true
}

func configureHuman(r, s, v any) bool {
	rm := r.(*game.Mobile)
	rm.Female = util.RandomBool()
	if rm.Body != uo.BodyCounselor && rm.Female {
		rm.Name = game.ListMember("FemaleName").String()
		rm.Body = uo.BodyHumanFemale
	} else if rm.Body != uo.BodyCounselor && !rm.Female {
		rm.Name = game.ListMember("MaleName").String()
		rm.Body = uo.BodyHumanMale
	}
	rm.Hue = uo.Hue(game.ListMember("SkinHue").Int())
	return true
}

func dressHuman(r, s, v any) bool {
	rm := r.(*game.Mobile)
	if rm.Body == uo.BodyCounselor {
		return true
	}
	if rm.Female {
		if util.Random(1, 20) > 17 {
			// 15% of women in Britannia wear pants
			rm.Equip(game.NewItem(string(game.ListMember("Shirt"))))
			rm.Equip(game.NewItem(string(game.ListMember("Pants"))))
		} else {
			rm.Equip(game.NewItem(string(game.ListMember("Dress"))))
		}
		rm.Equip(game.NewItem(string(game.ListMember("FemaleHair"))))
	} else {
		if util.Random(1, 20) == 0 {
			// 5% of men in Britannia wear dresses
			rm.Equip(game.NewItem(string(game.ListMember("Dress"))))
		} else {
			rm.Equip(game.NewItem(string(game.ListMember("Shirt"))))
			rm.Equip(game.NewItem(string(game.ListMember("Pants"))))
		}
		rm.Equip(game.NewItem(string(game.ListMember("MaleHair"))))
	}
	rm.Equip(game.NewItem(string(game.ListMember("Shoes"))))
	return true
}

func equipVendor(r, s, v any) bool {
	rm := r.(*game.Mobile)
	bp := rm.Equipment[uo.LayerNPCBuyRestockContainer]
	for _, tn := range game.TemplateLists[v.(string)] {
		i := game.NewItem(tn.String())
		if i == nil {
			log.Printf("warning: vendor list %s references non-existent item template %s",
				v.(string), tn.String())
			continue
		}
		bp.DropInto(i, true)
	}
	return true
}
