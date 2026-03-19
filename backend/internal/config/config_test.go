package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yanicksenn/ruthless/backend/internal/config"
)

func TestLoad(t *testing.T) {
	content := `
public {
  limits {
    max_card_text_length: 123
  }
}
`
	tmpfile, err := os.CreateTemp("", "config.*.textproto")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	cfg, err := config.Load(tmpfile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Public)
	require.NotNil(t, cfg.Public.Limits)
	assert.Equal(t, uint32(123), cfg.Public.Limits.MaxCardTextLength)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("non-existent.textproto")
	assert.Error(t, err)
}

func TestLoad_InvalidFormat(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "invalid.*.textproto")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte("invalid format"))
	require.NoError(t, err)
	tmpfile.Close()

	_, err = config.Load(tmpfile.Name())
	assert.Error(t, err)
}
