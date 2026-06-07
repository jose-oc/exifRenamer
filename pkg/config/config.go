package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// NamingConfig holds formatting rules and patterns for directories and file names.
type NamingConfig struct {
	FolderPattern string `mapstructure:"folder_pattern"`
	FilePattern   string `mapstructure:"file_pattern"`
	DefaultPrefix string `mapstructure:"default_prefix"`
	DefaultSuffix string `mapstructure:"default_suffix"`
}

// Config represents the application configuration.
type Config struct {
	Destination        string       `mapstructure:"destination"`
	Recursive          bool         `mapstructure:"recursive"`
	DryRun             bool         `mapstructure:"dry_run"`
	LowercaseExtension bool         `mapstructure:"lowercase_extension"`
	Verbose            bool         `mapstructure:"verbose"`
	Quiet              bool         `mapstructure:"quiet"`
	NoColor            bool         `mapstructure:"no_color"`
	Naming             NamingConfig `mapstructure:"naming"`
}

// LoadConfig reads configuration files and unmarshals them into the Config struct.
// It assumes any Cobra flags have already been bound to Viper.
func LoadConfig(configFile string) (*Config, error) {
	viper.SetDefault("destination", "~/Pictures/Sorted")
	viper.SetDefault("recursive", true)
	viper.SetDefault("dry_run", false)
	viper.SetDefault("lowercase_extension", true)
	viper.SetDefault("verbose", false)
	viper.SetDefault("quiet", false)
	viper.SetDefault("no_color", false)
	viper.SetDefault("naming.folder_pattern", "%Y/%Y-%m/%Y-%m-%d-%s")
	viper.SetDefault("naming.file_pattern", "%p%Y%m%d-%H%M%S-%s%c%e")
	viper.SetDefault("naming.default_prefix", "")
	viper.SetDefault("naming.default_suffix", "")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// Default search paths:
		// 1. Current directory (for config.yaml or .exifrenamer.yaml)
		// 2. Home directory (for .exifrenamer.yaml or config.yaml)
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
		viper.SetConfigName(".exifrenamer")
	}

	// Read config file (if it exists)
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file is not found, unless a specific config file was requested
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && configFile != "" {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Resolve the home directory in destination if it starts with "~"
	cfg.Destination = ExpandHomeDir(cfg.Destination)

	return &cfg, nil
}

// ExpandHomeDir expands the leading tilde ~ in a path to the user's home directory.
func ExpandHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
