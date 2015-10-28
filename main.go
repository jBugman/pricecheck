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

func testNN(filename string, nn gobrain.FeedForward, maxPrice float64) {
	items, err := poe.LoadFromFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		fmt.Println(item)
		inputs := preparePattern(item)
		fmt.Println(item.Price, "->", nn.Update(inputs)[0]*maxPrice)
	}
}

func loadItems() []poe.Item {
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
	return items
}

func main() {

	items := loadItems()
	patterns, maxPrice := preparePatterns(items)

	fmt.Println("== Training ==")
	nn := gobrain.FeedForward{}
	nn.Init(size, int(hiddenLayerK*float64(size)), 1)
	nn.Train(patterns, iterations, learningRate, mFactor, true)

	nn.Test(patterns)

	fmt.Println("== Test ==")
	testNN("test_jewels.txt", nn, maxPrice)
}
