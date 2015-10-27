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

var propertyRegexes = [][]string{
	{"Life", "(\\d*?)% increased maximum Life"},
	{"Mana", "(\\d*?)% increased maximum Mana"},
	{"Reduced Mana Cost", "(\\d*?)% reduced Mana Cost of Skills"},
	{"Rarity of Items found", "(\\d*?)% increased Rarity of Items found"},

	{"Energy Shield gained on hit", "\\+(\\d*?) Energy Shield gained for each Enemy hit by your Attacks"},
	{"Physical Attack Damage Leeched as Life", "0\\.(\\d*?)% of Physical Attack Damage Leeched as Life"},

	{"Chance to Block Spells while Dual Wielding", "(\\d*?)% additional Chance to Block Spells while Dual Wielding"},
	{"Block Chance while Dual Wielding", "(\\d*?)% additional Block Chance while Dual Wielding"},

	{"Chance to Knock Back", "(\\d*?)% chance to Knock Enemies Back on hit"},
	{"Stun Duration", "(\\d*?)% increased Stun Duration on Enemies"},

	{"Area Damage", "(\\d*?)% increased Area Damage"},
	{"Damage over Time", "(\\d*?)% increased Damage over Time"},
	{"Projectile Speed", "(\\d*?)% increased Projectile Speed"},
	{"Attack and Cast Speed", "(\\d*?)% increased Attack and Cast Speed"},

	{"Ignite Duration", "(\\d*?)% increased Ignite Duration on Enemies"},
	{"Shock Duration", "(\\d*?)% increased Shock Duration on Enemies"},
	{"Freeze Duration", "(\\d*?)% increased Freeze Duration on Enemies"},
	{"Chance to Ignite", "(\\d*?)% chance to Ignite"},
	{"Chance to Shock", "(\\d*?)% chance to Shock"},
	{"Chance to Freeze", "(\\d*?)% chance to Freeze"},

	{"Dexterity", "\\+(\\d*?) to Dexterity"},
	{"Intelligence", "\\+(\\d*?) to Intelligence"},
	{"Strength", "\\+(\\d*?) to Strength"},
	{"Strength and Intelligence", "\\+(\\d*?) to Strength and Intelligence"},

	{"Minion Life", "(\\d*?)% increased Minion Life"},
	{"Minion Damage", "Minions deal (\\d*?)% increased Damage"},
	{"Minion Chance to Block", "Minions have (\\d*?)% Chance to Block"},
	{"Minion Elemental Resistances", "Minions have \\+(\\d*?)% to all Elemental Resistances"},

	{"Totems Life", "(\\d*?)% increased Totem Life"},
	{"Totems Damage", "Totems deal (\\d*?)% increased Damage"},
	{"Totems Elemental Resistances", "Totems gain \\+(\\d*?)% to all Elemental Resistances"},

	{"Mine Damage", "(\\d*?)% increased Mine Damage"},

	{"Fire Resistance", "\\+(\\d*?)% to Fire Resistance"},
	{"Cold Resistance", "\\+(\\d*?)% to Cold Resistance"},
	{"Lightning Resistance", "\\+(\\d*?)% to Lightning Resistance"},
	{"Cold and Lightning Resistances", "\\+(\\d*?)% to Cold and Lightning Resistances"},
	{"Fire and Lightning Resistances", "\\+(\\d*?)% to Fire and Lightning Resistances"},
	{"Fire and Cold Resistances", "\\+(\\d*?)% to Fire and Cold Resistances"},
	{"Chaos Resistance", "\\+(\\d*?)% to Chaos Resistance"},

	{"Attack Speed", "(\\d*?)% increased Attack Speed"},
	{"Attack Speed with Wands", "(\\d*?)% increased Attack Speed with Wands"},
	{"Attack Speed with Swords", "(\\d*?)% increased Attack Speed with Swords"},
	{"Attack Speed with Axes", "(\\d*?)% increased Attack Speed with Axes"},
	{"Attack Speed while holding a Shield", "(\\d*?)% increased Attack Speed while holding a Shield"},
	{"Attack Speed while Dual Wielding", "(\\d*?)% increased Attack Speed while Dual Wielding"},

	{"Cast Speed", "(\\d*?)% increased Cast Speed"},
	{"Cast Speed with Cold Skills", "(\\d*?)% increased Cast Speed with Cold Skills"},
	{"Cast Speed with Fire Skills", "(\\d*?)% increased Cast Speed with Fire Skills"},
	{"Cast Speed with Lightning Skills", "(\\d*?)% increased Cast Speed with Lightning Skills"},

	{"Physical Damage with Daggers", "(\\d*?)% increased Physical Damage with Daggers"},
	{"Physical Damage with Axes", "(\\d*?)% increased Physical Damage with Axes"},
	{"Physical Weapon Damage while Dual Wielding", "(\\d*?)% increased Physical Weapon Damage while Dual Wielding"},

	{"Damage", "(\\d*?)% increased Damage"},
	{"Cold Damage", "(\\d*?)% increased Cold Damage"},
	{"Fire Damage", "(\\d*?)% increased Fire Damage"},
	{"Lightning Damage", "(\\d*?)% increased Lightning Damage"},
	{"Physical Damage", "(\\d*?)% increased Physical Damage"},
	{"Melee Damage", "(\\d*?)% increased Melee Damage"},
	{"Projectile Damage", "(\\d*?)% increased Projectile Damage"},

	{"Spell Damage while Dual Wielding", "(\\d*?)% increased Spell Damage while Dual Wielding"},
	{"Spell Damage while holding a Shield", "(\\d*?)% increased Spell Damage while holding a Shield"},

	{"Global Critical Strike Chance", "(\\d*?)% increased Global Critical Strike Chance"},
	{"Weapon Critical Strike Chance while Dual Wielding", "(\\d*?)% increased Weapon Critical Strike Chance while Dual Wielding"},
	{"Critical Strike Chance for Spells", "(\\d*?)% increased Critical Strike Chance for Spells"},
	{"Critical Strike Chance with Cold Skills", "(\\d*?)% increased Critical Strike Chance with Cold Skills"},
	{"Critical Strike Chance with Two Handed Melee Weapons", "(\\d*?)% increased Critical Strike Chance with Two Handed Melee Weapons"},

	{"Global Critical Strike Multiplier", "(\\d*?)% increased Global Critical Strike Multiplier"},
	{"Critical Strike Multiplier with Elemental Skills", "(\\d*?)% increased Critical Strike Multiplier with Elemental Skills"},
	{"Critical Strike Multiplier with Lightning Skills", "(\\d*?)% increased Critical Strike Multiplier with Lightning Skills"},
	{"Critical Strike Multiplier with Fire Skills", "(\\d*?)% increased Critical Strike Multiplier with Fire Skills"},
	{"Critical Strike Multiplier with Cold Skills", "(\\d*?)% increased Critical Strike Multiplier with Cold Skills"},
	{"Critical Strike Multiplier for Spells", "(\\d*?)% increased Critical Strike Multiplier for Spells"},
	{"Critical Strike Multiplier with One Handed Melee Weapons", "(\\d*?)% increased Critical Strike Multiplier with One Handed Melee Weapons"},
	{"Critical Strike Multiplier with Two Handed Melee Weapons", "(\\d*?)% increased Critical Strike Multiplier with Two Handed Melee Weapons"},
}

// Properties returns list of all known properties
func Properties() []string {
	list := make([]string, len(propertyRegexes))
	for i := range propertyRegexes {
		list[i] = propertyRegexes[i][0]
	}
	return list
}

// Item is basic PoE item model
type Item struct {
	Name   string
	Params map[string]int // Maybe should make properties float (i.e. leech)
	Price  float64
}

type parser struct {
	regexes     map[string]*regexp.Regexp
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
		regexes: make(map[string]*regexp.Regexp),
	}
	for _, r := range propertyRegexes {
		p.regexes[r[0]] = regexp.MustCompile(r[1])
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
		Params: make(map[string]int),
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
				val, err := strconv.Atoi(matches[1])
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
