package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"gitlab.benzinga.io/benzinga/ftp-engine/api"
	"gitlab.benzinga.io/benzinga/ftp-engine/config"
	"gitlab.benzinga.io/benzinga/ftp-engine/instr"
	"gitlab.benzinga.io/benzinga/ftp-engine/process"
	"gitlab.benzinga.io/benzinga/ftp-engine/process/ravenpack"
	"gitlab.benzinga.io/benzinga/ftp-engine/rstore"
	"gitlab.benzinga.io/benzinga/ftp-engine/sender/ftp"
	"gitlab.benzinga.io/benzinga/ftp-engine/worker/kafka"
)

var build string // 0:8 GIT SHA injected at build time in Dockerfile

func main() {

	buildString := func() string {
		if build != "" {
			return build
		}
		return "testing-unset"
	}()

	cfg, err := config.LoadConfig(buildString)
	if err != nil {
		log.Fatalln("Load Config Error", err)
	}

	// Write Logo
	if cfg.AppEnv == config.DevelopmentEnv || cfg.AppEnv == config.TestingEnv {
		os.Stderr.Write(logo)
	}

	// Load Logger
	logger, err := cfg.LoadLogger()
	if err != nil {
		log.Fatalln("Load Logger Error", err)
	}

	logger = logger.With(zap.String("name", cfg.Kafka.GroupID))
	logger.With(zap.String("build", buildString), zap.String("environment", cfg.AppEnv.String()))
	logger.Info("Initializing")

	// Load Prometheus
	inst, err := instr.NewCollector(cfg.AppName)
	if err != nil {
		logger.Fatal("Load Prometheus Collector Error", zap.Error(err))
	}

	// Load Tracing
	tracer, closer, err := cfg.LoadTracer()
	if err != nil {
		// handle the case where the work failed three times
		logger.Fatal("Load Tracing Failed", zap.Error(err))
	}
	// Set Global Tracer
	opentracing.SetGlobalTracer(tracer)

	// Load Sender
	sender, err := ftp.NewFTPSender(cfg, logger)
	if err != nil {
		logger.Fatal("Load FTP Sender Error", zap.Error(err))
	}

	router := api.LoadRoutes(cfg, logger, sender)
	logger.Info("Starting HTTP Server", zap.String("listen", cfg.ListenAPI()))
	// Start API Server
	srv := &http.Server{
		Addr:    cfg.ListenAPI(),
		Handler: router,
	}
	go func() {
		// serve connections
		if srvErr := srv.ListenAndServe(); srvErr != nil && srvErr != http.ErrServerClosed {
			logger.Fatal("Server Error", zap.Error(srvErr))
		}
	}()

	defer func() {
		closer.Close()
		if closerErr := sender.Close(); closerErr != nil {
			logger.Error("Sender Close Error", zap.Error(closerErr))
		}
		if syncErr := logger.Sync(); syncErr != nil {
			log.Println("Log Sync Error", syncErr)
		}
	}()

	// Cancel Context
	ctx, cancel := context.WithCancel(context.Background())

	// Load Redis
	logger.Info("Loading Redis")
	rClient, err := rstore.NewClient(logger, cfg.RedisURL)
	if err != nil {
		logger.Fatal("Load Redis Error", zap.Error(err))
	}

	// Load Processor
	var processor process.Processor
	switch cfg.Processor.Type {
	case config.RavenpackProcessor:
		processor = ravenpack.NewRavenpackProcessor(cfg, rClient, logger)
	default:
		logger.Fatal("Unsupported Processor Type", zap.Stringer("type", cfg.Processor.Type))
	}

	// Init Worker
	logger.Info("Initializing Kafka Worker",
		zap.Strings("brokers", cfg.Kafka.Brokers),
		zap.String("topic", cfg.Kafka.Topic),
		zap.String("group_id", cfg.Kafka.GroupID))

	kw, err := kafka.NewKafkaWorker(cfg, logger, inst, sender, processor)
	if err != nil {
		logger.Fatal("Load Kafka Worker Error", zap.Error(err))
	}

	// Start Worker
	logger.Info("Starting Worker")

	go kw.Work(ctx)

	logger.Info("Worker Started")

	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	subCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(subCtx); err != nil {
		logger.Error("Server Shutdown Error", zap.Error(err))
	}

	logger.Warn("Shutdown Signal Received")

	cancel()

	close(quit)

}

var logo = []byte(`[48;5;248m[38;5;16m▄[48;5;255m[38;5;231m▄▄▄▄▄▄[48;5;250m▄[48;5;233m▄[48;5;16m [48;5;244m [48;5;231m       [48;5;232m [49m
[48;5;16m [48;5;231m  [48;5;16m   [48;5;232m[38;5;234m▄[48;5;231m  [48;5;16m     [38;5;254m▄[48;5;231m [38;5;255m▄[48;5;249m[38;5;16m▄[48;5;16m [49m  [39;49mBENZINGA Amp Engine
[48;5;16m [48;5;231m    [38;5;255m▄  [48;5;234m[38;5;102m▄[48;5;16m   [38;5;236m▄[48;5;245m[38;5;231m▄[48;5;231m [38;5;59m▄[48;5;233m[38;5;16m▄[48;5;16m  [49m  [39;49mCopyright 2019 BENZINGA
[48;5;16m [48;5;231m  [48;5;16m    [48;5;231m  [48;5;234m[38;5;243m▄[48;5;16m [38;5;188m▄[48;5;231m [38;5;254m▄[48;5;250m[38;5;16m▄[48;5;16m    [49m  [39;49mFOR INTERNAL USE ONLY.
[48;5;16m [48;5;231m  [48;5;255m[38;5;231m▄▄▄[48;5;231m [38;5;255m▄[38;5;239m▄[48;5;16m [48;5;242m[38;5;244m▄[48;5;231m  [48;5;255m[38;5;231m▄▄▄▄[48;5;231m [48;5;237m [39;49m
`)
