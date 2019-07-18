package rstore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.benzinga.io/benzinga/ftp-engine/config"
)

func TestStatus(t *testing.T) {
	// Load Config
	cfg, err := config.LoadConfig("test")
	require.NoError(t, err)

	logger, err := cfg.LoadLogger()
	require.NoError(t, err)

	c, err := NewClient(logger, cfg.RedisURL)
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, c.Status(ctx))
}
