package bzkaf

import (
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/xid"
)

const (
	ContentModelsEventMsgType MessageType = "content_models_event"
	FTPDelivery               MessageType = "ftp_delivery"
)

// MessageType ...
type MessageType string

// String ...
func (m MessageType) String() string {
	return string(m)
}

// Envelope ...
type Envelope struct {
	Trace       opentracing.TextMapCarrier `json:"trace,omitempty"`
	ID          string                     `json:"id"`
	MessageType MessageType                `json:"message_type,omitempty"`
	Message     json.RawMessage            `json:"message"`
}

// json.RawMessage is used because it allows us to delay unmarshaling
// of Message until after other fields can be processed.

// NewEnvelope ...
func NewEnvelope(msgType MessageType, msg json.RawMessage) *Envelope {
	return &Envelope{
		ID:          xid.New().String(),
		MessageType: msgType,
		Message:     msg,
		Trace:       opentracing.TextMapCarrier{},
	}
}

// Marshal ...
func (e *Envelope) Marshal() ([]byte, error) {
	return jsoniter.Marshal(e)
}
