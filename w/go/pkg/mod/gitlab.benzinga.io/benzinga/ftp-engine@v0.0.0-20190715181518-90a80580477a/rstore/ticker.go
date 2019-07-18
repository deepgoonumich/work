package rstore

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tlog "github.com/opentracing/opentracing-go/log"
	otredis "github.com/smacker/opentracing-go-redis"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/zap"

	"gitlab.benzinga.io/benzinga/reference-service/reference"
)

func symbolExchangeKey(symbol, exchange string) string {
	return strings.ToUpper(strings.Join([]string{ftpEnginePrefix, "symbol-exchange", symbol, exchange}, ":"))
}

func symbolCurrencyKey(symbol, currency string) string {
	return strings.ToUpper(strings.Join([]string{ftpEnginePrefix, "symbol-currency", symbol, currency}, ":"))
}

func (c *Client) GetSymbolExchange(ctx context.Context, symbol, exchange string) (*reference.Instrument, error) {
	span, subCtx := opentracing.StartSpanFromContext(ctx, "redis.GetSymbolExchange")
	defer span.Finish()
	ext.DBType.Set(span, "redis")

	key := symbolExchangeKey(symbol, exchange)
	span.LogFields(tlog.String("key", key))

	logger := c.logger.With(zap.String("symbol", symbol), zap.String("exchange", exchange), zap.String("key", key))

	client := otredis.WrapRedisClient(subCtx, c.client)
	res, err := client.Get(key).Bytes()
	if err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		logger.Error("Redis GetSymbolExchange Error", zap.Error(err))
		return nil, err
	}

	var instr reference.Instrument
	if err := msgpack.Unmarshal(res, &instr); err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		logger.Error("GetSymbolExchange Unmarshal Error", zap.Error(err))
		return nil, err
	}

	return &instr, nil
}

func (c *Client) GetSymbolCurrency(ctx context.Context, symbol, currency string) (*reference.Instrument, error) {
	span, subCtx := opentracing.StartSpanFromContext(ctx, "redis.GetSymbolCurrency")
	defer span.Finish()
	ext.DBType.Set(span, "redis")

	key := symbolCurrencyKey(symbol, currency)
	span.LogFields(tlog.String("key", key))

	logger := c.logger.With(zap.String("symbol", symbol), zap.String("currency", currency), zap.String("key", key))

	client := otredis.WrapRedisClient(subCtx, c.client)
	res, err := client.Get(key).Bytes()
	if err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		logger.Error("Redis GetSymbolCurrency Error", zap.Error(err))
		return nil, err
	}

	var instr reference.Instrument
	if err := msgpack.Unmarshal(res, &instr); err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		logger.Error("GetSymbolCurrency Unmarshal Error", zap.Error(err))
		return nil, err
	}

	return &instr, nil
}

func (c *Client) PutSymbolCurrency(ctx context.Context, inst *reference.Instrument) error {
	span, subCtx := opentracing.StartSpanFromContext(ctx, "redis.PutSymbolCurrency")
	defer span.Finish()
	ext.DBType.Set(span, "redis")

	key := symbolCurrencyKey(inst.Symbol, inst.CurrencyID)
	span.LogFields(tlog.String("key", key))

	val, err := msgpack.Marshal(&inst)
	if err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		c.logger.Error("PutSymbolCurrency Marshal Error", zap.Error(err))
		return err
	}

	client := otredis.WrapRedisClient(subCtx, c.client)
	if err := client.Set(key, val, 0).Err(); err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		c.logger.Error("Redis PutSymbolCurrency Set Error", zap.Error(err))
		return err
	}
	return nil
}

func (c *Client) PutSymbolExchange(ctx context.Context, inst *reference.Instrument) error {
	span, subCtx := opentracing.StartSpanFromContext(ctx, "redis.PutSymbolExchange")
	defer span.Finish()
	ext.DBType.Set(span, "redis")

	key := symbolExchangeKey(inst.Symbol, inst.Exchange)
	span.LogFields(tlog.String("key", key))

	val, err := msgpack.Marshal(&inst)
	if err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		c.logger.Error("PutSymbolExchange Marshal Error", zap.Error(err))
		return err
	}

	client := otredis.WrapRedisClient(subCtx, c.client)
	if err := client.Set(key, val, 0).Err(); err != nil {
		span.LogFields(tlog.Error(err))
		ext.Error.Set(span, true)
		c.logger.Error("Redis PutSymbolExchange Set Error", zap.Error(err))
		return err
	}
	return nil
}
