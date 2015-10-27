package poe

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var currencyRatio = map[string]float64{
	"chaos":  1,
	"exalt":  50,
	"fuse":   0.5,
	"regret": 1.5,
	"jew":    0.125,
	"alt":    0.0667,
	"alch":   0.333,
}

// Properties returns list of all known properties
func Properties() []string {
	return instance().properties
}

// Item is basic PoE item model
type Item struct {
	Name   string
	Params map[string]float64
	Price  float64
}

type parser struct {
	regexes     map[string]*regexp.Regexp
	properties  []string
	initialized bool
}

var parserInstance parser

func instance() parser {
	if !parserInstance.initialized {
		parserInstance = newParser()
	}
	return parserInstance
}

func newParser() parser {
	p := parser{
		regexes:    make(map[string]*regexp.Regexp),
		properties: make([]string, len(propertyRegexes)),
	}
	for i, r := range propertyRegexes {
		p.regexes[r[0]] = regexp.MustCompile(r[1])
		p.properties[i] = r[0]
	}
	return p
}

// LoadFromFile parses multiple items from text file
func LoadFromFile(filename string) ([]Item, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	blocks := strings.Split(string(data), "\n\n")
	items, err := instance().parseItems(blocks)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (p parser) parsePrice(line string) (float64, error) {
	parts := strings.Fields(line)
	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}
	k, ok := currencyRatio[parts[2]]
	if !ok {
		return 0, errors.New("Undefined currency: " + parts[2])
	}
	return value * k, nil
}

func (p parser) parseItems(blocks []string) ([]Item, error) {
	var items []Item
	for _, block := range blocks {
		item, err := p.parseItem(block, false)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// ParseItem parses PoE textual item representation to Item model
func ParseItem(text string) (Item, error) {
	return instance().parseItem(text, true)
}

func (p parser) parseItem(text string, strict bool) (Item, error) {
	lines := strings.Split(text, "\n")
	item := Item{
		Name:   lines[0],
		Params: make(map[string]float64),
	}
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "$ ") {
			price, err := p.parsePrice(line)
			if err != nil {
				return Item{}, err
			}
			item.Price = price
			continue
		}

		var matched bool
		for k, r := range p.regexes {
			if matches := r.FindStringSubmatch(line); matches != nil {
				val, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					return Item{}, err
				}
				item.Params[k] = val
				matched = true
			}
		}
		if !matched {
			if strict {
				return Item{}, errors.New("Unknown property: " + line)
			}
			fmt.Println("[!] Unknown property:", line)
		}
	}
	return item, nil
}
