package main

import (
	"errors"
	"math"
	"slices"
	"sort"
	"time"
)

// Item weight per tier from items.xml (ao-bin-dumps 2026-11-26). Raw and refined resources
// weigh the same, and enchantment adds nothing.
var tierWeight = map[int]float64{4: 0.51, 5: 0.76, 6: 1.14, 7: 1.71, 8: 2.56}

var tradeTiers = []int{4, 5, 6, 7, 8}

type TradeOption struct {
	Item     Item      `json:"item"`
	Buy      int       `json:"buy"`  // cheapest sell order in the start city
	Sell     int       `json:"sell"` // cheapest sell order in the destination, which we match
	AvgBuy   bool      `json:"avgBuy"`
	AvgSell  bool      `json:"avgSell"`
	Unit     float64   `json:"unit"` // profit per item after the selling fees
	Units    int       `json:"units"`
	ByBudget bool      `json:"byBudget"` // the budget, not the weight, limits the load
	Trip     float64   `json:"trip"`
	Need     int       `json:"need"` // silver to buy the load
	Updated  time.Time `json:"updated"`
}

type TradeDest struct {
	City  string         `json:"city"`
	Best  *TradeOption   `json:"best"`
	Tiers []*TradeOption `json:"tiers"` // T4..T8; nil where no item of that tier has both prices
}

type TradeResult struct {
	From         string      `json:"from"`
	Enchant      int         `json:"enchant"`
	Kg           float64     `json:"kg"`
	Budget       int         `json:"budget"`
	Premium      bool        `json:"premium"`
	SellFee      float64     `json:"sellFee"`
	Destinations []TradeDest `json:"destinations"`
}

// tradeItems are the raw and refined resources of one enchantment, T4-T8.
func tradeItems(enchant int) ([]Item, error) {
	if enchant < 0 || enchant > 4 {
		return nil, errors.New("enchant must be 0-4")
	}
	var out []Item
	for _, res := range []string{"ore", "wood", "fiber", "hide", "rock"} {
		if enchant > 0 && res == "rock" { // enchanted stone has no refined form; same rule as the calculator
			continue
		}
		r := resources[res]
		for _, t := range tradeTiers {
			out = append(out,
				Item{ID: itemID(r.raw, t, enchant), Resource: res, Tier: t, Enchant: enchant, Count: 1},
				Item{ID: itemID(r.refined, t, enchant), Resource: res, Tier: t, Enchant: enchant, Refined: true, Count: 1})
		}
	}
	return out, nil
}

func better(a, b *TradeOption) bool {
	return b == nil || a.Trip > b.Trip || (a.Trip == b.Trip && a.Unit > b.Unit)
}

// computeTrade is pure: buy in `from`, carry kg (and at most budget silver, 0 = no limit), sell elsewhere.
func computeTrade(from string, items []Item, prices map[string]map[string]Price, kg float64, budget int, premium bool) TradeResult {
	fee := sellFee(premium)
	res := TradeResult{From: from, Kg: kg, Budget: budget, Premium: premium, SellFee: fee}
	if len(items) > 0 {
		res.Enchant = items[0].Enchant
	}
	for _, city := range cities {
		if city == from {
			continue
		}
		d := TradeDest{City: city, Tiers: make([]*TradeOption, len(tradeTiers))}
		for _, it := range items {
			b, s := prices[it.ID][from], prices[it.ID][city]
			if b.SellMin <= 0 || s.SellMin <= 0 {
				continue
			}
			units := int(math.Floor(kg/tierWeight[it.Tier] + 1e-9))
			byBudget := budget > 0 && budget/b.SellMin < units
			if byBudget {
				units = budget / b.SellMin
			}
			unit := float64(s.SellMin)*(1-fee) - float64(b.SellMin)
			o := &TradeOption{
				Item: it, Buy: b.SellMin, Sell: s.SellMin, AvgBuy: b.Avg, AvgSell: s.Avg,
				Unit: unit, Units: units, ByBudget: byBudget, Trip: unit * float64(units), Need: units * b.SellMin,
				Updated: oldest(b, s),
			}
			if i := slices.Index(tradeTiers, it.Tier); better(o, d.Tiers[i]) {
				d.Tiers[i] = o
			}
			if better(o, d.Best) {
				d.Best = o
			}
		}
		res.Destinations = append(res.Destinations, d)
	}
	sort.SliceStable(res.Destinations, func(i, j int) bool {
		a, b := res.Destinations[i].Best, res.Destinations[j].Best
		if (a == nil) != (b == nil) {
			return a != nil
		}
		return a != nil && a.Trip > b.Trip
	})
	return res
}
