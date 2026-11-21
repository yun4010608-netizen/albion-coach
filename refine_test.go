package main

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

func TestRecipe(t *testing.T) {
	cases := []struct {
		res           string
		tier, enchant int
		raw, lower    string
		count         int
	}{
		{"ore", 2, 0, "T2_ORE", "", 1},
		{"ore", 4, 1, "T4_ORE_LEVEL1@1", "T3_METALBAR", 2},           // T4.x takes plain T3
		{"hide", 5, 2, "T5_HIDE_LEVEL2@2", "T4_LEATHER_LEVEL2@2", 3}, // T5.x+ same enchant
		{"wood", 8, 0, "T8_WOOD", "T7_PLANKS", 5},
	}
	for _, c := range cases {
		r, err := newRecipe(c.res, c.tier, c.enchant)
		if err != nil {
			t.Fatal(c, err)
		}
		lower := ""
		if r.Lower != nil {
			lower = r.Lower.ID
		}
		if r.Raw.ID != c.raw || lower != c.lower || r.Raw.Count != c.count {
			t.Errorf("%s T%d.%d: got %s x%d + %q", c.res, c.tier, c.enchant, r.Raw.ID, r.Raw.Count, lower)
		}
	}
	for _, bad := range [][3]any{{"rock", 4, 1}, {"ore", 3, 1}, {"gold", 4, 0}, {"ore", 9, 0}} {
		if _, err := newRecipe(bad[0].(string), bad[1].(int), bad[2].(int)); err == nil {
			t.Errorf("%v should fail", bad)
		}
	}
}

func TestReturnRate(t *testing.T) {
	for _, c := range []struct {
		spec, focus bool
		want        float64
	}{{false, false, .1525}, {true, false, .3671}, {false, true, .4350}, {true, true, .5392}} {
		if got := returnRate(c.spec, c.focus); math.Abs(got-c.want) > .0001 {
			t.Errorf("rrr(%v,%v)=%.4f want %.4f", c.spec, c.focus, got, c.want)
		}
	}
}

func TestComputeRefine(t *testing.T) {
	rec, _ := newRecipe("ore", 5, 0)
	now := time.Now().UTC()
	p := func(v int) Price { return Price{SellMin: v, SellMinDate: now} }
	prices := map[string]map[string]Price{
		"T5_ORE":      {"Thetford": p(300), "Martlock": p(250)},
		"T4_METALBAR": {"Thetford": p(330), "Martlock": p(300)},
		"T5_METALBAR": {"Thetford": p(1240), "Martlock": p(1300)},
	}
	res := computeRefine(rec, prices, true, 500, "Thetford", now)
	if res.Rows[0].Missing || res.Rows[len(res.Rows)-1].Missing != true {
		t.Fatal("priced rows must sort before missing ones")
	}
	var th Row
	for _, r := range res.Rows {
		if r.City == "Thetford" {
			th = r
		}
	}
	// cost 3*300+330=1230; keep 1-0.3671; fee 2^5*0.1125*5 = 18; sell 1240*(1-0.065)
	want := 1240*(1-0.065) - (1230*(1-returnRate(true, false)) + 18)
	if math.Abs(th.Profit-want) > 0.01 {
		t.Errorf("Thetford profit %.2f want %.2f", th.Profit, want)
	}
	if res.Route == nil || res.Route.Buy != "Martlock" || res.Route.Refine != "Thetford" || res.Route.Sell != "Martlock" {
		t.Errorf("route %+v", res.Route)
	}
	// A stale sell quote must drop out of the route; ages are reported even for incomplete rows.
	prices["T5_METALBAR"]["Martlock"] = Price{SellMin: 9999, SellMinDate: now.Add(-7 * time.Hour)}
	prices["T5_ORE"]["Lymhurst"] = Price{SellMin: 100, SellMinDate: now.Add(-time.Hour)}
	res = computeRefine(rec, prices, true, 0, "Thetford", now)
	if res.Route == nil || res.Route.Sell != "Thetford" {
		t.Errorf("stale price used in route: %+v", res.Route)
	}
	for _, r := range res.Rows {
		if r.City == "Lymhurst" && (!r.Missing || r.Updated.IsZero()) {
			t.Errorf("incomplete row must be missing but dated: %+v", r)
		}
	}
}

func TestMarketFallbackAndBackoff(t *testing.T) {
	now := time.Now().UTC()
	ts := func(d time.Duration) string { return now.Add(d).Format("2006-01-02T15:04:05") }
	day := now.Add(-24 * time.Hour).Truncate(24 * time.Hour).Format("2006-01-02T15:04:05")
	calls, limited := 0, false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if limited {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		if strings.Contains(r.URL.Path, "/history/") {
			w.Write([]byte(`[{"item_id":"T4_ORE","location":"Thetford","data":[{"avg_price":150,"timestamp":"` + day + `"}]},
				{"item_id":"T4_ORE","location":"Martlock","data":[{"avg_price":999,"timestamp":"` + day + `"}]}]`))
			return
		}
		w.Write([]byte(`[{"item_id":"T4_ORE","city":"Martlock","sell_price_min":120,"sell_price_min_date":"` + ts(-time.Hour) + `"},
			{"item_id":"T4_ORE","city":"Thetford","sell_price_min":0,"sell_price_min_date":"0001-01-01T00:00:00"}]`))
	}))
	defer srv.Close()
	m := NewMarket(srv.URL)
	got, err := m.Prices(context.Background(), []string{"T4_ORE"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if p := got["T4_ORE"]["Martlock"]; p.SellMin != 120 || p.Avg {
		t.Errorf("a fresh live order must win over the average: %+v", p)
	}
	if p := got["T4_ORE"]["Thetford"]; p.SellMin != 150 || !p.Avg {
		t.Errorf("a missing live order must fall back to the daily average: %+v", p)
	}

	limited, m.minAge = true, 0
	if _, err := m.Prices(context.Background(), []string{"T4_ORE"}, true); err != nil {
		t.Fatal("a 429 with cached data must serve the cache:", err)
	}
	before := calls
	if _, err := m.Prices(context.Background(), []string{"T5_ORE"}, false); err == nil || calls != before {
		t.Fatalf("while blocked, uncached keys must fail without calling upstream (calls %d->%d, err=%v)", before, calls, err)
	}
}

func TestValidateChat(t *testing.T) {
	ok := []ChatMsg{{"user", "hi"}, {"assistant", "yo"}, {"user", "next"}}
	if err := validateChat(ok); err != nil {
		t.Fatal(err)
	}
	// 8000 Korean characters are 24,000 bytes; the limit counts characters.
	if err := validateChat([]ChatMsg{{"user", strings.Repeat("섬", 8000)}}); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]ChatMsg{nil, {{"assistant", "x"}}, {{"user", "a"}, {"assistant", "b"}}, {{"user", " "}}, {{"system", "x"}}, {{"user", strings.Repeat("섬", 8001)}}} {
		if validateChat(bad) == nil {
			t.Errorf("%v should fail", bad)
		}
	}
}

func TestSources(t *testing.T) {
	var msg anthropic.BetaMessage
	raw := `{"content":[
		{"type":"server_tool_use","id":"s1","name":"web_search","input":{"query":"albion island"}},
		{"type":"text","text":"섬은 ","citations":[
			{"type":"web_search_result_location","url":"https://wiki.albiononline.com/wiki/Island","title":"Island [Wiki]","encrypted_index":"x","cited_text":"a"},
			{"type":"web_search_result_location","url":"https://wiki.albiononline.com/wiki/Island","title":"dup","encrypted_index":"x","cited_text":"a"}]},
		{"type":"text","text":"프리미엄이 필요합니다.","citations":[
			{"type":"web_search_result_location","url":"https://example.com/a_(b)","title":"","encrypted_index":"x","cited_text":"b"}]}]}`
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	want := "\n\n출처:\n- [Island (Wiki)](https://wiki.albiononline.com/wiki/Island)\n- [https://example.com/a_(b)](https://example.com/a_(b%29)"
	if got := sources(msg.Content); got != want {
		t.Errorf("sources = %q, want %q", got, want)
	}
	if got := sources(nil); got != "" {
		t.Errorf("no citations should add nothing, got %q", got)
	}
	if got := withoutSources("답입니다." + want); got != "답입니다." {
		t.Errorf("withoutSources = %q", got)
	}
	if got := withoutSources("출처가 없는 답"); got != "출처가 없는 답" {
		t.Errorf("withoutSources changed a plain answer: %q", got)
	}
}
