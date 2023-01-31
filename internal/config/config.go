// Config package contains constants and shared metrics of the server module.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/internal/config
package config

import (
	"database/sql"
	"os"
	"crypto/rsa"
	"encoding/json"
	"log"

)

// Database and Server context timeout values.
const (
	ContextDBTimeout  = 5
	ContextSrvTimeout = 10
	RequestTimeout  = 5
)

// Optional hashing Key.
var Key string
// Optional assymetric encription private Key.
var PrivateKey *rsa.PrivateKey
// Shared SQL database instance.
var DB *sql.DB
// Flag for SQL database use.
var DBFlag bool

// GetEnv function is used for retrieving variables passed in the command prompt.
func GetEnv(key string, fallback *string) *string {
	if value, ok := os.LookupEnv(key); ok {
		return &value
	}
	return fallback
}

type AgentConfig struct {
    Address string `json:"address"`
	ReportInterval string `json:"report_interval"`
    PollInterval string `"poll_interval"`
	CryptoKey string `json:"crypto_key"`
}

type ServerConfig struct {
    Address string `json:"address"`
	Restore string `json:"restore"`
    StoreInterval string `json:"store_interval"`
	StoreFile string `json:"store_file"`
	DatabaseDsn string `json:"database_dsn"`
	CryptoKey string `json:"crypto_key"`
}

// NewAgentConfig creates a new instance of AgentConfig
func NewAgentConfig() AgentConfig {
	agentConfig := AgentConfig{}
	agentConfig.Address = "127.0.0.1:8080"
	agentConfig.ReportInterval = "10"
	agentConfig.PollInterval = "2"
	return agentConfig
 }

 // NewServerConfig creates a new instance of ServerConfig
func NewServerConfig() ServerConfig {
	serverConfig := ServerConfig{}
	serverConfig.Address = "127.0.0.1:8080"
	serverConfig.Restore = "true"
	serverConfig.StoreInterval = "300"
	serverConfig.StoreFile = "/tmp/devops-metrics-db.json"
	return serverConfig
 }

 func LoadAgentConfiguration(file *string, config AgentConfig) AgentConfig {
    configFile, err := os.Open(*file)
    defer configFile.Close()
    if err != nil {
        log.Printf("Error happened when loading agent configuration. Err: %s", err)
    }
    jsonParser := json.NewDecoder(configFile)
    jsonParser.Decode(&config)
    return config
}

func LoadServerConfiguration(file *string, config ServerConfig) ServerConfig {
    configFile, err := os.Open(*file)
    defer configFile.Close()
    if err != nil {
        log.Printf("Error happened when loading agent configuration. Err: %s", err)
    }
    jsonParser := json.NewDecoder(configFile)
    jsonParser.Decode(&config)
    return config
}