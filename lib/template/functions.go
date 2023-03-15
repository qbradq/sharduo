package template

import (
	"fmt"
	"log"
	"strconv"
	txtmp "text/template"

	"github.com/qbradq/sharduo/lib/uo"
)

// Global function map for templates
var templateFuncMap = txtmp.FuncMap{
	"EquipPlayer": equipPlayer,      // EquipPlayer generates an equipment list for a player mobile
	"New":         templateNew,      // New creates a new object from the named template, adds it to the world datastores, then returns the string representation of the object's serial
	"PartialHue":  partialHue,       // PartialHue sets the partial hue flag
	"RandomNew":   randomNew,        // RandomNew creates a new object of a template randomly selected from the named list
	"RandomBool":  randomBool,       // RandomBool returns a random boolean value
	"Random":      randomListMember, // Random returns a random string from the named list, or an empty string if the named list was not found
}

func randomBool() bool {
	return tm.rng.RandomBool()
}

func templateNew(name string) string {
	o := Create(name)
	if o == nil {
		return "0"
	}
	return o.Serial().String()
}

func randomListMember(list string) string {
	l, ok := tm.lists.Get(list)
	if !ok || len(l) == 0 {
		log.Printf("list %s not found\n", list)
		return ""
	}
	return l[tm.rng.Random(0, len(l)-1)]
}

func randomNew(list string) string {
	tn := randomListMember(list)
	if tn == "" {
		return "0"
	}
	return templateNew(tn)
}

func partialHue(hue string) string {
	v, err := strconv.ParseInt(hue, 0, 32)
	if err != nil {
		return hue
	}
	h := uo.Hue(v).SetPartialHue()
	return fmt.Sprintf("%d", h)
}

func equipPlayer(isFemale string) string {
	// Required equipment
	ret := templateNew("PlayerBankBox")
	ret += "," + templateNew("PlayerBackpack")
	// Hair
	if isFemale != "" {
		// 1% of female players are bald
		if tm.rng.Random(1, 100) != 100 {
			ret += "," + randomNew("FemaleHair")
		}
	} else {
		// 5% of male players are bald
		if tm.rng.Random(1, 20) != 20 {
			ret += "," + randomNew("MaleHair")
		}
	}
	// Basic clothing
	// TODO Add dresses for ladies
	ret += "," + randomNew("Shirt")
	ret += "," + randomNew("Pants")
	ret += "," + randomNew("Shoes")
	return ret
}
