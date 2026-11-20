package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
)

// Recipes and rates come from the game's own data files (ao-data/ao-bin-dumps, checked 2026-11-20):
// items.xml craftingrequirements, craftingmodifiers.xml, gamedata_asia.xml.

type resource struct {
	raw, refined, bonusCity string
}

var resources = map[string]resource{
	"ore":   {"ORE", "METALBAR", "Thetford"},
	"wood":  {"WOOD", "PLANKS", "Fort Sterling"},
	"fiber": {"FIBER", "CLOTH", "Lymhurst"},
	"hide":  {"HIDE", "LEATHER", "Martlock"},
	"rock":  {"ROCK", "STONEBLOCK", "Bridgewatch"},
}

// Refining markets on the Asia server (Black Market buys gear only).
var cities = []string{"Bridgewatch", "Brecilien", "Caerleon", "Fort Sterling", "Lymhurst", "Martlock", "Thetford"}

var rawPerCraft = map[int]int{2: 1, 3: 2, 4: 2, 5: 3, 6: 4, 7: 5, 8: 5}

const (
	baseBonus      = 0.18  // refiningbonus in every royal city, Caerleon, Brecilien
	specialtyBonus = 0.40  // the city's specialty resource
	focusBonus     = 0.59  // ActionFocus CraftingEfficiency
	setupFee       = 0.025 // sell order setup fee
	nutritionPerIV = 0.1125
	staleAfter     = 6 * time.Hour // older quotes are shown but kept out of the transport route
)

type Item struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Tier     int    `json:"tier"`
	Enchant  int    `json:"enchant"`
	Refined  bool   `json:"refined"`
	Count    int    `json:"count"`
}

type Recipe struct {
	Raw     Item  `json:"raw"`
	Lower   *Item `json:"lower"`
	Product Item  `json:"product"`
}

func itemID(code string, tier, enchant int) string {
	if enchant == 0 {
		return fmt.Sprintf("T%d_%s", tier, code)
	}
	return fmt.Sprintf("T%d_%s_LEVEL%d@%d", tier, code, enchant, enchant)
}

func newRecipe(res string, tier, enchant int) (Recipe, error) {
	r, ok := resources[res]
	switch {
	case !ok:
		return Recipe{}, errors.New("resource must be ore, wood, fiber, hide or rock")
	case tier < 2 || tier > 8:
		return Recipe{}, errors.New("tier must be 2-8")
	case enchant < 0 || enchant > 4:
		return Recipe{}, errors.New("enchant must be 0-4")
	case enchant > 0 && tier < 4:
		return Recipe{}, errors.New("enchanted refining starts at T4")
	case enchant > 0 && res == "rock":
		// Enchanted rock refines into batches of plain stone blocks; not modelled.
		return Recipe{}, errors.New("enchanted rock is not supported")
	}
	item := func(code string, t, e int, refined bool, n int) Item {
		return Item{ID: itemID(code, t, e), Resource: res, Tier: t, Enchant: e, Refined: refined, Count: n}
	}
	rec := Recipe{
		Raw:     item(r.raw, tier, enchant, false, rawPerCraft[tier]),
		Product: item(r.refined, tier, enchant, true, 1),
	}
	if tier > 2 {
		lowerEnchant := enchant
		if tier == 4 { // T4.x takes a plain T3 refined item; T5.x+ the same enchantment
			lowerEnchant = 0
		}
		low := item(r.refined, tier-1, lowerEnchant, true, 1)
		rec.Lower = &low
	}
	return rec, nil
}

// returnRate is the share of ingredients given back: 1 - 1/(1+bonus).
func returnRate(specialty, focus bool) float64 {
	b := baseBonus
	if specialty {
		b += specialtyBonus
	}
	if focus {
		b += focusBonus
	}
	return 1 - 1/(1+b)
}

func sellFee(premium bool) float64 {
	if premium {
		return 0.04 + setupFee
	}
	return 0.08 + setupFee
}

// stationFee is silver per craft: item value (2^(tier+enchant)) x 0.1125 nutrition x fee per 100 nutrition.
func stationFee(p Item, feePer100 int) float64 {
	iv := math.Pow(2, float64(p.Tier+p.Enchant))
	return iv * nutritionPerIV * float64(feePer100) / 100
}

type Row struct {
	City         string    `json:"city"`
	Bonus        bool      `json:"bonus"`
	RawPrice     int       `json:"rawPrice"`
	LowerPrice   int       `json:"lowerPrice"`
	ProductPrice int       `json:"productPrice"`
	RRR          float64   `json:"rrr"`
	RRRFocus     float64   `json:"rrrFocus"`
	Profit       float64   `json:"profit"`
	ProfitFocus  float64   `json:"profitFocus"`
	Margin       float64   `json:"margin"`
	MarginFocus  float64   `json:"marginFocus"`
	Updated      time.Time `json:"updated"`
	Missing      bool      `json:"missing"`
	AvgRaw       bool      `json:"avgRaw"`
	AvgLower     bool      `json:"avgLower"`
	AvgProduct   bool      `json:"avgProduct"`
}

type Route struct {
	Buy         string    `json:"buy"`
	Refine      string    `json:"refine"`
	Sell        string    `json:"sell"`
	Profit      float64   `json:"profit"`
	ProfitFocus float64   `json:"profitFocus"`
	Updated     time.Time `json:"updated"`
}

// oldest is the oldest quote among prices that exist; zero when none do.
func oldest(ps ...Price) time.Time {
	var t time.Time
	for _, p := range ps {
		if p.SellMin > 0 && (t.IsZero() || p.SellMinDate.Before(t)) {
			t = p.SellMinDate
		}
	}
	return t
}

type RefineResult struct {
	Recipe
	Premium   bool    `json:"premium"`
	FeePer100 int     `json:"feePer100"`
	SellFee   float64 `json:"sellFee"`
	BonusCity string  `json:"bonusCity"`
	Rows      []Row   `json:"rows"`
	Route     *Route  `json:"route"`
}

// profit per refined item, given ingredient cost, sale price and return rate.
func profit(cost, price, rrr, fee, sell float64) (float64, float64) {
	effective := cost*(1-rrr) + fee
	p := price*(1-sell) - effective
	return p, p / effective
}

// computeRefine is pure: prices[itemID][city].
func computeRefine(rec Recipe, prices map[string]map[string]Price, premium bool, feePer100 int, bonusCity string, now time.Time) RefineResult {
	sell := sellFee(premium)
	fee := stationFee(rec.Product, feePer100)
	res := RefineResult{Recipe: rec, Premium: premium, FeePer100: feePer100, SellFee: sell, BonusCity: bonusCity}

	// cost of one craft's ingredients in a city, the oldest ingredient quote, and whether all exist.
	cost := func(city string) (float64, time.Time, bool) {
		raw := prices[rec.Raw.ID][city]
		c, ok := float64(raw.SellMin*rec.Raw.Count), raw.SellMin > 0
		low := Price{}
		if rec.Lower != nil {
			low = prices[rec.Lower.ID][city]
			c += float64(low.SellMin)
			ok = ok && low.SellMin > 0
		}
		return c, oldest(raw, low), ok
	}

	for _, city := range cities {
		bonus := city == bonusCity
		row := Row{
			City: city, Bonus: bonus,
			RawPrice:     prices[rec.Raw.ID][city].SellMin,
			ProductPrice: prices[rec.Product.ID][city].SellMin,
			RRR:          returnRate(bonus, false),
			RRRFocus:     returnRate(bonus, true),
		}
		row.AvgRaw, row.AvgProduct = prices[rec.Raw.ID][city].Avg, prices[rec.Product.ID][city].Avg
		if rec.Lower != nil {
			row.LowerPrice = prices[rec.Lower.ID][city].SellMin
			row.AvgLower = prices[rec.Lower.ID][city].Avg
		}
		c, _, ok := cost(city)
		prod := prices[rec.Product.ID][city]
		row.Missing = !ok || prod.SellMin == 0
		row.Updated = oldest(prices[rec.Raw.ID][city], prices[rec.Product.ID][city])
		if rec.Lower != nil {
			row.Updated = oldest(prices[rec.Raw.ID][city], prices[rec.Lower.ID][city], prices[rec.Product.ID][city])
		}
		if !row.Missing {
			row.Profit, row.Margin = profit(c, float64(prod.SellMin), row.RRR, fee, sell)
			row.ProfitFocus, row.MarginFocus = profit(c, float64(prod.SellMin), row.RRRFocus, fee, sell)
		}
		res.Rows = append(res.Rows, row)
	}
	sort.SliceStable(res.Rows, func(i, j int) bool {
		a, b := res.Rows[i], res.Rows[j]
		if a.Missing != b.Missing {
			return !a.Missing
		}
		return a.Profit > b.Profit
	})

	// Transport route: buy ingredients where cheapest, refine in the specialty city, sell where dearest.
	// Only fresh live orders: a route built on stale or averaged prices would mislead.
	fresh := func(t time.Time) bool { return now.Sub(t) < staleAfter }
	buy, sellCity := "", ""
	var minCost, maxPrice float64
	var buyAt, sellAt time.Time
	for _, city := range cities {
		live := !prices[rec.Raw.ID][city].Avg && (rec.Lower == nil || !prices[rec.Lower.ID][city].Avg)
		if c, at, ok := cost(city); ok && live && fresh(at) && (buy == "" || c < minCost) {
			buy, minCost, buyAt = city, c, at
		}
		if p := prices[rec.Product.ID][city]; p.SellMin > 0 && !p.Avg && fresh(p.SellMinDate) && float64(p.SellMin) > maxPrice {
			sellCity, maxPrice, sellAt = city, float64(p.SellMin), p.SellMinDate
		}
	}
	if buy != "" && sellCity != "" {
		p, _ := profit(minCost, maxPrice, returnRate(true, false), fee, sell)
		pf, _ := profit(minCost, maxPrice, returnRate(true, true), fee, sell)
		res.Route = &Route{Buy: buy, Refine: bonusCity, Sell: sellCity, Profit: p, ProfitFocus: pf, Updated: oldest(Price{SellMin: 1, SellMinDate: buyAt}, Price{SellMin: 1, SellMinDate: sellAt})}
	}
	return res
}
