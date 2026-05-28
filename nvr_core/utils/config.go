package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"nvr_core/security"
	"os"
)

// Define structs that match the JSON structure.
// The `json:"..."` tags tell Go which JSON field maps to which variable.

type Config struct {
	Server  ServerConfig   `json:"server"`
	// Cameras []CameraConfig `json:"cameras"`
}

type ServerConfig struct {
	Port          int    `json:"port"`
	WebPort       int    `json:"webPort"`
	DBPath        string `json:"db_path"`
	StoragePath   string `json:"storage_path"`
	KeyPath       string `json:"key_path"`
	RetentionDays int    `json:"retention_days"`
	masterKey     []byte `json:"-"`
}

// type CameraConfig struct {
// 	ID      int    `json:"id"`
// 	Name    string `json:"name"`
// 	URL     string `json:"url"`
// 	Enabled bool   `json:"enabled"`
// }

// LoadConfig reads the file at the given path and returns a filled Config struct
func LoadConfig(path string) (*Config, error) {
	// Read the file content
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config json: %w", err)
	}

	if cfg.Server.KeyPath == "" {
		cfg.Server.KeyPath = "./master.key"
	}

	// (Optional)
	if cfg.Server.DBPath == "" {
		cfg.Server.DBPath = "./"
	}

	// (Optional) Validate basic settings
	if cfg.Server.StoragePath == "" {
		cfg.Server.StoragePath = "./recordings" // Default fallback
	}

	cfg.Server.InitMasterKey()

	return &cfg, nil
}

// Make sure we have valid KeyPath before run this.
func (s *ServerConfig) InitMasterKey() {

	// Load or generate the key
	key, err := security.LoadOrCreateMasterKey(s.KeyPath)
	if err != nil {
		log.Fatalf("Security initialization failed: %v", err)
	}

	s.PopulateMasterKey(key)

}


func (s *ServerConfig) PopulateMasterKey(key []byte) {
	s.masterKey = key
}

func (s *ServerConfig) MasterKey() ([]byte) {
	return s.masterKey
}
