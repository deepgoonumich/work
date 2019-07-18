package ftp

import (
	"bytes"
	"context"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/jlaffaye/ftp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"

	"gitlab.benzinga.io/benzinga/ftp-engine/config"
	"gitlab.benzinga.io/benzinga/ftp-engine/process"
	"gitlab.benzinga.io/benzinga/ftp-engine/sender"
)

type Sender struct {
	sync.Mutex
	cfg   *config.Config
	log   *zap.Logger
	conn  *ftp.ServerConn
	retry *retrier.Retrier
}

// re: sync.Mutex, some FTP servers do not allow sending commands via control channel with transfer in progress on same connection
// so we will delay status checks until transfers complete

var _ = sender.Sender(&Sender{}) // check interface

const testFilename = ".bztest"

func NewFTPSender(cfg *config.Config, logger *zap.Logger) (*Sender, error) {

	s := Sender{
		log: logger.Named("ftp"),
		cfg: cfg,
	}

	// Configure Retrier
	if cfg.FTP.SendRetires > 0 {
		s.retry = retrier.New(retrier.ExponentialBackoff(cfg.FTP.SendRetires, 500*time.Millisecond), nil)
	}

	// Start New Connection
	if err := s.connect(); err != nil {
		return nil, err
	}

	// Login
	if err := s.login(); err != nil {
		if quitErr := s.conn.Quit(); quitErr != nil {
			logger.Error("Quit Connection Error", zap.Error(err))
		}
		return nil, err
	}

	// Check Path Writeable
	if err := s.checkPath(); err != nil {
		if quitErr := s.conn.Quit(); quitErr != nil {
			logger.Error("Quit Connection Error", zap.Error(quitErr))
		}
		return nil, err
	}

	// Start keepalive if configured
	if s.cfg.FTP.KeepAliveInterval != 0 {
		go s.startKeepAlive()
	} else {
		s.log.Info("No FTP keepalive configured")
	}

	return &s, nil
}

func (s *Sender) Send(ctx context.Context, data *process.Output) error {

	span, subCtx := opentracing.StartSpanFromContext(ctx, "FTP Send")
	ext.PeerService.Set(span, "ftp")
	ext.PeerAddress.Set(span, s.cfg.FTP.Host)
	span.LogFields(otlog.String("file.name", data.Filename), otlog.String("ftp.username", s.cfg.FTP.Username))
	defer span.Finish()

	// Use Retrier if configured
	s.Lock()
	defer s.Unlock()
	if s.retry != nil {

		err := s.retry.RunCtx(subCtx, func(ctx context.Context) error {
			if storErr := s.conn.Stor(data.Filename, data.Data); storErr != nil {
				s.log.Error("FTP Write Error, will retry.", zap.String("filename", data.Filename))
				span.LogFields(otlog.Error(storErr))

				if strings.Contains(storErr.Error(), "broken pipe") {
					if reconnectErr := s.reconnect(); reconnectErr != nil {
						s.log.Fatal("FTP reconnect error", zap.Error(reconnectErr))
					}
				}

				return storErr
			}
			return nil
		})
		if err != nil {
			s.log.Error("FTP Write Error, retries exceeded", zap.String("filename", data.Filename))
			span.LogFields(otlog.Error(err))
			return err
		}

	} else if err := s.conn.Stor(data.Filename, data.Data); err != nil {
		s.log.Error("FTP Write Error", zap.String("filename", data.Filename))
		span.LogFields(otlog.Error(err))

		if strings.Contains(err.Error(), "broken pipe") {
			if reconnectErr := s.reconnect(); reconnectErr != nil {
				s.log.Fatal("FTP reconnect error", zap.Error(reconnectErr))
			}
		}

		return err
	}

	s.log.Info("FTP Write Success", zap.String("host", s.cfg.FTP.Host), zap.String("filename", data.Filename))
	return nil
}

// Check Path ensures the path given is a writeable directory` by creating then removing a test file,
// there is no error if the test file cannot be deleted
func (s *Sender) checkPath() error {
	if err := s.conn.ChangeDir(s.cfg.FTP.Path); err != nil {
		s.log.Error("Change Directory Error", zap.Error(err))
		return err
	}
	if err := s.conn.Stor(testFilename, bytes.NewBufferString("testing")); err != nil {
		s.log.Error("Create Test File Error", zap.Error(err), zap.String("filepath", path.Join(s.cfg.FTP.Path, testFilename)))
		return err
	}
	if err := s.conn.Delete(testFilename); err != nil {
		s.log.Error("Remove Test File Error", zap.Error(err), zap.String("filepath", path.Join(s.cfg.FTP.Path, testFilename)))
	}
	return nil
}

func (s *Sender) startKeepAlive() {
	ticker := time.NewTicker(s.cfg.FTP.KeepAliveInterval)
	s.log.Info("Starting Period FTP keepalive", zap.Duration("interval", s.cfg.FTP.KeepAliveInterval))

	//nolint linter (gosimple) complains about for loop w/select, but complains about range implementation
	// since we don't need the value from the ticker
	for {
		select {
		case <-ticker.C:
			s.Lock()
			if err := s.conn.NoOp(); err != nil {
				s.log.Error("keepalive NoOp Error", zap.Error(err))
			} else {
				s.log.Debug("keepalive NoOp Success")
			}
			s.Unlock()
		}
	}

}

func (s *Sender) login() error {
	if err := s.conn.Login(s.cfg.FTP.Username, s.cfg.FTP.Password); err != nil {
		s.log.Error("Login Error", zap.Error(err), zap.String("ftp_username", s.cfg.FTP.Username))
		return err
	}
	return nil
}

func (s *Sender) Status() error {

	s.Lock()
	defer s.Unlock()

	if err := s.conn.NoOp(); err != nil {
		s.log.Error("FTP NoOp Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *Sender) Close() error {
	if err := s.conn.Logout(); err != nil {
		s.log.Error("Logout Error", zap.Error(err))
	}
	return s.disconnect()
}

func (s *Sender) reconnect() error {
	if err := s.disconnect(); err != nil {
		s.log.Error("Reconnect - Disconnect - Error", zap.Error(err))
	}
	if err := s.connect(); err != nil {
		s.log.Error("Reconnect - Connect - Error", zap.Error(err))
		return err
	}
	if err := s.login(); err != nil {
		s.log.Error("Reconnect - Login - Error", zap.Error(err))
		return err
	}
	return nil
}

func (s *Sender) connect() error {
	s.log.Info("Connecting FTP Client", zap.String("addr", s.cfg.FTP.Host))

	ftpConn, err := ftp.DialTimeout(s.cfg.FTP.Host, 5*time.Second) // TODO(darwin): use timeout from config
	if err != nil {
		return err
	}

	s.conn = ftpConn

	return nil
}

func (s *Sender) disconnect() error {
	s.log.Info("Disconnecting")
	return s.conn.Quit()
}
