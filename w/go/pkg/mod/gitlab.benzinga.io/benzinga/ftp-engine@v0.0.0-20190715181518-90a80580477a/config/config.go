package config

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.benzinga.io/benzinga/content-models/models"
)

type Config struct {
	AppName string `validate:"required"`
	// AppBuild Git SHA[0:8] of current release. The value is injected into build pipeline.
	AppBuild   string          `validate:"required"`
	AppEnv     AppEnv          `validate:"required"`
	ListenHost string          `validate:"required"`
	ListenPort string          `validate:"required"`
	Debug      bool            `validate:"required"`
	RedisURL   string          `validate:"required"`
	Processor  ProcessorConfig `validate:"required"`
	Kafka      KafkaConfig     `validate:"required"`
	FTP        FTPConfig       `validate:"required"`
}

type ProcessorConfig struct {
	Type                ProcessorType      `validate:"required"`
	AcceptedEvents      []models.EventType `validate:"required"`
	IgnoreUpdatedBefore *time.Time
}

type FTPConfig struct {
	Host              string `validate:"required"`
	Username          string
	Password          string
	Path              string        `validate:"required"`
	ConnTimeout       time.Duration `validate:"required"`
	KeepAliveInterval time.Duration
	SendRetires       int `validate:"required"`
}

type KafkaConfig struct {
	Brokers []string `validate:"required"`
	Topic   string   `validate:"required"`
	// GroupID should be unique for output client, if there are multiple instances for a single receiver, this should be the same for each
	GroupID     string `validate:"required"`
	Username    string
	Password    string
	TLSKeyPath  string
	TLSCertPath string
	TLSCAPath   string
}

const AppName = "ftp-engine"

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

// ProcessorType indicates formater/Processor to use for output
type ProcessorType string

const (
	// RavenpackProcessor ...
	RavenpackProcessor = "ravenpack"
	// DefaultProcessor ...
	DefaultProcessor = "default"
)

// String returns ProcessorType as string
func (p ProcessorType) String() string {
	return string(p)
}

func LoadConfig(appBuild string) (*Config, error) {

	// Load Config from Env
	v := viper.New()
	v.AutomaticEnv()
	v.AddConfigPath("config")
	v.AddConfigPath("../config")
	v.AddConfigPath("../../config")
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

	// Determine Processor
	processor := strings.ToLower(v.GetString("PROCESSOR"))
	var processorType ProcessorType
	switch processor {
	case RavenpackProcessor:
		processorType = RavenpackProcessor
	case DefaultProcessor:
		processorType = DefaultProcessor
	default:
		return nil, errors.New("invalid processor specified")
	}

	// Determine Accepted Events
	eventsSelection := strings.Split(v.GetString("PROCESSOR_EVENTS"), ",")
	var processorEvents []models.EventType
	for _, v := range eventsSelection {
		switch strings.ToLower(v) {
		case strings.ToLower(string(models.Created)):
			processorEvents = append(processorEvents, models.Created)
		case strings.ToLower(string(models.Updated)):
			processorEvents = append(processorEvents, models.Updated)
		case strings.ToLower(string(models.Removed)):
			processorEvents = append(processorEvents, models.Removed)
		default:
			return nil, fmt.Errorf("invalid processor event type '%s'", v)
		}
	}

	c := Config{
		AppName:    AppName,
		AppBuild:   appBuild,
		AppEnv:     runEnv,
		Debug:      v.GetBool("DEBUG"),
		ListenPort: v.GetString("LISTEN_PORT"),
		ListenHost: v.GetString("LISTEN_HOST"),
		RedisURL:   v.GetString("REDIS_URL"),
		Processor: ProcessorConfig{
			Type:           processorType,
			AcceptedEvents: processorEvents,
		},
		FTP: FTPConfig{
			Host:              v.GetString("FTP_HOST"),
			Path:              v.GetString("FTP_PATH"),
			Username:          v.GetString("FTP_USERNAME"),
			Password:          v.GetString("FTP_PASSWORD"),
			ConnTimeout:       v.GetDuration("FTP_CONNECT_TIMEOUT"),
			KeepAliveInterval: v.GetDuration("FTP_KEEPALIVE_INTERVAL"),
			SendRetires:       v.GetInt("FTP_SEND_RETRIES"),
		},
		Kafka: KafkaConfig{
			Brokers:     strings.Split(v.GetString("KAFKA_BROKERS"), ","),
			Topic:       v.GetString("KAFKA_TOPIC"),
			GroupID:     v.GetString("KAFKA_GROUP_ID"),
			Username:    v.GetString("KAFKA_USERNAME"),
			Password:    v.GetString("KAFKA_PASSWORD"),
			TLSCAPath:   v.GetString("KAFKA_TLS_CA"),
			TLSCertPath: v.GetString("KAFKA_TLS_CERT"),
			TLSKeyPath:  v.GetString("KAFKA_TLS_KEY"),
		},
	}

	// Set Ignore Updated Before, this setting tells worker to ignore content with an `UpdatedAt` before this time
	if v := v.GetString("IGNORE_UPDATED_BEFORE"); v != "" {
		ignoreBefore, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		c.Processor.IgnoreUpdatedBefore = &ignoreBefore
	}

	// Validate Config
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return nil, fmt.Errorf("config validation failed: %s", err)
	}

	return &c, nil
}

func (c *Config) ListenAPI() string {
	return c.ListenHost + ":" + c.ListenPort
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
