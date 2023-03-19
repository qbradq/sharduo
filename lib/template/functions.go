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
	"DressHuman": dressHuman,       // Creates clothing and hair for a human
	"New":        templateNew,      // New creates a new object from the named template, adds it to the world datastores, then returns the string representation of the object's serial
	"PartialHue": partialHue,       // Sets the partial hue flag
	"RandomNew":  randomNew,        // RandomNew creates a new object of a template randomly selected from the named list
	"RandomBool": randomBool,       // RandomBool returns a random boolean value
	"Random":     randomListMember, // Random returns a random string from the named list, or an empty string if the named list was not found
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

func randomListMember(name string) string {
	l, ok := tm.lists.Get(name)
	if !ok || len(l) == 0 {
		log.Printf("list %s not found\n", name)
		return ""
	}
	return l[tm.rng.Random(0, len(l)-1)]
}

func randomNew(name string) string {
	tn := randomListMember(name)
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

func dressHuman() string {
	ret := randomNew("Shoes")
	if templateContext["IsFemale"] != "" {
		if tm.rng.Random(1, 20) > 17 {
			// 15% of women in Britannia wear pants
			ret += "," + randomNew("Shirt") +
				"," + randomNew("Pants") +
				"," + randomNew("FemaleHair")
		} else {
			ret += "," + randomNew("Dress") +
				"," + randomNew("FemaleHair")
		}
	} else {
		ret += "," + randomNew("Shirt") +
			"," + randomNew("Pants") +
			"," + randomNew("MaleHair")
	}
	return ret
}
