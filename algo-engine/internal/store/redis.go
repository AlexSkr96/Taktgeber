package store

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client *redis.Client
}

func NewStore(addr string) *Store {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Store{client: client}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
