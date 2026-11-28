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
	// Refine is set when the load is refined in the start city before hauling; Buy is then the
	// ingredient silver per refined item after the return rate.
	Refine *TradeRefine `json:"refine,omitempty"`
}

type TradeRefine struct {
	Raw        Item    `json:"raw"` // Count is raw per craft
	RawPrice   int     `json:"rawPrice"`
	Lower      *Item   `json:"lower"`
	LowerPrice int     `json:"lowerPrice"`
	RRR        float64 `json:"rrr"`
}

type TradeDest struct {
	City  string         `json:"city"`
	Road  *Road          `json:"road"` // nil when the zone map is not loaded
	Best  *TradeOption   `json:"best"`
	Tiers []*TradeOption `json:"tiers"` // T4..T8; nil where no item of that tier has both prices
}

type TradeResult struct {
	From         string      `json:"from"`
	Enchant      int         `json:"enchant"`
	Kg           float64     `json:"kg"`
	Budget       int         `json:"budget"`
	Premium      bool        `json:"premium"`
	Focus        bool        `json:"focus"`
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

// tradePriceIDs adds the refining ingredients (raw and the tier-below refined item) to the items.
func tradePriceIDs(items []Item) []string {
	seen := map[string]bool{}
	var ids []string
	add := func(id string) {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for _, it := range items {
		add(it.ID)
		if rec, err := newRecipe(it.Resource, it.Tier, it.Enchant); it.Refined && err == nil {
			add(rec.Raw.ID)
			if rec.Lower != nil {
				add(rec.Lower.ID)
			}
		}
	}
	return ids
}

// computeTrade is pure: buy in `from` (or buy the ingredients there and refine), carry kg
// (and at most budget silver, 0 = no limit), sell in every other city.
func computeTrade(from string, items []Item, prices map[string]map[string]Price, kg float64, budget int, premium, focus bool, roads Roads) TradeResult {
	fee := sellFee(premium)
	res := TradeResult{From: from, Kg: kg, Budget: budget, Premium: premium, Focus: focus, SellFee: fee}
	if len(items) > 0 {
		res.Enchant = items[0].Enchant
	}
	// option prices one load of `it` that costs `cost` silver per item in the start city.
	option := func(it Item, cost float64, s Price, avgBuy bool, updated time.Time, ref *TradeRefine) *TradeOption {
		units := int(math.Floor(kg/tierWeight[it.Tier] + 1e-9))
		byBudget := budget > 0 && float64(budget) < cost*float64(units)
		if byBudget {
			units = int(float64(budget) / cost)
		}
		unit := float64(s.SellMin)*(1-fee) - cost
		return &TradeOption{
			Item: it, Buy: int(math.Round(cost)), Sell: s.SellMin, AvgBuy: avgBuy, AvgSell: s.Avg,
			Unit: unit, Units: units, ByBudget: byBudget, Trip: unit * float64(units), Need: int(math.Round(cost * float64(units))),
			Updated: updated, Refine: ref,
		}
	}
	for _, city := range cities {
		if city == from {
			continue
		}
		d := TradeDest{City: city, Tiers: make([]*TradeOption, len(tradeTiers))}
		if r, ok := roads[from][city]; ok {
			d.Road = &r
		}
		consider := func(o *TradeOption) {
			if i := slices.Index(tradeTiers, o.Item.Tier); better(o, d.Tiers[i]) {
				d.Tiers[i] = o
			}
			if better(o, d.Best) {
				d.Best = o
			}
		}
		for _, it := range items {
			s := prices[it.ID][city]
			if s.SellMin <= 0 {
				continue
			}
			if b := prices[it.ID][from]; b.SellMin > 0 {
				consider(option(it, float64(b.SellMin), s, b.Avg, oldest(b, s), nil))
			}
			if !it.Refined {
				continue
			}
			// Refine in the start city: same recipe and return rate as the refining calculator; station fee left out.
			rec, err := newRecipe(it.Resource, it.Tier, it.Enchant)
			raw := prices[rec.Raw.ID][from]
			if err != nil || raw.SellMin <= 0 {
				continue
			}
			cost, low := float64(raw.SellMin*rec.Raw.Count), Price{}
			if rec.Lower != nil {
				if low = prices[rec.Lower.ID][from]; low.SellMin <= 0 {
					continue
				}
				cost += float64(low.SellMin)
			}
			rrr := returnRate(from == resources[it.Resource].bonusCity, focus)
			ref := &TradeRefine{Raw: rec.Raw, RawPrice: raw.SellMin, Lower: rec.Lower, LowerPrice: low.SellMin, RRR: rrr}
			consider(option(it, cost*(1-rrr), s, raw.Avg || low.Avg, oldest(raw, low, s), ref))
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
