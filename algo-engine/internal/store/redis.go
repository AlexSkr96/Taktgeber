package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/AlexSkr96/Taktgeber/algo-engine/types"
	"github.com/redis/go-redis/v9"
)

var pricesKey = "prices"

type Store struct {
	client *redis.Client
}

func NewStore(addr string) *Store {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Store{client: client}
}

func newKey(category, name string) string {
	return fmt.Sprintf("%s:%s", category, name)
}

func (s *Store) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *Store) AddPrice(ctx context.Context, coin string, price float64, timestamp time.Time) error {
	pricePoint := redis.Z{
		Score:  float64(timestamp.UnixMilli()),
		Member: price,
	}

	key := newKey(pricesKey, coin)
	s.client.ZAdd(ctx, key, pricePoint)

	return nil
}

func (s *Store) GetRecentPrices(ctx context.Context, coin string, since time.Duration) ([]types.PricePoint, error) {
	pricePoints := []types.PricePoint{}
	key := newKey(pricesKey, coin)
	minScore := time.Now().Add(-since).UnixMilli()

	result, err := s.client.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
		Key:     key,
		Start:   minScore,
		Stop:    "+inf",
		ByScore: true,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("Couldn't get recent prices from Redis:\n%v", err)
	}

	for _, subResult := range result {
		sPrice, ok := subResult.Member.(string)
		if !ok {
			return nil, fmt.Errorf("Couldn't parse \"prices\" Redis member %v to string:\n%v", subResult.Member, err)
		}
		price, err := strconv.ParseFloat(sPrice, 64)
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse string \"%v\" to float64:\n%v", sPrice, err)
		}
		pricePoints = append(pricePoints,
			types.PricePoint{
				UnixTimestamp: int64(subResult.Score),
				Price:         price,
			},
		)
	}

	return pricePoints, nil
}
