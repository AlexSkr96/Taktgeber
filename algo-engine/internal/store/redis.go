package store

import (
	"context"
	"fmt"
	"time"

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

func (s *Store) AddPricePoint(ctx context.Context, coin string, price float64, timestamp time.Time) error {
	pricePoint := redis.Z{
		Score:  float64(timestamp.UnixMilli()),
		Member: price,
	}

	s.client.ZAdd(ctx, fmt.Sprintf("price:%s", coin), pricePoint)

	return nil
}
