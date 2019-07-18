package instr

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector contains various prometheus metrics collectors
type Collector struct {
	// ContentAcknowledged ...
	ContentAcknowledged *prometheus.CounterVec
	// ContentReceiveErrors
	ContentReceiveErrors *prometheus.CounterVec
	// ContentRejected ...
	ContentRejected *prometheus.CounterVec
	// ContentAccepted ...
	ContentAccepted *prometheus.CounterVec

	// ContentProcessingLatency ...
	ContentProcessingLatency *prometheus.SummaryVec

	// ContentSent ...
	ContentSent *prometheus.CounterVec
	// ContentSendErrors ...
	ContentSendErrors *prometheus.CounterVec
}

// NewCollector returns initialized prometheus collector
func NewCollector(appName string) (*Collector, error) {
	var collectors []prometheus.Collector

	contentAcknowledged := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strings.Replace(appName, "-", "_", -1),
			Subsystem: "receive",
			Name:      "content_acknowledged",
			Help:      "content objects fetched and acknowledged from queue",
		},
		[]string{"kafka_group_id", "kafka_topic"},
	)
	collectors = append(collectors, contentAcknowledged)

	contentRecvError := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strings.Replace(appName, "-", "_", -1),
			Subsystem: "receive",
			Name:      "content_receive_error",
			Help:      "errors getting content from queue",
		},
		[]string{"kafka_group_id", "kafka_topic"},
	)
	collectors = append(collectors, contentRecvError)

	contentRejected := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strings.Replace(appName, "-", "_", -1),
			Subsystem: "receive",
			Name:      "content_rejected",
			Help:      "content objects rejected from queue",
		},
		[]string{"kafka_group_id", "kafka_topic", "reason"},
	)
	collectors = append(collectors, contentRejected)

	contentAccepted := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strings.Replace(appName, "-", "_", -1),
			Subsystem: "content",
			Name:      "content_accepted",
			Help:      "content objects accepted from adaptor sources",
		},
		[]string{"kafka_group_id", "kafka_topic"},
	)
	collectors = append(collectors, contentAccepted)

	contentSendErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strings.Replace(appName, "-", "_", -1),
			Subsystem: "content",
			Name:      "content_send_errors",
			Help:      "content objects with error on send",
		},
		[]string{"kafka_group_id", "kafka_topic"},
	)
	collectors = append(collectors, contentSendErrors)

	contentSent := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: strings.Replace(appName, "-", "_", -1),
			Subsystem: "content",
			Name:      "content_sent",
			Help:      "content sent successfully",
		},
		[]string{"kafka_group_id", "kafka_topic"},
	)
	collectors = append(collectors, contentSent)

	contentProcessLatency := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: strings.Replace(appName, "-", "_", -1),
			Subsystem: "content",
			Name:      "content_processing_latency",
			Help:      "content objects processing latency, receive to send complete",
		},
		[]string{"kafka_group_id", "kafka_topic"},
	)
	collectors = append(collectors, contentProcessLatency)

	for _, c := range collectors {
		err := prometheus.Register(c)
		if err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
				// A counter for that metric has been registered before.
				// Use the old counter from now on.
				continue
			} else {
				// Something else went wrong!
				panic(err)
			}
		}

	}

	return &Collector{
		ContentAccepted:          contentAccepted,
		ContentRejected:          contentRejected,
		ContentAcknowledged:      contentAcknowledged,
		ContentReceiveErrors:     contentRecvError,
		ContentProcessingLatency: contentProcessLatency,
		ContentSendErrors:        contentSendErrors,
		ContentSent:              contentSent,
	}, nil
}
