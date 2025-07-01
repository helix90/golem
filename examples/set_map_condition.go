package main

import (
	"fmt"
	"golem/engine"
)

func main() {
	// Setup bot with sets and maps
	bot := engine.NewBot(false)
	bot.Sets["colors"] = map[string]struct{}{ "red": {}, "blue": {} }
	bot.Maps["countries"] = map[string]string{ "fr": "France", "de": "Germany" }

	sess := &engine.Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
	eval := engine.NewEvaluatorWithConfig(sess, nil, nil, "", bot)

	// Add to set
	eval.EvaluateTemplate(`<set name="colors">green</set>`)
	fmt.Println("Colors set after adding 'green':", bot.Sets["colors"])

	// Map lookup
	out, _ := eval.EvaluateTemplate(`<map name="countries">fr</map>`)
	fmt.Println("Map lookup for 'fr':", out)

	// Condition with set
	out, _ = eval.EvaluateTemplate(`<condition set="colors" value="blue">Blue is in the set.</condition>`)
	fmt.Println("Condition (set contains 'blue'):", out)

	// Condition with map
	out, _ = eval.EvaluateTemplate(`<condition map="countries" key="de">Hallo!</condition>`)
	fmt.Println("Condition (map has 'de'):", out)

	// Condition with <li> children
	tmpl := `<condition>
	<li set="colors" value="red">Red found</li>
	<li map="countries" key="fr">Bonjour</li>
	<li>Default</li>
</condition>`
	out, _ = eval.EvaluateTemplate(tmpl)
	fmt.Println("Condition with <li> children:", out)

	// Create a bot and define a set of pets
	bot.Sets["PETS"] = map[string]struct{}{
		"DOG":    {},
		"CAT":    {},
		"PARROT": {},
	}

	// Example 1: Check if "CAT" is a pet (with default response)
	tmpl1 := `<condition>
	<li set="PETS" value="CAT">Yes, that's a pet!</li>
	<li>Sorry, that's not a pet.</li>
</condition>`
	out1, _ := eval.EvaluateTemplate(tmpl1)
	fmt.Printf("Is CAT a pet? %s\n", out1)

	// Example 2: Check if "HORSE" is a pet (with default response)
	tmpl2 := `<condition>
	<li set="PETS" value="HORSE">Yes, that's a pet!</li>
	<li>Sorry, that's not a pet.</li>
</condition>`
	out2, _ := eval.EvaluateTemplate(tmpl2)
	fmt.Printf("Is HORSE a pet? %s\n", out2)

	// Example 3: Set and get favorite pet
	_, _ = eval.EvaluateTemplate(`<set name="FAV_PET">DOG</set>`)
	favPet, _ := eval.EvaluateTemplate(`<get name="FAV_PET"/>`)
	fmt.Printf("Favorite pet: %s\n", favPet)

	// Example 4: Talk about favorite pet
	tmpl4 := `My favorite pet is <get name="FAV_PET"/>!`
	out4, _ := eval.EvaluateTemplate(tmpl4)
	fmt.Println(out4)
} 