package main

import "testing"

func TestLoadRoads(t *testing.T) {
	roads, err := LoadRoads(zonesJSON)
	if err != nil {
		t.Fatal(err)
	}
	for _, from := range cities {
		if len(roads[from]) != len(cities)-1 {
			t.Fatalf("%s has roads to %d cities", from, len(roads[from]))
		}
	}
	// Martlock -> Thetford stays in blue zones (game data 2026-11-26).
	if r := roads["Martlock"]["Thetford"]; !r.Found || r.Red+r.Black+r.Yellow != 0 || len(r.Zones) == 0 {
		t.Errorf("Martlock->Thetford = %+v", r)
	}
	// Caerleon cannot be reached without red zones.
	if r := roads["Martlock"]["Caerleon"]; !r.Found || r.Red == 0 {
		t.Errorf("Martlock->Caerleon = %+v", r)
	}
	// Fort Sterling has a yellow road, and a shorter one through red.
	if r := roads["Martlock"]["Fort Sterling"]; r.Red != 0 || r.Shortcut == nil || r.Shortcut.Red == 0 || r.Shortcut.Hops >= len(r.Zones) {
		t.Errorf("Martlock->Fort Sterling = %+v shortcut %+v", r, r.Shortcut)
	}
	// Brecilien is reached through the Mists only.
	if r := roads["Martlock"]["Brecilien"]; r.Found || len(r.Zones) != 0 {
		t.Errorf("Martlock->Brecilien = %+v", r)
	}
	// Roads read the same both ways.
	if a, b := roads["Martlock"]["Lymhurst"], roads["Lymhurst"]["Martlock"]; len(a.Zones) != len(b.Zones) || a.Red != b.Red {
		t.Errorf("asymmetric: %d/%d zones", len(a.Zones), len(b.Zones))
	}
}

func TestCheapest(t *testing.T) {
	// 0 - 1 - 3 and 0 - 2 - 3, where zone 1 costs more.
	adj := [][]int{{1, 2}, {0, 3}, {0, 3}, {1, 2}}
	path, ok := cheapest(adj, 0, 3, func(i int) int { return map[int]int{1: 100}[i] + 1 })
	if !ok || len(path) != 1 || path[0] != 2 {
		t.Errorf("path = %v %v", path, ok)
	}
	if path, ok := cheapest([][]int{{1}, {0}}, 0, 1, func(int) int { return 1 }); !ok || len(path) != 0 {
		t.Errorf("neighbours: %v %v", path, ok)
	}
	if _, ok := cheapest([][]int{{}, {}}, 0, 1, func(int) int { return 1 }); ok {
		t.Error("unreachable should report false")
	}
}
