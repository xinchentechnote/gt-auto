package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinchentechnote/gt-auto/pkg/config"
)

func TestParseConfig(t *testing.T) {
	config, err := config.ParseConfig("testdata/gw-auto.toml")
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	assert.NotNil(t, config)
	assert.Len(t, config.Simulators, 2)
	// Check the first simulator
	assert.Equal(t, "szse_bin_oms_1", config.Simulators[0].Name)
	assert.Equal(t, "oms", config.Simulators[0].Type)
	assert.Equal(t, "tcp", config.Simulators[0].Communication)
	assert.Equal(t, "binary-szse", config.Simulators[0].Protocol)
	assert.Equal(t, "localhost:9001", config.Simulators[0].ServerAddress)
	assert.Equal(t, "", config.Simulators[0].ListenAddress)
	assert.False(t, config.Simulators[0].AutoStart)
	// Check the second simulator
	assert.Equal(t, "szse_bin_tgw_1", config.Simulators[1].Name)
	assert.Equal(t, "tgw", config.Simulators[1].Type)
	assert.Equal(t, "tcp", config.Simulators[1].Communication)
	assert.Equal(t, "binary-szse", config.Simulators[1].Protocol)
	assert.Equal(t, "", config.Simulators[1].ServerAddress)
	assert.Equal(t, ":9001", config.Simulators[1].ListenAddress)
	assert.True(t, config.Simulators[1].AutoStart)
}
