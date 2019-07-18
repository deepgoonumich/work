package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tlog "github.com/opentracing/opentracing-go/log"
	"github.com/spf13/viper"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.benzinga.io/benzinga/ftp-engine/rstore"
	"gitlab.benzinga.io/benzinga/reference-service/reference"
)

const AppName = "ftp-engine-updater"

// AppEnv
type AppEnv string

const (
	// ProductionEnv ...
	ProductionEnv AppEnv = "production"
	// DevelopmentEnv ...
	DevelopmentEnv AppEnv = "development"
	// StagingEnv ...
	StagingEnv AppEnv = "staging"
	// TestingEnv
	TestingEnv AppEnv = "testing"
)

func (e AppEnv) String() string {
	return string(e)
}

type Config struct {
	AppName string `validate:"required"`
	// AppBuild Git SHA[0:8] of current release. The value is injected into build pipeline.
	AppBuild       string        `validate:"required"`
	AppEnv         AppEnv        `validate:"required"`
	Debug          bool          `validate:"required"`
	UpdateInterval time.Duration `validate:"required"`
	RefDBEndpoint  string        `validate:"required"`
	RedisURL       string        `validate:"required"`
}

func (c *Config) LoadLogger() (*zap.Logger, error) {
	if c.AppEnv == DevelopmentEnv || c.AppEnv == TestingEnv {
		return zap.NewDevelopment()
	}

	if c.Debug {
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return cfg.Build()
	}

	return zap.NewProduction()
}

func (c *Config) LoadTracer() (opentracing.Tracer, io.Closer, error) {

	cfg, err := jaegerconfig.FromEnv()
	if err != nil {
		return nil, nil, err
	}
	cfg.Tags = append(cfg.Tags, opentracing.Tag{Key: "build", Value: c.AppBuild}, opentracing.Tag{Key: "environment", Value: c.AppEnv.String()})
	cfg.ServiceName = c.AppName

	if c.AppEnv == DevelopmentEnv || c.AppEnv == TestingEnv {
		cfg.Reporter.LogSpans = true
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}

	return tracer, closer, nil
}

func LoadConfig(appBuild string) (*Config, error) {

	// Load Config from Env
	v := viper.New()
	v.AutomaticEnv()
	v.AddConfigPath("config")
	v.AddConfigPath("../config")
	v.AddConfigPath("../../config")
	v.AddConfigPath("../../../config")
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Config file not found, using ENV vars. Note: This is expected behavior. Msg: %s", err)
	}

	// Determine Runtime Environment
	env := strings.ToLower(v.GetString("ENVIRONMENT"))
	var runEnv AppEnv
	switch env {
	case ProductionEnv.String():
		runEnv = ProductionEnv
	case DevelopmentEnv.String():
		runEnv = DevelopmentEnv
	case StagingEnv.String():
		runEnv = StagingEnv
	default:
		runEnv = TestingEnv
		log.Println("invalid or testing environment specified, using testing environment")
	}

	c := Config{
		AppName:        AppName,
		AppBuild:       appBuild,
		AppEnv:         runEnv,
		Debug:          v.GetBool("DEBUG"),
		RedisURL:       v.GetString("REDIS_URL"),
		UpdateInterval: v.GetDuration("REFDB_UPDATE_INTERVAL"),
		RefDBEndpoint:  v.GetString("REFDB_ENDPOINT"),
	}

	// Validate Config
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return nil, fmt.Errorf("config validation failed: %s", err)
	}

	return &c, nil
}

var build string // 0:8 GIT SHA injected at build time in Dockerfile

func main() {
	buildString := func() string {
		if build != "" {
			return build
		}
		return "testing-unset"
	}()

	cfg, err := LoadConfig(buildString)
	if err != nil {
		log.Fatalln("Load Config Error", err)
	}

	// Load Logger
	logger, err := cfg.LoadLogger()
	if err != nil {
		log.Fatalln("Load Logger Error", err)
	}

	logger = logger.With(zap.String("build", buildString), zap.String("environment", cfg.AppEnv.String()))
	logger.Info("Initializing")

	// Load Tracing
	tracer, closer, err := cfg.LoadTracer()
	if err != nil {
		// handle the case where the work failed three times
		logger.Fatal("Load Tracing Failed", zap.Error(err))
	}
	// Set Global Tracer
	opentracing.SetGlobalTracer(tracer)

	// Cancel Context
	ctx, cancel := context.WithCancel(context.Background())

	// Do Initial Load
	if err := refreshTickers(ctx, logger, cfg); err != nil {
		logger.Fatal("Unable to do Inital Ticker Loading", zap.Error(err))
	}

	// Start Refresh Worker
	go refreshWorker(ctx, logger, cfg)

	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	closer.Close()
	if syncErr := logger.Sync(); syncErr != nil {
		log.Println("Log Sync Error", syncErr)
	}

	logger.Warn("Shutdown Signal Received")

	cancel()

	close(quit)
}

func refreshWorker(ctx context.Context, logger *zap.Logger, cfg *Config) {

	ticker := time.NewTicker(cfg.UpdateInterval)

	logger.Info("Starting Refresh Worker", zap.Duration("refresh_interval", cfg.UpdateInterval))

	for {
		select {
		case <-ticker.C:
			logger.Debug("Starting Update")

			r := retrier.New(retrier.ExponentialBackoff(3, 1*time.Minute), nil)

			err := r.RunCtx(ctx, func(subCtx context.Context) error {
				if refreshErr := refreshTickers(subCtx, logger, cfg); refreshErr != nil {
					logger.Error("Refresh Worker Error", zap.Error(refreshErr))
					return refreshErr
				}
				return nil
			})
			if err != nil {
				logger.Fatal("Update Failed After Retries", zap.Error(err))
			}
			logger.Debug("Refresh Successful")

		case <-ctx.Done():
			logger.Info("Stopping Refresh Worker")
			return
		}
	}
}

func refreshTickers(ctx context.Context, logger *zap.Logger, cfg *Config) error {
	span, subCtx := opentracing.StartSpanFromContext(ctx, "updater.refreshTickers")
	defer span.Finish()

	ext.HTTPUrl.Set(span, cfg.RefDBEndpoint)
	ext.HTTPMethod.Set(span, http.MethodGet)

	start := time.Now()

	resp, err := http.Get(cfg.RefDBEndpoint)
	if err != nil {
		span.LogFields(tlog.Error(err))
		logger.Error("Get refDB instruments error", zap.String("endpoint", cfg.RefDBEndpoint))
		return err
	}

	logger.Debug("refDb Downloaded", zap.Duration("elapsed", time.Since(start)))
	ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("Response Body Close Error", zap.Error(closeErr))
		}
	}()

	var data reference.FinancialData

	if unmarshalErr := jsoniter.NewDecoder(resp.Body).Decode(&data); unmarshalErr != nil {
		span.LogFields(tlog.Error(err))
		logger.Error("Decode Instruments JSON Error", zap.Error(unmarshalErr))
		return unmarshalErr
	}
	logger.Debug("Instruments JSON Decoded")

	rClient, err := rstore.NewClient(logger, cfg.RedisURL)
	if err != nil {
		span.LogFields(tlog.Error(err))
		logger.Error("rstore new client error", zap.Error(err))
		return err
	}

	for i := 0; i < len(data.Instruments); i++ {
		if err := rClient.PutSymbolCurrency(subCtx, &data.Instruments[i]); err != nil {
			span.LogFields(tlog.Error(err))
			logger.Error("refresh PutSymbolCurrency error", zap.Error(err))
			return err
		}
		if err := rClient.PutSymbolExchange(subCtx, &data.Instruments[i]); err != nil {
			span.LogFields(tlog.Error(err))
			logger.Error("refresh PutSymbolExchange error", zap.Error(err))
			return err
		}
	}

	logger.Info("Updated Tickers", zap.Int("count", len(data.Instruments)), zap.Duration("total_duration", time.Since(start)))

	return nil
}
