package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Gameinfo reads public player stats from the official Asia gameinfo API (no CORS, so it goes through the server).
type Gameinfo struct {
	base string
	http *http.Client
}

func NewGameinfo(base string) *Gameinfo {
	return &Gameinfo{base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: 15 * time.Second}}
}

func (g *Gameinfo) get(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "albion-coach/1.0")
	resp, err := g.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gameinfo API 응답 %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(v)
}

// Player finds a player by exact name and returns their public fame stats.
func (g *Gameinfo) Player(ctx context.Context, name string) (map[string]any, error) {
	var found struct {
		Players []struct{ Id, Name string } `json:"players"`
	}
	if err := g.get(ctx, "/search?q="+url.QueryEscape(name), &found); err != nil {
		return nil, err
	}
	for _, p := range found.Players {
		if strings.EqualFold(p.Name, name) {
			var full map[string]any
			if err := g.get(ctx, "/players/"+url.PathEscape(p.Id), &full); err != nil {
				return nil, err
			}
			out := map[string]any{}
			for _, k := range []string{"Name", "GuildName", "AllianceName", "KillFame", "DeathFame", "FameRatio", "LifetimeStatistics"} {
				if v, ok := full[k]; ok {
					out[k] = v
				}
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("Asia 서버에서 %q 캐릭터를 찾지 못함 (이름은 대소문자까지 정확히)", name)
}
