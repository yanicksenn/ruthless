package config

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/prototext"
	ruthlespb "github.com/yanicksenn/ruthless/api/v1"
)

// Load loads the configuration from a textproto file.
func Load(path string) (*ruthlespb.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &ruthlespb.Config{}
	opts := prototext.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal textproto: %w", err)
	}

	return cfg, nil
}
