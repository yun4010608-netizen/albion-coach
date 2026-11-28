package main

import (
	"testing"
	"time"
)

func TestTradeItems(t *testing.T) {
	plain, _ := tradeItems(0)
	if len(plain) != 5*2*5 {
		t.Fatalf("plain items = %d, want 50", len(plain))
	}
	ench, _ := tradeItems(2)
	if len(ench) != 4*2*5 { // no enchanted stone
		t.Fatalf("enchanted items = %d, want 40", len(ench))
	}
	if ench[0].ID != "T4_ORE_LEVEL2@2" || !ench[1].Refined || ench[1].ID != "T4_METALBAR_LEVEL2@2" {
		t.Errorf("first enchanted items = %+v %+v", ench[0], ench[1])
	}
	if _, err := tradeItems(5); err == nil {
		t.Error("enchant 5 should fail")
	}
}

func TestComputeTrade(t *testing.T) {
	now := time.Now()
	p := func(v int) Price { return Price{SellMin: v, SellMinDate: now} }
	t4 := Item{ID: "T4_PLANKS", Resource: "wood", Tier: 4, Refined: true}
	t6 := Item{ID: "T6_PLANKS", Resource: "wood", Tier: 6, Refined: true}
	prices := map[string]map[string]Price{
		// Martlock -> Lymhurst: T4 +~38/unit (light), T6 +~190/unit (heavier)
		t4.ID: {"Martlock": p(100), "Lymhurst": p(148), "Thetford": p(90)},
		t6.ID: {"Martlock": p(1000), "Lymhurst": p(1284), "Caerleon": {SellMin: 2000, SellMinDate: now, Avg: true}},
	}
	r := computeTrade("Martlock", []Item{t4, t6}, prices, 517, 0, true, nil)
	if len(r.Destinations) != 6 {
		t.Fatalf("destinations = %d, want 6 (start city excluded)", len(r.Destinations))
	}
	for _, d := range r.Destinations {
		if d.City == "Martlock" {
			t.Fatal("start city listed as a destination")
		}
	}
	// Caerleon: (2000*0.935-1000) = 870/unit * 453 units = 394,110; ranked first and marked as an average price.
	top := r.Destinations[0]
	if top.City != "Caerleon" || !top.Best.AvgSell || top.Best.Units != 453 {
		t.Fatalf("top = %s %+v", top.City, top.Best)
	}
	var ly TradeDest
	for _, d := range r.Destinations {
		if d.City == "Lymhurst" {
			ly = d
		}
	}
	// T4: 517/0.51 = 1013 units, 148*0.935-100 = 38.38/unit; T6: 517/1.14 = 453 units, 1284*0.935-1000 = 200.54/unit
	if ly.Tiers[0].Units != 1013 || ly.Tiers[2].Units != 453 || ly.Tiers[1] != nil {
		t.Fatalf("Lymhurst tiers = %+v", ly.Tiers)
	}
	if ly.Best != ly.Tiers[2] {
		t.Errorf("best should be T6 (more per trip): %+v", ly.Best)
	}
	if d := ly.Best.Trip - 200.54*453; d > 0.01 || d < -0.01 {
		t.Errorf("trip = %v", ly.Best.Trip)
	}
	// Thetford sells T4 below the Martlock price: a loss, still shown.
	for _, d := range r.Destinations {
		if d.City == "Thetford" && (d.Best == nil || d.Best.Trip >= 0) {
			t.Errorf("Thetford should be a loss: %+v", d.Best)
		}
		if d.City == "Bridgewatch" && d.Best != nil {
			t.Errorf("Bridgewatch has no prices: %+v", d.Best)
		}
	}
	if last := r.Destinations[len(r.Destinations)-1]; last.Best != nil {
		t.Errorf("unpriced cities go last, got %s", last.City)
	}

	// A budget smaller than the load caps the units.
	b := computeTrade("Martlock", []Item{t6}, prices, 517, 50_000, true, nil)
	for _, d := range b.Destinations {
		if d.City == "Lymhurst" && (d.Best.Units != 50 || !d.Best.ByBudget || d.Best.Need != 50_000) {
			t.Errorf("budget cap: %+v", d.Best)
		}
	}
}
