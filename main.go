package main

import (
	"fmt"
	"log"

	"github.com/goml/gobrain"

	"github.com/jBugman/pricecheck/poe"
)

const (
	iterations   = 2501
	learningRate = 0.7
	mFactor      = 0.4
	hiddenLayerK = 0.3
)

var (
	properties = poe.Properties()
	size       = len(properties)
)

func preparePatterns(items []poe.Item) ([][][]float64, float64) {
	patterns := make([][][]float64, len(items))
	var maxPrice float64
	for _, item := range items {
		if item.Price > maxPrice {
			maxPrice = item.Price
		}
	}
	for i, item := range items {
		patterns[i] = [][]float64{
			preparePattern(item),
			{item.Price / maxPrice},
		}
	}
	return patterns, maxPrice
}

func preparePattern(item poe.Item) []float64 {
	p := make([]float64, size)
	for i, prop := range properties {
		val, ok := item.Params[prop]
		if ok {
			p[i] = float64(val)
		}
	}
	return p
}

func main() {

	fmt.Println("== Items ==")
	//	items, err := poe.LoadFromFile("data.txt")
	items, err := poe.LoadFromJsonFile("jewels.json")
	if err != nil {
		log.Fatal(err)
	}
	items2, err := poe.LoadFromJsonFile("jewels2.json")
	if err != nil {
		log.Fatal(err)
	}
	items = append(items, items2...)
	fmt.Println("Items loaded:", len(items))
	//	for _, it := range items {
	//		fmt.Println(it)
	//	}

	fmt.Println("== Feature vectors ==")
	patterns, maxPrice := preparePatterns(items)
	//	for _, p := range patterns {
	//		fmt.Println(p)
	//	}

	fmt.Println("== Training ==")
	nn := gobrain.FeedForward{}
	nn.Init(size, int(hiddenLayerK*float64(size)), 1)
	nn.Train(patterns, iterations, learningRate, mFactor, true)

	nn.Test(patterns)

	fmt.Println("== Test ==")
	testFunc := func(text, price string) {
		item, err := poe.ParseItem(text)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(item)
		inputs := preparePattern(item)
		fmt.Println(price, "->", nn.Update(inputs)[0]*maxPrice)
	}

	testFunc(`Gale Edge Crimson Jewel
13% increased Physical Weapon Damage while Dual Wielding
8% increased Attack Speed with Swords
3% reduced Mana Cost of Skills`, "5 chaos")
	testFunc(`Blight Glimmer Crimson Jewel
15% increased Weapon Critical Strike Chance while Dual Wielding
5% increased maximum Life
14% increased Freeze Duration on Enemies
2% chance to Freeze`, "20 chaos")
	testFunc(`Spirit Splinter Crimson Jewel
15% increased Fire Damage
10% increased Critical Strike Multiplier with One Handed Melee Weapons
4% increased Attack and Cast Speed`, "2 chaos")
	testFunc(`Chimeric Flame Crimson Jewel
6% increased Attack Speed while holding a Shield
0.2% of Physical Attack Damage Leeched as Life
11% increased Area Damage
Totems gain +6% to all Elemental Resistances`, "3 chaos")
	testFunc(`Ghoul Ruin Viridian Jewel
14% increased Physical Damage with Claws
13% increased Chaos Damage
5% increased Attack Speed`, "10 chaos")
}
