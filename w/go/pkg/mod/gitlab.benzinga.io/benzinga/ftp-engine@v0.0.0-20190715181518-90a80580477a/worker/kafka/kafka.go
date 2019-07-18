package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	_ "github.com/segmentio/kafka-go/gzip" // needed for decompression
	"github.com/segmentio/kafka-go/lz4"    // needed for decompression
	"github.com/segmentio/kafka-go/sasl/scram"
	_ "github.com/segmentio/kafka-go/zstd" // needed for decompression
	"go.uber.org/zap"

	"gitlab.benzinga.io/benzinga/bzkaf"
	"gitlab.benzinga.io/benzinga/content-models/models"
	"gitlab.benzinga.io/benzinga/ftp-engine/config"
	"gitlab.benzinga.io/benzinga/ftp-engine/instr"
	"gitlab.benzinga.io/benzinga/ftp-engine/process"
	"gitlab.benzinga.io/benzinga/ftp-engine/sender"
	"gitlab.benzinga.io/benzinga/ftp-engine/worker"
)

var _ = worker.Worker(&Worker{}) // check interface

type Worker struct {
	log       *zap.Logger
	cfg       *config.Config
	instr     *instr.Collector
	reader    *kafka.Reader
	writer    *kafka.Writer
	processor process.Processor
	sender    sender.Sender
}

func loadTLSConfig(keyPath, certPath, caPath string) (*tls.Config, error) {

	var tlsConfig tls.Config

	if keyPath != "" && certPath != "" {
		cer, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}

		tlsConfig.Certificates = []tls.Certificate{cer}
	}

	// Use Custom CA if given
	if caPath != "" {
		caCertBytes, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(caCertBytes); !ok {
			return nil, fmt.Errorf("unable to append CA bytes")
		}

		tlsConfig.RootCAs = pool
	}

	return &tlsConfig, nil
}

func NewKafkaWorker(cfg *config.Config, logger *zap.Logger, inst *instr.Collector, s sender.Sender, p process.Processor) (*Worker, error) {

	readerConfig := kafka.ReaderConfig{
		Brokers:               cfg.Kafka.Brokers,
		Topic:                 cfg.Kafka.Topic,
		GroupID:               cfg.Kafka.GroupID,
		WatchPartitionChanges: true,
		// RetentionTime optionally sets the length of time the consumer group will be saved
		// by the broker
		// Only used when GroupID is set
		RetentionTime:     time.Hour * 168, // one week
		HeartbeatInterval: time.Second * 2,
		MaxWait:           time.Second * 10,
		MinBytes:          10e3, // 10KB
		MaxBytes:          10e8, // 100MB
	}

	writerConfig := kafka.WriterConfig{
		Brokers:          cfg.Kafka.Brokers,
		Topic:            "third-party-deliveries",
		CompressionCodec: lz4.NewCompressionCodec(),
	}

	// If Username and Password set use scram auth
	if cfg.Kafka.Username != "" && cfg.Kafka.Password != "" {
		algo := scram.SHA256
		mech, err := scram.Mechanism(algo, cfg.Kafka.Username, cfg.Kafka.Password)
		if err != nil {
			return nil, err
		}
		writerConfig.Dialer = &kafka.Dialer{
			SASLMechanism: mech,
		}
		readerConfig.Dialer = &kafka.Dialer{
			SASLMechanism: mech,
		}
		logger.Info("Kafka Username & Password set, attempting connection with scram auth", zap.String("algo", algo.Name()))
	} else {
		logger.Info("Kafka Username & Password unset, attempting connection without scram auth")
	}

	// If CA Cert Path or TLS Cert & TLS Key paths given use custom TLS config
	if cfg.Kafka.TLSCAPath != "" || cfg.Kafka.TLSCertPath != "" && cfg.Kafka.TLSKeyPath != "" {
		logger.Debug("Loading Kafka TLS")
		tlsConfig, err := loadTLSConfig(cfg.Kafka.TLSKeyPath, cfg.Kafka.TLSCertPath, cfg.Kafka.TLSCAPath)
		if err != nil {
			logger.Error("Error Loading Kafka TLS", zap.Error(err))
			return nil, err
		}

		if readerConfig.Dialer == nil && writerConfig.Dialer == nil {
			readerConfig.Dialer = &kafka.Dialer{
				TLS: tlsConfig,
			}
			writerConfig.Dialer = &kafka.Dialer{
				TLS: tlsConfig,
			}
		} else {
			readerConfig.Dialer.TLS = tlsConfig
			writerConfig.Dialer.TLS = tlsConfig
		}

		logger.Info("Kafka TLS connection configured")
	} else {
		logger.Info("Kafka TLS connection not configured")
	}

	if err := readerConfig.Validate(); err != nil {
		return nil, err
	}

	if err := writerConfig.Validate(); err != nil {
		return nil, err
	}

	r := kafka.NewReader(readerConfig)
	w := kafka.NewWriter(writerConfig)

	return &Worker{logger.Named("worker:kafka"), cfg, inst, r, w, p, s}, nil
}

func (w *Worker) Disconnect() (err error) {
	err = w.reader.Close()
	if err != nil {
		w.log.Error("Reader Close Error", zap.Error(err))
	}
	err = w.writer.Close()
	if err != nil {
		w.log.Error("Writer Close Error", zap.Error(err))
	}
	return err
}

func (w *Worker) Work(ctx context.Context) {

	var workStart time.Time

work:
	for {
		select {
		default:
			w.log.Debug("Waiting for new content from Kafka")

			// Fetch Message
			msg, err := w.reader.FetchMessage(ctx)
			span, subCtx := opentracing.StartSpanFromContext(ctx, "New Kafka Message")
			if err != nil {

				span.LogFields(otlog.Error(err))

				if err == io.EOF {
					w.log.Fatal("Kafka Reader Closed", zap.Error(err))
				}

				w.log.Error("Fetch Message Error", zap.Error(err))
				w.instr.ContentReceiveErrors.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic}).Inc()
				span.Finish()
				continue work
			}

			workStart = time.Now()

			w.instr.ContentAccepted.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic}).Inc()
			msgLog := w.log.With(zap.Int64("offset", msg.Offset), zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Time("msg_time", msg.Time), zap.String("kafka_group_id", w.cfg.Kafka.GroupID))
			span.LogFields(otlog.Int64("offset", msg.Offset), otlog.String("topic", msg.Topic), otlog.Int("partition", msg.Partition), otlog.String("group_id", w.cfg.Kafka.GroupID))
			msgLog.Debug("Kafka Message Received")

			//
			// Process Message
			//

			// Unmarshal Kafka Envelope
			var envelope bzkaf.Envelope
			if err := json.Unmarshal(msg.Value, &envelope); err != nil {
				span.LogFields(otlog.Error(err))
				msgLog.Error("Unmarshal Kafka Envelope Error", zap.Error(err))
				w.instr.ContentReceiveErrors.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic}).Inc()
				span.Finish()
				continue work
			}
			msgLog = msgLog.With(zap.String("envelope_id", envelope.ID))
			msgLog.Debug("Kafka Message Envelope Unmarshaled")

			if envelope.MessageType != bzkaf.ContentModelsEventMsgType {
				// This should never happen unless there is an issue with Kafka configuration
				w.instr.ContentRejected.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic, "reason": "invalid_evenlope_message_type"}).Inc()
				msgLog.Error("Invalid Message Type", zap.String("message_type", envelope.MessageType.String()))
				w.commitMessages(subCtx, msgLog, workStart, msg)
				span.Finish()
				continue work
			}

			// ToDo(darwin): Unpack/use tracing data

			// Unmarshal Event
			var event models.Event
			if err := json.Unmarshal(envelope.Message, &event); err != nil {
				span.LogFields(otlog.Error(err))
				w.instr.ContentReceiveErrors.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic}).Inc()
				msgLog.Error("Unmarshal Kafka Envelope Error", zap.Error(err))
				span.Finish()
				continue work
			}

			span.LogFields(otlog.Int64("event_id", event.ID), otlog.Int64("node_id", event.NodeID), otlog.String("envelope_id", envelope.ID), otlog.String("event_type", string(event.Event)))
			msgLog = msgLog.With(zap.Int64("event_id", event.ID), zap.Int64("node_id", event.NodeID))
			msgLog.Debug("Event Unmarshaled")

			// Filter Event
			// Check Event Updated Not Before Ignore Value
			if w.cfg.Processor.IgnoreUpdatedBefore != nil && w.cfg.Processor.IgnoreUpdatedBefore.Before(event.Content.UpdatedAt.Time) {
				w.instr.ContentRejected.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic, "reason": "updated_before_ignore_value"}).Inc()
				w.commitMessages(subCtx, msgLog, workStart, msg)
				span.LogFields(otlog.String("updated_at", event.Content.UpdatedAt.Time.String()), otlog.String("ignore_before", w.cfg.Processor.IgnoreUpdatedBefore.String()))
				msgLog.Info("Ignoring Event, is updated before ignore value", zap.Time("updated_at", event.Content.UpdatedAt.Time), zap.Time("ignore_before", *w.cfg.Processor.IgnoreUpdatedBefore))
				span.Finish()
				continue work
			}
			// Check Event Content Type
			if contentType, ok := process.ContentTypeMappings[strings.ToLower(event.Content.Type)]; !ok {
				w.instr.ContentRejected.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic, "reason": "unwanted_content_type"}).Inc()
				w.commitMessages(subCtx, msgLog, workStart, msg)
				span.LogFields(otlog.String("content_type", event.Content.Type))
				msgLog.Info("Ignoring Event, is not wanted content type", zap.String("content_type", event.Content.Type))
				span.Finish()
				continue work
			} else {
				msgLog.Debug("Content Type Valid", zap.String("content_type", contentType.String()))
			}
			// Check Event is of Accepted Event Type
			var match bool
			for i := 0; i < len(w.cfg.Processor.AcceptedEvents); i++ {

				if event.Event == w.cfg.Processor.AcceptedEvents[i] {
					// Send Message if Event is of Accepted type
					match = true
					if err := w.processAndSend(subCtx, &event); err != nil {
						span.LogFields(otlog.Error(err))
						msgLog.Error("Processor/Send Error", zap.Error(err))
						w.instr.ContentSendErrors.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic}).Inc()
						span.Finish()
						continue work
					}
					msgLog.Debug("Content Sent")

				}
			}

			if !match {
				// Unaccepted event type, acknowledged, but was not sent
				msgLog.Debug("Unaccepted Event Type", zap.String("event_type", string(event.Event)), zap.Time("event_content_updated_at", event.Content.UpdatedAt.Time))
				w.instr.ContentRejected.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic, "reason": "unwanted_event_type"}).Inc()
			} else {
				w.instr.ContentSent.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic}).Inc()
			}
			span.LogFields(otlog.Bool("is_accepted_event_type", match), otlog.String("event_content_updated_at", event.Content.UpdatedAt.String()))

			// Acknowledge/Commit
			w.commitMessages(subCtx, msgLog, workStart, msg)
			w.instr.ContentProcessingLatency.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic}).Observe(time.Since(workStart).Seconds())
			span.Finish()

		case <-ctx.Done():
			w.log.Info("Receiver Context Done, disconnecting.")
			if err := w.Disconnect(); err != nil {
				w.log.Error("Close Kafka Reader Connection Error", zap.Error(err))
			}
			break work
		}

	}

}

func (w *Worker) recordFTPDelivery(ctx context.Context, o *process.Output, event *models.Event) error {

	record := worker.FTPDeliveryRecord{
		NodeID:          event.NodeID,
		EventID:         event.ID,
		EventType:       event.Event,
		ConsumerGroupID: w.cfg.Kafka.GroupID,
		FTPHost:         w.cfg.FTP.Host,
		FTPPath:         w.cfg.FTP.Path,
		Filename:        o.Filename,
		SHA256Checksum:  o.Checksum,
		Timestamp:       time.Now().UTC(),
		SizeBytes:       o.Size,
	}

	// Marshal Record
	recordJSON, err := jsoniter.Marshal(&record)
	if err != nil {
		w.log.Error("Marshal FTP Delivery Record Error", zap.Error(err))
		return err
	}

	// New Envelope
	envelope := bzkaf.NewEnvelope(bzkaf.FTPDelivery, recordJSON)
	envelopeJSON, err := envelope.Marshal()
	if err != nil {
		w.log.Error("Envelope FTP Delivery Confirmation Marshal Error", zap.Error(err))
		return err
	}

	// Create Message
	msg := kafka.Message{
		Value: envelopeJSON,
	}

	// Send FTP Delivery Confirmation
	if err := w.writer.WriteMessages(ctx, msg); err != nil {
		w.log.Error("Kafka Write FTP Delivery Confirmation Error")
		return err
	}

	w.log.Debug("Delivery Confirmation Sent", zap.String("envelope.id", envelope.ID), zap.String("data.checksum", o.Checksum), zap.Int("data.size", o.Size))

	return nil
}

func (w *Worker) processAndSend(ctx context.Context, event *models.Event) error {
	output, err := w.processor.Convert(event)
	if err != nil {
		return err
	}
	if err := w.sender.Send(ctx, output); err != nil {
		return err
	}
	if err := w.recordFTPDelivery(ctx, output, event); err != nil {
		w.log.Error("Record FTP Delivery Error", zap.Error(err))
	}
	return nil
}

func (w *Worker) commitMessages(ctx context.Context, msgLog *zap.Logger, start time.Time, msg ...kafka.Message) {
	if err := w.reader.CommitMessages(ctx, msg...); err != nil {
		msgLog.Error("Kafka Commit Error", zap.Error(err))
	}
	w.instr.ContentAcknowledged.With(prometheus.Labels{"kafka_group_id": w.cfg.Kafka.GroupID, "kafka_topic": w.cfg.Kafka.Topic}).Inc()
	msgLog.Debug("Message Commit Success", zap.Duration("total_latency", time.Since(start)))
}
