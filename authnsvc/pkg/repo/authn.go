package repo

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type AuthNRepository interface {
	Blacklist(ctx context.Context, jti string, exp time.Duration) (bool, error)
	IsBlacklisted(ctx context.Context, jti string) bool
}

type BasicAuthRepo struct {
	rdb *redis.Client
}

func NewBasicAuthRepo(rdb *redis.Client) AuthNRepository {
	if rdb == nil {
		return nil
	}

	return &BasicAuthRepo{
		rdb: rdb,
	}
}

func (b *BasicAuthRepo) Blacklist(ctx context.Context, jti string, exp time.Duration) (bool, error) {
	err := b.rdb.SetEX(ctx, jti, "", exp).Err()
	if err != nil {
		return false, err
	}
	return true, err
}

func (b *BasicAuthRepo) IsBlacklisted(ctx context.Context, jti string) bool {
	_, err := b.rdb.Get(ctx, jti).Result()
	// check if err == redis.Nil [key does not exist]
	return err != redis.Nil
}
