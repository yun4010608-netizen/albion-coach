package main

import (
	"encoding/json"
	"fmt"
	"slices"
)

// zones.json is the open-world zone graph from ao-bin-dumps cluster/world.xml (2026-11-26),
// extracted by scripts/zones.py. Roads of Avalon and the Mists are not in it.

type Zone struct {
	Name string `json:"name"`
	Kind string `json:"kind"` // city, blue, yellow, red, black
}

type RoadSummary struct {
	Hops  int `json:"hops"`
	Red   int `json:"red"`
	Black int `json:"black"`
}

// Road is the way between two cities that crosses the fewest black, then red zones, then the fewest zones.
type Road struct {
	Found  bool   `json:"found"` // false: no fixed road (Brecilien is reached through the Mists)
	Zones  []Zone `json:"zones"` // between the two cities, in travel order
	Yellow int    `json:"yellow"`
	Red    int    `json:"red"`
	Black  int    `json:"black"`
	// Shortcut is a shorter way that crosses more red or black zones, if there is one.
	Shortcut *RoadSummary `json:"shortcut,omitempty"`
}

// Roads holds the road between every pair of trade cities, computed once at start.
type Roads map[string]map[string]Road

func LoadRoads(data []byte) (Roads, error) {
	var raw struct {
		Zones []struct {
			ID, Name, Kind string
		} `json:"zones"`
		Edges [][2]string `json:"edges"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	n := len(raw.Zones)
	index := make(map[string]int, n)
	zones := make([]Zone, n)
	for i, z := range raw.Zones {
		index[z.ID] = i
		zones[i] = Zone{Name: z.Name, Kind: z.Kind}
	}
	adj := make([][]int, n)
	for _, e := range raw.Edges {
		a, ok1 := index[e[0]]
		b, ok2 := index[e[1]]
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("zones.json: edge %v names an unknown zone", e)
		}
		adj[a], adj[b] = append(adj[a], b), append(adj[b], a)
	}
	city := map[string]int{}
	for i, z := range zones {
		if z.Kind == "city" && slices.Contains(cities, z.Name) {
			if _, dup := city[z.Name]; dup {
				return nil, fmt.Errorf("zones.json: two cities named %s", z.Name)
			}
			city[z.Name] = i
		}
	}
	for _, c := range cities {
		if _, ok := city[c]; !ok {
			return nil, fmt.Errorf("zones.json: no zone for %s", c)
		}
	}

	safest := func(i int) int { return map[string]int{"black": 1_000_000, "red": 1_000}[zones[i].Kind] + 1 }
	shortest := func(int) int { return 1 }
	roads := Roads{}
	for _, from := range cities {
		roads[from] = map[string]Road{}
		for _, to := range cities {
			if from == to {
				continue
			}
			path, found := cheapest(adj, city[from], city[to], safest)
			r := Road{Found: found, Zones: []Zone{}}
			for _, i := range path {
				z := zones[i]
				r.Zones = append(r.Zones, z)
				switch z.Kind {
				case "yellow":
					r.Yellow++
				case "red":
					r.Red++
				case "black":
					r.Black++
				}
			}
			if short, _ := cheapest(adj, city[from], city[to], shortest); found && len(short) < len(path) {
				s := RoadSummary{Hops: len(short)}
				for _, i := range short {
					s.Red += b2i(zones[i].Kind == "red")
					s.Black += b2i(zones[i].Kind == "black")
				}
				r.Shortcut = &s
			}
			roads[from][to] = r
		}
	}
	return roads, nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// cheapest is Dijkstra over zones (cost paid on entering a zone); it returns the zones strictly between
// a and b, and false when b cannot be reached. Plain O(V²): ~500 zones, run 84 times at start.
func cheapest(adj [][]int, a, b int, cost func(int) int) ([]int, bool) {
	const inf = int(^uint(0) >> 1)
	dist := make([]int, len(adj))
	prev := make([]int, len(adj))
	done := make([]bool, len(adj))
	for i := range dist {
		dist[i], prev[i] = inf, -1
	}
	dist[a] = 0
	for {
		u := -1
		for i := range dist {
			if !done[i] && dist[i] < inf && (u < 0 || dist[i] < dist[u]) {
				u = i
			}
		}
		if u < 0 {
			return nil, false
		}
		if u == b {
			break
		}
		done[u] = true
		for _, v := range adj[u] {
			if d := dist[u] + cost(v); d < dist[v] {
				dist[v], prev[v] = d, u
			}
		}
	}
	var path []int
	for i := prev[b]; i != a; i = prev[i] {
		path = append(path, i)
	}
	slices.Reverse(path)
	return path, true
}
