package poe

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type ElasticItem struct {
	Info struct {
		FullName string `json:"fullName"`
		Name     string `json:"name"`
		TypeLine string `json:"typeLine"`
	} `json:"info"`
	Mods  map[string]interface{} `json:"modsTotal"`
	Price struct {
		ChaosPrice    float64 `json:"chaosEquiv"`
		OriginalPrice float64 `json:"amount"`
		Currency      string  `json:"currency"`
	} `json:"shop"`
}

func LoadFromJsonFile(filename string) ([]Item, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var source struct {
		Hits struct {
			Total int
			Hits  []struct {
				Source ElasticItem `json:"_source"`
			}
		}
	}
	err = json.Unmarshal(data, &source)
	if err != nil {
		return nil, err
	}
	items := make([]Item, len(source.Hits.Hits))
	props := make(map[string]bool)

	// internal mod validation, remove after full check
	for _, p := range Properties() {
		props[p] = true
	}

Items:
	for i, src := range source.Hits.Hits {

		// internal mod validation, remove after full check
		mods := make(map[string]float64)
		for k, v := range src.Source.Mods {
			if _, ok := props[k]; ok {
				switch t := v.(type) {
				case float64:
					mods[k] = float64(t)
				default:
					log.Println("[!] Unsupported property type:", k, v)
				}
			} else {
				log.Println("[!] Unknown property:", k)
				continue Items
			}
		}
		items[i] = Item{
			Name:   src.Source.Info.FullName,
			Params: mods,
			Price:  src.Source.Price.ChaosPrice,
		}
	}
	return items, nil
}
