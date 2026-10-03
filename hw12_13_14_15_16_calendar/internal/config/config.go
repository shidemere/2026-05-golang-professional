// Package config read config from file and construct configs for whole project.
package config

import (
	"fmt"
	"log"
	"os"

	"go.yaml.in/yaml/v3"
)

// Config is global configuration for project.
type Config struct {
	LoggerConf   LoggerConfig   `yaml:"logger"`
	ServerConf   ServerConfig   `yaml:"server"`
	PostgresConf PostgresConfig `yaml:"postgres"`
	UseInMemory  bool           `yaml:"use_in_memory"`
}

// LoggerConfig contains configuration for logger.
type LoggerConfig struct {
	Level     string `yaml:"level"`
	AddSource bool   `yaml:"add_source"`
}

// ServerConfig contains configuration for servecr.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// PostgresConfig contains configuration for postgres database.
type PostgresConfig struct {
	// TODO: add typecheck for all configs.
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	SSLmode      string `yaml:"sslmode"`
	DBName       string `yaml:"dbname"`
	PoolMaxConns int    `yaml:"pool_max_conns"`
	PoolMinConns int    `yaml:"pool_min_conns"`
}

// TODO: Use defaults.

// NewConfig build global configuration with subconfigss.
func NewConfig(path string) Config {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("can't parse config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		log.Fatalf("Error decoding yaml: %v", err)
	}

	err = setEnvVariables(&config)
	if err != nil {
		log.Fatalf("can't parse configuration: %v", err)
	}

	return config
}

// TODO: Collect this somehow.
func setEnvVariables(config *Config) error {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LoggerConf.Level = logLevel
	}

	if dbHost := os.Getenv("POSTGRES_HOST"); dbHost != "" {
		config.PostgresConf.Host = dbHost
	} else if config.PostgresConf.Host == "" {
		return fmt.Errorf("need to specify POSTGRES_HOST for application running")
	}

	if dbPort := os.Getenv("POSTGRES_PORT"); dbPort != "" {
		config.PostgresConf.Port = dbPort
	} else if config.PostgresConf.Port == "" {
		return fmt.Errorf("need to specify POSGRES_PORT for application running")
	}

	if dbName := os.Getenv("POSTGRES_DATABASE"); dbName != "" {
		config.PostgresConf.DBName = dbName
	} else if config.PostgresConf.DBName == "" {
		return fmt.Errorf("need to specify POSTGRES_DATABASE for application running")
	}

	if dbUser := os.Getenv("POSTGRES_USER"); dbUser != "" {
		config.PostgresConf.User = dbUser
	} else {
		return fmt.Errorf("need to specify POSTGRES_USER for application running")
	}

	if dbPassword := os.Getenv("POSTGRES_PASSWORD"); dbPassword != "" {
		config.PostgresConf.Password = dbPassword
	} else {
		return fmt.Errorf("need to specify POSTGRES_PASSWORD for application running")
	}
	return nil
}
