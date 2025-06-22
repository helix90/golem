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
} 