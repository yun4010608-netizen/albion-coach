package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Price is one item's cheapest sell order in one city. A zero SellMin means no data.
type Price struct {
	ItemID      string    `json:"itemId"`
	City        string    `json:"city"`
	SellMin     int       `json:"sellMin"`
	SellMinDate time.Time `json:"sellMinDate"`
	Avg         bool      `json:"avg"` // daily average trade price, used when no live order is fresh
}

type cacheEntry struct {
	at     time.Time
	prices map[string]map[string]Price
}

// Market reads the Albion Online Data Project API for the Asia server.
// shortcut: cache is per instance, so upstream calls scale with replicas; move it to a shared store (e.g. Redis) past ~5 replicas or on 429s.
type Market struct {
	base    string
	http    *http.Client
	ttl     time.Duration
	minAge  time.Duration // a forced refresh still reuses data younger than this
	sf      singleflight.Group
	mu      sync.Mutex
	cache   map[string]cacheEntry
	blocked time.Time
}

func NewMarket(base string) *Market {
	return &Market{
		base:   strings.TrimRight(base, "/"),
		http:   &http.Client{Timeout: 15 * time.Second}, // default transport sends Accept-Encoding: gzip, required by AODP
		ttl:    5 * time.Minute,
		minAge: time.Minute,
		cache:  map[string]cacheEntry{},
	}
}

type aodpRow struct {
	ItemID      string `json:"item_id"`
	City        string `json:"city"`
	SellMin     int    `json:"sell_price_min"`
	SellMinDate string `json:"sell_price_min_date"`
}

// Prices returns prices[itemID][city]; fresh asks for data newer than minAge.
func (m *Market) Prices(ctx context.Context, ids []string, fresh bool) (map[string]map[string]Price, error) {
	ids = append([]string(nil), ids...)
	sort.Strings(ids)
	key := strings.Join(ids, ",")

	m.mu.Lock()
	e, hit := m.cache[key]
	blocked := time.Now().Before(m.blocked)
	m.mu.Unlock()
	age := time.Since(e.at)
	if hit && (age < m.minAge || (!fresh && age < m.ttl) || blocked) {
		return e.prices, nil
	}
	if blocked {
		return nil, fmt.Errorf("시세 API 호출 한도에 걸려 %s까지 쉬는 중", m.blocked.Local().Format("15:04"))
	}

	// One caller leaving must not fail the fetch other callers share.
	v, err, _ := m.sf.Do(key, func() (any, error) { return m.fetch(context.WithoutCancel(ctx), ids) })
	if err != nil {
		if hit { // stale beats nothing
			return e.prices, nil
		}
		return nil, err
	}
	p := v.(map[string]map[string]Price)
	m.mu.Lock()
	m.cache[key] = cacheEntry{at: time.Now(), prices: p}
	m.mu.Unlock()
	return p, nil
}

// get decodes one AODP endpoint into v, backing off for 5 minutes on a 429
// (repeated calls after a 429 earn a firewall ban).
func (m *Market) get(ctx context.Context, path string, ids []string, extra url.Values, v any) error {
	q := url.Values{"locations": {strings.Join(cities, ",")}, "qualities": {"1"}}
	for k, vs := range extra {
		q[k] = vs
	}
	u := fmt.Sprintf("%s/api/v2/stats/%s/%s.json?%s", m.base, path, strings.Join(ids, ","), q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "albion-coach/1.0")
	resp, err := m.http.Do(req)
	if err != nil {
		return fmt.Errorf("시세 API 연결 실패: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		m.mu.Lock()
		m.blocked = time.Now().Add(5 * time.Minute)
		m.mu.Unlock()
		return fmt.Errorf("시세 API 호출 한도 초과, 5분 뒤 다시 시도")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("시세 API 응답 %d", resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(v); err != nil {
		return fmt.Errorf("시세 API 응답 해석 실패: %w", err)
	}
	return nil
}

type aodpHistory struct {
	ItemID   string `json:"item_id"`
	Location string `json:"location"`
	Data     []struct {
		AvgPrice  int    `json:"avg_price"`
		Timestamp string `json:"timestamp"`
	} `json:"data"`
}

// AODP dates carry no zone and are UTC; 0001-01-01 marks "no data".
func aodpTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05", s)
	return t.UTC()
}

// fetch takes the cheapest live sell order per city. Asia order data is sparse, so when it is
// missing or older than staleAfter, the city's latest daily average (last 7 days) fills in, marked Avg.
func (m *Market) fetch(ctx context.Context, ids []string) (map[string]map[string]Price, error) {
	var rows []aodpRow
	if err := m.get(ctx, "prices", ids, nil, &rows); err != nil {
		return nil, err
	}
	out := map[string]map[string]Price{}
	for _, r := range rows {
		if out[r.ItemID] == nil {
			out[r.ItemID] = map[string]Price{}
		}
		t := aodpTime(r.SellMinDate)
		if t.Year() < 2000 {
			r.SellMin = 0
		}
		out[r.ItemID][r.City] = Price{ItemID: r.ItemID, City: r.City, SellMin: r.SellMin, SellMinDate: t}
	}

	now := time.Now()
	var hist []aodpHistory
	since := now.AddDate(0, 0, -7).Format("2006-01-02")
	if err := m.get(ctx, "history", ids, url.Values{"time-scale": {"24"}, "date": {since}}, &hist); err != nil {
		return out, nil // live orders alone are still a valid answer
	}
	for _, h := range hist {
		if len(h.Data) == 0 {
			continue
		}
		last := h.Data[len(h.Data)-1]
		day := aodpTime(last.Timestamp)
		cur := out[h.ItemID][h.Location]
		liveFresh := cur.SellMin > 0 && now.Sub(cur.SellMinDate) < staleAfter
		if last.AvgPrice <= 0 || liveFresh || now.Sub(day) > 7*24*time.Hour || (cur.SellMin > 0 && !day.After(cur.SellMinDate)) {
			continue
		}
		if out[h.ItemID] == nil {
			out[h.ItemID] = map[string]Price{}
		}
		out[h.ItemID][h.Location] = Price{ItemID: h.ItemID, City: h.Location, SellMin: last.AvgPrice, SellMinDate: day, Avg: true}
	}
	return out, nil
}
