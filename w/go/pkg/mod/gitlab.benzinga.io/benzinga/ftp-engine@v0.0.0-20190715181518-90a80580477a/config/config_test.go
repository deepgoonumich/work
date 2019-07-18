package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBuild = "testing"

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig(testBuild)
	require.NoError(t, err)
	require.NotNil(t, cfg)

}

func TestListenAPI(t *testing.T) {
	cfg, err := LoadConfig(testBuild)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.NotEmpty(t, cfg.ListenAPI())
}

func TestLoadLogger(t *testing.T) {
	cfg, err := LoadConfig(testBuild)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	lgr, err := cfg.LoadLogger()
	require.NoError(t, err)
	require.NotNil(t, lgr)

	lgr.Info("Testing")
}

func TestLoadTracer(t *testing.T) {
	cfg, err := LoadConfig(testBuild)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	tracer, closer, err := cfg.LoadTracer()
	require.NoError(t, err)
	assert.NotNil(t, tracer)

	assert.NoError(t, closer.Close())
}
