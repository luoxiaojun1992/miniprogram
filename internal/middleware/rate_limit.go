package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var rateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return current
`)

// RateLimitStore is the storage backend used by rate limit middleware.
type RateLimitStore interface {
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)
}

type redisRateLimitStore struct {
	client redis.UniversalClient
}

// NewRedisRateLimitStore creates a redis-backed rate-limit store.
func NewRedisRateLimitStore(client redis.UniversalClient) RateLimitStore {
	if client == nil {
		return nil
	}
	return &redisRateLimitStore{client: client}
}

func (s *redisRateLimitStore) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	res, err := rateLimitScript.Run(ctx, s.client, []string{key}, int64(window.Seconds())).Int64()
	if err != nil {
		return 0, err
	}
	return res, nil
}

// RateLimitMiddleware limits requests by client IP + route in a fixed window.
func RateLimitMiddleware(store RateLimitStore, limit int64, window time.Duration, log *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if store == nil || limit <= 0 || window <= 0 {
			ctx.Next()
			return
		}

		route := ctx.FullPath()
		if route == "" {
			route = ctx.Request.URL.Path
		}
		route = strings.ReplaceAll(route, " ", "")
		key := fmt.Sprintf("ratelimit:%s:%s", ctx.ClientIP(), route)

		count, err := store.Increment(ctx.Request.Context(), key, window)
		if err != nil {
			if log != nil {
				log.WithError(err).WithField("rate_limit_key", key).Warn("rate limit backend failed, allowing request")
			}
			ctx.Next()
			return
		}
		if count > limit {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429001,
				"message": "请求过于频繁，请稍后重试",
				"data":    nil,
			})
			return
		}
		ctx.Next()
	}
}
