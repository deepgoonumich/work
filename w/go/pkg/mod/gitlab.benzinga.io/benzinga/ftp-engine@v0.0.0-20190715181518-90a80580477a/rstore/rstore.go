package rstore

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tlog "github.com/opentracing/opentracing-go/log"
	otredis "github.com/smacker/opentracing-go-redis"
	"go.uber.org/zap"
)

type Client struct {
	client *redis.Client
	logger *zap.Logger
}

const ftpEnginePrefix = "ftp-engine"

var ErrKeyNotFound = errors.New("key not found in redis")

func NewClient(l *zap.Logger, redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	c := redis.NewClient(opts)

	if err := c.Set(ftpEnginePrefix+":"+"test", "", time.Millisecond*5).Err(); err != nil {
		return nil, err
	}

	return &Client{
		client: c,
		logger: l.Named("redis"),
	}, nil
}

func (c *Client) Status(ctx context.Context) error {
	span, subCtx := opentracing.StartSpanFromContext(ctx, "redis.Status")
	defer span.Finish()
	ext.DBType.Set(span, "redis")

	client := otredis.WrapRedisClient(subCtx, c.client)
	if err := client.Ping().Err(); err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		c.logger.Error("Redis Ping Error", zap.Error(err))
		return err
	}
	return nil
}

func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		c.logger.Error("Redis Close Error", zap.Error(err))
		return err
	}
	return nil
}
