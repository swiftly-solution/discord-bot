package store

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	Honeypot    = "honeypot"
	Autopublish = "autopublish"
	Forum       = "forum"
)

const (
	selectChannels = `SELECT channel_id FROM channel_roles WHERE role = $1`
	insertChannel  = `INSERT INTO channel_roles (id, guild_id, channel_id, role, created_at)
 SELECT $1, $2, $3, $4, now()
 WHERE NOT EXISTS (SELECT 1 FROM channel_roles WHERE guild_id = $2 AND channel_id = $3 AND role = $4)`
	deleteChannel = `DELETE FROM channel_roles WHERE guild_id = $1 AND channel_id = $2 AND role = $3`
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Channels(ctx context.Context, role string) ([]string, error) {
	rows, err := s.db.Query(ctx, selectChannels, role)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

// Add gives the channel the role; false means it already had it.
func (s *Store) Add(ctx context.Context, guildID, channelID, role string) (bool, error) {
	tag, err := s.db.Exec(ctx, insertChannel, uuid.NewString(), guildID, channelID, role)
	return tag.RowsAffected() > 0, err
}

// Remove takes the role from the channel; false means it did not have it.
func (s *Store) Remove(ctx context.Context, guildID, channelID, role string) (bool, error) {
	tag, err := s.db.Exec(ctx, deleteChannel, guildID, channelID, role)
	return tag.RowsAffected() > 0, err
}

// Cache is an in-memory copy of one role's channels, keyed by channel id, that a module refreshes.
type Cache[V any] struct {
	mu sync.RWMutex
	m  map[string]V
}

func (c *Cache[V]) Get(channelID string) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[channelID]
	return v, ok
}

func (c *Cache[V]) Replace(m map[string]V) {
	c.mu.Lock()
	c.m = m
	c.mu.Unlock()
}
