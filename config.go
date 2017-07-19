package main

import (
	"os"
	"fmt"
	"strings"
	"io"
	"path/filepath"
	"log"
	"github.com/BurntSushi/toml"
)

// BuildVersion returns the build version of adns
var buildVersion = "1.0.0"

// BuildVersion returns the build version of adns
var configVersion = "1.0.0"

type Config struct {
	Version			string
	Log			string
	LogLevel		int
	DnsBind			string
	ApiBind			string
	MysqlConnectionString	string
}

var defaultConfig = `# version this config was generated from
version = "%s"

# location of the log file
log = "adns.log"

# what kind of information should be logged, 0 = errors and important operations, 1 = dns queries, 2 = debug
loglevel = 0

# address to bind to for the DNS server
bindbind = "0.0.0.0:53"

# address to bind to for the API server
apibind = "127.0.0.1:8080"

# mysql server connection string
mysqlconnectionstring = "root:123456@tcp(127.0.0.1:3306)/adns?charset=utf8mb4"
`

// Config is the global configuration
var config Config

// LoadConfig loads the given config file
func loadConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := generateConfig(path); err != nil {
			return err
		}
	}

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return fmt.Errorf("could not load config: %s", err)
	}

	if config.Version != configVersion {
		if config.Version == "" {
			config.Version = "none"
		}

		log.Printf("warning, adns.toml is out of date!\nconfig v%s\nadns config v%s\nadns v%s\nplease update your config\n", config.Version, configVersion, buildVersion)
	} else {
		log.Printf("adns v%s\n", buildVersion)
	}

	return nil
}

// Generate config file if config file not exists
func generateConfig(path string) error {
	output, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not generate config: %s", err)
	}
	defer output.Close()

	r := strings.NewReader(fmt.Sprintf(defaultConfig, configVersion))
	if _, err := io.Copy(output, r); err != nil {
		return fmt.Errorf("could not copy default config: %s", err)
	}

	if abs, err := filepath.Abs(path); err == nil {
		log.Printf("generated default config %s\n", abs)
	}

	return nil
}




