package config

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// GwAutoConfig represents the configuration
type GwAutoConfig struct {
	// Simulators is a list of simulator configurations
	Simulators   []SimulatorConfig `toml:"simulators"`
	SimulatorMap map[string]SimulatorConfig
}

// InitConfigMap convert slice to map
func (c *GwAutoConfig) InitConfigMap() {
	c.SimulatorMap = make(map[string]SimulatorConfig)
	if c.Simulators != nil {
		for _, s := range c.Simulators {
			c.SimulatorMap[s.Name] = s
		}
	}
}

// SimulatorConfig represents the configuration for a simulator
// It includes the
// name, shuld be unique
// type, type can be oms or tgw
// communication: common types are tcp, udp, http
// protocol,  protocol can be binary-szse, json-szse, etc.
// server_address, the address of the server to connect to
// listen_address, the address to listen on for incoming connections
// auto_start, whether to start the simulator automatically
type SimulatorConfig struct {
	Name          string `toml:"name"`
	Type          string `toml:"type"`
	Communication string `toml:"communication"`
	Protocol      string `toml:"protocol"`
	ServerAddress string `toml:"server_address"`
	ListenAddress string `toml:"listen_address"`
	AutoStart     bool   `toml:"auto_start"`
}

// ParseConfig reads the configuration file and returns a GwAutoConfig object
func ParseConfig(filePath string) (*GwAutoConfig, error) {
	var config GwAutoConfig
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open config file: %v", err)
		return nil, err
	}
	defer file.Close()

	if _, err := toml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatalf("failed to decode toml: %v", err)
		return nil, err
	}

	fmt.Printf("Parsed config: %+v\n", config.Simulators)
	return &config, nil
}
