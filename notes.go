package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Notes are facts the coach found on the web, kept in PostgreSQL so every replica sees the same set.

type Source struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type Note struct {
	ID      int64     `json:"id"`
	Topic   string    `json:"topic"`
	Body    string    `json:"body"`
	Sources []Source  `json:"sources"`
	Updated time.Time `json:"updated"`
}

type Notes struct {
	db *pgxpool.Pool

	mu    sync.Mutex
	ready bool
}

// The advisory lock lets replicas that start together create the table once; IF NOT EXISTS makes reruns no-ops.
const notesSchema = `
SELECT pg_advisory_xact_lock(7316001);
CREATE TABLE IF NOT EXISTS notes (
	id         bigserial PRIMARY KEY,
	topic_key  text NOT NULL UNIQUE,
	topic      text NOT NULL,
	body       text NOT NULL,
	sources    jsonb NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);`

const (
	maxTopic     = 80
	maxNoteBody  = 2000
	maxNoteFound = 5
	maxNotesList = 500
)

// OpenNotes does not touch the database; the schema is ensured on first use, so a database that
// starts later than the app (or restarts) only delays notes and never takes the app down.
func OpenNotes(url string) (*Notes, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.ConnConfig.ConnectTimeout = 3 * time.Second // a down database fails fast instead of holding requests
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &Notes{db: pool}, nil
}

// checked re-runs the schema on next use if the table vanished (e.g. the volume was recreated).
func (n *Notes) checked(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42P01" { // undefined_table
		n.mu.Lock()
		n.ready = false
		n.mu.Unlock()
	}
	return err
}

func (n *Notes) ensure(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.ready {
		return nil
	}
	tx, err := n.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("자료 저장소에 연결하지 못했습니다: %w", err)
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	for _, stmt := range strings.Split(notesSchema, ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("자료 저장소를 준비하지 못했습니다: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	n.ready = true
	return nil
}

// topicKey makes "개인 섬  업그레이드" and "개인 섬 업그레이드" the same note.
func topicKey(topic string) string {
	return strings.ToLower(strings.Join(strings.Fields(topic), " "))
}

var errBadNote = errors.New("bad note")

func validateNote(topic, body string, sources []Source) error {
	switch {
	case topicKey(topic) == "" || utf8.RuneCountInString(topic) > maxTopic:
		return fmt.Errorf("%w: topic must be 1-%d characters", errBadNote, maxTopic)
	case strings.TrimSpace(body) == "" || utf8.RuneCountInString(body) > maxNoteBody:
		return fmt.Errorf("%w: body must be 1-%d characters", errBadNote, maxNoteBody)
	case len(sources) == 0:
		return fmt.Errorf("%w: at least one source", errBadNote)
	}
	return nil
}

// Save writes a note by topic: saving the same topic again replaces it, so a retried request never duplicates.
func (n *Notes) Save(ctx context.Context, topic, body string, sources []Source) error {
	topic = strings.Join(strings.Fields(topic), " ")
	if err := validateNote(topic, body, sources); err != nil {
		return err
	}
	if err := n.ensure(ctx); err != nil {
		return err
	}
	_, err := n.db.Exec(ctx, `
INSERT INTO notes (topic_key, topic, body, sources) VALUES ($1, $2, $3, $4)
ON CONFLICT (topic_key) DO UPDATE SET topic = EXCLUDED.topic, body = EXCLUDED.body, sources = EXCLUDED.sources, updated_at = now()`,
		topicKey(topic), topic, strings.TrimSpace(body), sources)
	return n.checked(err)
}

func likePattern(word string) string {
	return "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(word) + "%"
}

// Find returns the notes matching the most query words (2+ characters), newest first among equals.
func (n *Notes) Find(ctx context.Context, query string) ([]Note, error) {
	var pats []string
	for _, w := range strings.Fields(query) {
		if utf8.RuneCountInString(w) >= 2 && len(pats) < 8 {
			pats = append(pats, likePattern(w))
		}
	}
	if len(pats) == 0 {
		return []Note{}, nil
	}
	if err := n.ensure(ctx); err != nil {
		return nil, err
	}
	return n.query(ctx, `
SELECT id, topic, body, sources, updated_at FROM notes
WHERE topic ILIKE ANY($1) OR body ILIKE ANY($1)
ORDER BY (SELECT count(*) FROM unnest($1::text[]) p WHERE topic ILIKE p OR body ILIKE p) DESC, updated_at DESC
LIMIT $2`, pats, maxNoteFound)
}

func (n *Notes) List(ctx context.Context) ([]Note, error) {
	if err := n.ensure(ctx); err != nil {
		return nil, err
	}
	return n.query(ctx, `SELECT id, topic, body, sources, updated_at FROM notes ORDER BY updated_at DESC, id DESC LIMIT $1`, maxNotesList)
}

// Delete is idempotent: deleting a note that is already gone succeeds.
func (n *Notes) Delete(ctx context.Context, id int64) error {
	if err := n.ensure(ctx); err != nil {
		return err
	}
	_, err := n.db.Exec(ctx, `DELETE FROM notes WHERE id = $1`, id)
	return n.checked(err)
}

func (n *Notes) query(ctx context.Context, sql string, args ...any) ([]Note, error) {
	rows, err := n.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, n.checked(err)
	}
	defer rows.Close()
	out := []Note{}
	for rows.Next() {
		var x Note
		if err := rows.Scan(&x.ID, &x.Topic, &x.Body, &x.Sources, &x.Updated); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, n.checked(rows.Err()) // pgx reports query errors here, not from Query
}
