package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

//go:embed web
var webFS embed.FS

//go:embed names.json
var namesJSON []byte

//go:embed knowledge/guide.md
var guideMD string

type App struct {
	market   *Market
	gameinfo *Gameinfo
	names    map[string]string
	coach    *Coach
	draining atomic.Bool
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func (a *App) Refine(ctx context.Context, res string, tier, enchant int, premium bool, fee int, fresh bool) (RefineResult, error) {
	rec, err := newRecipe(res, tier, enchant)
	if err != nil {
		return RefineResult{}, err
	}
	ids := []string{rec.Raw.ID, rec.Product.ID}
	if rec.Lower != nil {
		ids = append(ids, rec.Lower.ID)
	}
	prices, err := a.market.Prices(ctx, ids, fresh)
	if err != nil {
		return RefineResult{}, err
	}
	for _, it := range []*Item{&rec.Raw, &rec.Product, rec.Lower} {
		if it != nil {
			it.Name = a.names[it.ID]
		}
	}
	return computeRefine(rec, prices, premium, fee, resources[res].bonusCity, time.Now()), nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	static, _ := fs.Sub(webFS, "web")
	files := http.FileServerFS(static)
	mux.Handle("GET /", files)
	mux.Handle("GET /fonts/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // font files carry their version in the name
		files.ServeHTTP(w, r)
	}))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if a.draining.Load() {
			fail(w, http.StatusServiceUnavailable, "draining")
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"coach": a.coach != nil})
	})

	mux.HandleFunc("GET /api/refine", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		tier, err1 := strconv.Atoi(q.Get("tier"))
		enchant, err2 := strconv.Atoi(q.Get("enchant"))
		fee, err3 := strconv.Atoi(q.Get("fee"))
		if q.Get("fee") == "" {
			fee, err3 = 0, nil
		}
		if err1 != nil || err2 != nil || err3 != nil || fee < 0 || fee > 5000 {
			fail(w, http.StatusBadRequest, "tier, enchant, fee(0-5000)는 숫자여야 합니다")
			return
		}
		res, err := a.Refine(r.Context(), q.Get("resource"), tier, enchant, q.Get("premium") == "1", fee, q.Get("fresh") == "1")
		if err != nil {
			code := http.StatusBadGateway
			if _, recErr := newRecipe(q.Get("resource"), tier, enchant); recErr != nil {
				code = http.StatusBadRequest
			}
			fail(w, code, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, res)
	})

	mux.HandleFunc("POST /api/chat", func(w http.ResponseWriter, r *http.Request) {
		if a.coach == nil {
			fail(w, http.StatusServiceUnavailable, "서버에 ANTHROPIC_API_KEY가 없어 코치를 쓸 수 없습니다")
			return
		}
		var body struct {
			Messages []ChatMsg `json:"messages"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&body); err != nil {
			fail(w, http.StatusBadRequest, "요청 형식이 잘못되었습니다")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
		defer cancel()
		reply, err := a.coach.Reply(ctx, body.Messages)
		if err != nil {
			if errors.Is(err, errBadChat) {
				fail(w, http.StatusBadRequest, err.Error())
				return
			}
			slog.Error("coach", "err", err)
			var apiErr *anthropic.Error
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests {
				fail(w, http.StatusServiceUnavailable, "Claude API 사용량 한도에 걸렸습니다. 잠시 뒤 다시 시도하세요")
				return
			}
			fail(w, http.StatusBadGateway, "코치 응답을 받지 못했습니다")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"reply": reply})
	})
	return mux
}

// healthcheck lets the scratch image probe itself (no curl inside).
func healthcheck() {
	c := http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get("http://127.0.0.1" + env("ADDR", ":8080") + "/healthz")
	if err != nil || resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		healthcheck()
	}
	app := &App{
		market:   NewMarket(env("AODP_BASE", "https://east.albion-online-data.com")),
		gameinfo: NewGameinfo(env("GAMEINFO_BASE", "https://gameinfo-sgp.albiononline.com/api/gameinfo")),
	}
	if err := json.Unmarshal(namesJSON, &app.names); err != nil {
		slog.Error("names.json", "err", err)
		os.Exit(1)
	}
	model := env("CLAUDE_MODEL", "claude-sonnet-5-5")
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		coach, err := NewCoach(anthropic.NewClient(), model, guideMD, app)
		if err != nil {
			slog.Error("coach tools", "err", err)
			os.Exit(1)
		}
		app.coach = coach
	} else {
		slog.Warn("ANTHROPIC_API_KEY not set: coach disabled, calculator still works")
	}

	srv := &http.Server{Addr: env("ADDR", ":8080"), Handler: app.routes(), ReadHeaderTimeout: 10 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		<-ctx.Done()
		// Shutdown closes the listener at once (the proxy fails over to other replicas)
		// and waits for in-flight coach answers before exiting.
		app.draining.Store(true)
		sctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
		defer cancel()
		if err := srv.Shutdown(sctx); err != nil {
			slog.Error("shutdown", "err", err)
		}
	}()
	slog.Info("listening", "addr", srv.Addr, "coach", app.coach != nil, "model", model)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("serve", "err", err)
		os.Exit(1)
	}
	// ListenAndServe returns as soon as Shutdown starts; exiting here would cut in-flight answers.
	<-drained
	slog.Info("stopped")
}
