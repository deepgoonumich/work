package ftp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	filedriver "github.com/goftp/file-driver"
	"github.com/goftp/server"
	"gitlab.benzinga.io/benzinga/ftp-engine/config"
)

func loadTestSender(t *testing.T) *Sender {
	cfg, err := config.LoadConfig("testing")
	require.NoError(t, err)

	logger, err := cfg.LoadLogger()
	require.NoError(t, err)

	// Start Test FTP Server
	factory := &filedriver.FileDriverFactory{
		RootPath: os.TempDir(),
		Perm:     server.NewSimplePerm("user", "group"),
	}

	opts := &server.ServerOpts{
		Factory:  factory,
		Port:     12345,
		Hostname: "127.0.0.1",
		Auth:     &server.SimpleAuth{Name: cfg.FTP.Username, Password: cfg.FTP.Password},
	}

	ftpServer := server.NewServer(opts)
	go func() {
		require.NoError(t, ftpServer.ListenAndServe())
		assert.NoError(t, ftpServer.Shutdown())
	}()

	cfg.FTP.Host = "localhost:12345"

	s, err := NewFTPSender(cfg, logger)
	require.NoError(t, err)

	return s
}

func TestFTP(t *testing.T) {
	s := loadTestSender(t)
	defer s.Close()

	// Test Status
	assert.NoError(t, s.Status())

	// Test checkPath
	assert.NoError(t, s.checkPath(), "checkPath tests FTP directory writeable")
}
