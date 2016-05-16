package main

import (
	"flag"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// DefaultConfigFile is the default path to the config file.
const DefaultConfigFile = "/etc/beacon.yml"

// Docker runtime configuration.
type Docker struct {
	Socket string
	HostIP string
	Label  string
}

// Validate the docker configuration.
func (c *Docker) Validate() error {
	if c == nil {
		return errors.New("missing Docker config object")
	}
	if c.Socket == "" {
		return errors.New("Docker.Socket may not be empty")
	}
	if c.HostIP == "" {
		return errors.New("Docker.HostIP may not be empty")
	}
	if c.Label == "" {
		return errors.New("Docker.Label may not be empty")
	}
	return nil
}

// SNS backend configuration.
type SNS struct {
	Region string
	Topic  string
}

// Validate the SNS configuration.
func (c *SNS) Validate() error {
	if c == nil {
		return errors.New("missing SNS config object")
	}
	if c.Region == "" {
		return errors.New("SNS.Region may not be empty")
	}
	if c.Topic == "" {
		return errors.New("SNS.Topic may not be empty")
	}
	return nil
}

// Backend configuration object.
type Backend struct {
	SNS    *SNS
	Filter map[string]string
}

// Validate the backend configuration.
func (c *Backend) Validate() error {
	if c.SNS != nil {
		return c.SNS.Validate()
	}
	return errors.New("backend not supported")
}

// Config holds Beacon configuration.
type Config struct {
	Backends []Backend
	Docker   Docker
}

// Validate the Beacon configuration.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("nil config object")
	}
	if err := c.Docker.Validate(); err != nil {
		return err
	}
	if len(c.Backends) == 0 {
		return errors.New("no backends configured")
	}
	for _, backend := range c.Backends {
		if err := backend.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// DefaultConfig generates a default configuration.
func DefaultConfig() *Config {
	dockerSocket := os.Getenv("DOCKER_HOST")
	if dockerSocket == "" {
		dockerSocket = "unix:///var/run/docker.sock"
	}
	dockerHostIP := os.Getenv("DOCKER_IP")
	if dockerHostIP == "" {
		dockerHostIP = "127.0.0.1"
	}

	return &Config{
		Docker: Docker{
			Socket: dockerSocket,
			HostIP: dockerHostIP,
		},
		Backends: []Backend{},
	}
}

// Configure Beacon. Loads configuration into a Config.
func Configure(args []string) *Config {
	var path string
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&path, "config", DefaultConfigFile, "The path to the config file.")
	flags.Parse(args[1:])

	data, err := ioutil.ReadFile(path)
	if err != nil {
		Logger.Fatalf("failed to read config %s: %s\n", path, err)
	}
	config := DefaultConfig()
	boop := map[interface{}]interface{}{}
	yaml.Unmarshal(data, boop)
	if err := yaml.Unmarshal(data, config); err != nil {
		Logger.Fatalf("failed to parse config %s: %s\n", path, err)
	}
	if err := config.Validate(); err != nil {
		Logger.Fatalf("configuration invalid: %s\n", err)
	}
	return config
}
