package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

func TestVerifiedNotes(t *testing.T) {
	seen := map[string]string{"https://a.example/1": "A", "https://b.example/2": "B"}
	got := verifiedNotes([]saveInput{
		{Topic: "섬", Body: "x", SourceURLs: []string{"https://a.example/1", "https://made.up/", "https://a.example/1"}},
		{Topic: "가짜", Body: "y", SourceURLs: []string{"https://made.up/"}},
	}, seen)
	if len(got) != 1 || len(got[0].Sources) != 1 || got[0].Sources[0] != (Source{URL: "https://a.example/1", Title: "A"}) {
		t.Fatalf("verifiedNotes = %+v", got)
	}
}

func TestSearchedPages(t *testing.T) {
	var msg anthropic.BetaMessage
	raw := `{"content":[{"type":"web_search_tool_result","tool_use_id":"s1","content":[
		{"type":"web_search_result","url":"https://a.example/1","title":"A","encrypted_content":"x","page_age":""}]},
		{"type":"web_search_tool_result","tool_use_id":"s2","content":{"type":"web_search_tool_result_error","error_code":"unavailable"}}]}`
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if got := searchedPages(msg.Content); len(got) != 1 || got["https://a.example/1"] != "A" {
		t.Fatalf("searchedPages = %v", got)
	}
}

func TestNoteSourcesCited(t *testing.T) {
	var msg anthropic.BetaMessage
	raw := `{"content":[{"type":"text","text":"섬은 2단계부터","citations":[
		{"type":"search_result_location","source":"https://a.example/1","title":"A guide","cited_text":"c","search_result_index":0,"start_block_index":0,"end_block_index":1}]}]}`
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if got := sources(msg.Content); got != "\n\n출처:\n- [A guide](https://a.example/1)" {
		t.Errorf("sources = %q", got)
	}
	blocks := noteResults([]Note{{Topic: "섬", Body: "b", Sources: []Source{{URL: "https://a.example/1"}}, Updated: time.Now()}}, time.Now())
	if sr := blocks[0].OfSearchResult; sr == nil || sr.Source != "https://a.example/1" || sr.Title != "섬" || !sr.Citations.Enabled.Value {
		t.Errorf("noteResults = %+v", blocks[0])
	}
}

func TestValidateNote(t *testing.T) {
	src := []Source{{URL: "https://a.example"}}
	if err := validateNote("섬", "b", src); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ topic, body string }{{" ", "b"}, {"섬", " "}, {strings.Repeat("섬", 81), "b"}} {
		if validateNote(c.topic, c.body, src) == nil {
			t.Errorf("%q/%q should fail", c.topic, c.body)
		}
	}
	if validateNote("섬", "b", nil) == nil {
		t.Error("a note needs a source")
	}
	if topicKey("  개인 섬   업그레이드 ") != topicKey("개인 섬 업그레이드") {
		t.Error("topic keys should ignore spacing")
	}
}

// Runs against a real PostgreSQL when TEST_DATABASE_URL is set (see README); skipped otherwise.
func TestNotesStore(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	// Two replicas starting together must both get a working schema.
	stores := make([]*Notes, 4)
	var wg sync.WaitGroup
	for i := range stores {
		s, err := OpenNotes(url)
		if err != nil {
			t.Fatal(err)
		}
		stores[i] = s
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.List(ctx); errOrNil(t, err) }()
	}
	wg.Wait()
	n := stores[0]
	if _, err := n.db.Exec(ctx, `TRUNCATE notes`); err != nil {
		t.Fatal(err)
	}
	src := []Source{{URL: "https://a.example/1", Title: "A"}}
	// Same topic twice (a retried request): one row, second body wins.
	for _, body := range []string{"첫 내용", "고친 내용"} {
		if err := n.Save(ctx, "개인 섬  업그레이드 비용", body, src); err != nil {
			t.Fatal(err)
		}
	}
	if err := stores[1].Save(ctx, "길드 가입", "50% 할인", src); err != nil {
		t.Fatal(err)
	}
	list, err := n.List(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("list = %+v, %v", list, err)
	}
	found, err := n.Find(ctx, "섬 업그레이드")
	if err != nil || len(found) != 1 || found[0].Body != "고친 내용" || found[0].Sources[0] != src[0] {
		t.Fatalf("find = %+v, %v", found, err)
	}
	if found, _ := n.Find(ctx, "50%"); len(found) != 1 || found[0].Topic != "길드 가입" {
		t.Errorf("a literal %% should match only itself: %+v", found)
	}
	if found, _ := n.Find(ctx, "x"); len(found) != 0 {
		t.Errorf("one-letter words are ignored: %+v", found)
	}
	for range 2 { // deleting twice is fine
		if err := n.Delete(ctx, list[0].ID); err != nil {
			t.Fatal(err)
		}
	}
	if list, _ := n.List(ctx); len(list) != 1 {
		t.Errorf("after delete: %+v", list)
	}
	// The table vanishing (volume recreated) heals on the next call after the failing one.
	if _, err := n.db.Exec(ctx, `DROP TABLE notes`); err != nil {
		t.Fatal(err)
	}
	if _, err := n.List(ctx); err == nil {
		t.Fatal("expected an error right after the drop")
	}
	if _, err := n.List(ctx); err != nil {
		t.Fatalf("schema should be recreated: %v", err)
	}
}

func errOrNil(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
