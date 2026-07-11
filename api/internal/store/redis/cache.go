package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Key prefixes
	channelVelocityPrefix = "channel:velocity:"
	focusOfferedPrefix    = "channel:velocity:offered:"
	sessionPrefix         = "user:session:"
	deepWorkPrefix        = "user:deepwork:"
	aiLimitPrefix         = "rate_limit:ai:"
	lastDigestPrefix      = "user:digest:last:"
)

// Client wraps a Redis client with Signal-specific methods.
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis client.
// NewCache creates a new Redis cache client from a connection URL.
func NewCache(ctx context.Context, url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		// Fallback: treat as addr:password format
		return NewClient(url, "")
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// NewClient creates a new Redis client with explicit addr and password.
func NewClient(addr, password string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 3,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Channel velocity tracking

func (c *Client) IncrChannelVelocity(ctx context.Context, channelID string) (int, error) {
	key := channelVelocityPrefix + channelID
	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Set TTL on first increment
	if count == 1 {
		c.rdb.Expire(ctx, key, 10*time.Minute)
	}
	return int(count), nil
}

func (c *Client) GetChannelVelocity(ctx context.Context, channelID string) (int, error) {
	val, err := c.rdb.Get(ctx, channelVelocityPrefix+channelID).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *Client) ResetChannelVelocity(ctx context.Context, channelID string) error {
	return c.rdb.Del(ctx, channelVelocityPrefix+channelID).Err()
}

func (c *Client) SetFocusOffered(ctx context.Context, channelID string, ttl time.Duration) error {
	return c.rdb.Set(ctx, focusOfferedPrefix+channelID, "1", ttl).Err()
}

func (c *Client) HasFocusBeenOffered(ctx context.Context, channelID string) (bool, error) {
	_, err := c.rdb.Get(ctx, focusOfferedPrefix+channelID).Result()
	if err == redis.Nil {
		return false, nil
	}
	return err == nil, err
}

// Session management

func (c *Client) SetSession(ctx context.Context, slackUserID string, accessToken string, ttl time.Duration) error {
	return c.rdb.Set(ctx, sessionPrefix+slackUserID, accessToken, ttl).Err()
}

func (c *Client) GetSession(ctx context.Context, slackUserID string) (string, error) {
	val, err := c.rdb.Get(ctx, sessionPrefix+slackUserID).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Deep work state

func (c *Client) SetDeepWork(ctx context.Context, userID string, duration time.Duration) error {
	return c.rdb.Set(ctx, deepWorkPrefix+userID, duration.String(), duration).Err()
}

func (c *Client) GetDeepWork(ctx context.Context, userID string) (time.Duration, error) {
	val, err := c.rdb.Get(ctx, deepWorkPrefix+userID).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return time.ParseDuration(val)
}

func (c *Client) ClearDeepWork(ctx context.Context, userID string) error {
	return c.rdb.Del(ctx, deepWorkPrefix+userID).Err()
}

// Rate limiting

func (c *Client) CheckAILimit(ctx context.Context, userID string, limit int) (bool, error) {
	key := aiLimitPrefix + userID
	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		c.rdb.Expire(ctx, key, 1*time.Minute)
	}
	return int(count) <= limit, nil
}

// Digest tracking

func (c *Client) SetLastDigest(ctx context.Context, userID string, timestamp time.Time) error {
	return c.rdb.Set(ctx, lastDigestPrefix+userID, timestamp.Unix(), 24*time.Hour).Err()
}

func (c *Client) GetLastDigest(ctx context.Context, userID string) (time.Time, error) {
	val, err := c.rdb.Get(ctx, lastDigestPrefix+userID).Int64()
	if err == redis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(val, 0), nil
}

// Raw returns the underlying Redis client for advanced operations.
func (c *Client) Raw() *redis.Client {
	return c.rdb
}
