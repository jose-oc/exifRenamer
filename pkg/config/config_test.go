package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig_Defaults(t *testing.T) {
	viper.Reset()
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Destination == "" {
		t.Error("expected non-empty default destination")
	}
	if !cfg.Recursive {
		t.Error("expected default recursive to be true")
	}
	if cfg.DryRun {
		t.Error("expected default dry_run to be false")
	}
	if !cfg.LowercaseExtension {
		t.Error("expected default lowercase_extension to be true")
	}
	if cfg.Naming.FolderPattern != "%Y/%Y-%m/%Y-%m-%d-%s" {
		t.Errorf("unexpected default folder pattern: %s", cfg.Naming.FolderPattern)
	}
	if cfg.Naming.FilePattern != "%p%Y%m%d-%H%M%S-%s%c%e" {
		t.Errorf("unexpected default file pattern: %s", cfg.Naming.FilePattern)
	}
}

func TestLoadConfig_WithFile(t *testing.T) {
	viper.Reset()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `
destination: "/tmp/custom_sorted"
recursive: false
dry_run: true
lowercase_extension: false
verbose: true
naming:
  folder_pattern: "%Y/%M/%D"
  file_pattern: "%P%Y-%M-%D%S%C%F"
  default_prefix: "pre_"
  default_suffix: "_suff"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Destination != "/tmp/custom_sorted" {
		t.Errorf("expected destination to be /tmp/custom_sorted, got %s", cfg.Destination)
	}
	if cfg.Recursive {
		t.Error("expected recursive to be false")
	}
	if !cfg.DryRun {
		t.Error("expected dry_run to be true")
	}
	if cfg.LowercaseExtension {
		t.Error("expected lowercase_extension to be false")
	}
	if cfg.Naming.FolderPattern != "%Y/%M/%D" {
		t.Errorf("expected folder pattern %%Y/%%M/%%D, got %s", cfg.Naming.FolderPattern)
	}
	if cfg.Naming.FilePattern != "%P%Y-%M-%D%S%C%F" {
		t.Errorf("expected file pattern %%P%%Y-%%M-%%D%%S%%C%%F, got %s", cfg.Naming.FilePattern)
	}
	if cfg.Naming.DefaultPrefix != "pre_" {
		t.Errorf("expected prefix pre_, got %s", cfg.Naming.DefaultPrefix)
	}
	if cfg.Naming.DefaultSuffix != "_suff" {
		t.Errorf("expected suffix _suff, got %s", cfg.Naming.DefaultSuffix)
	}
}

func TestExpandHomeDir(t *testing.T) {
	path := "/absolute/path"
	expanded := ExpandHomeDir(path)
	if expanded != path {
		t.Errorf("expected %s, got %s", path, expanded)
	}

	pathWithTilde := "~/Pictures"
	expandedWithTilde := ExpandHomeDir(pathWithTilde)
	if expandedWithTilde == pathWithTilde {
		t.Error("expected tilde path to be expanded")
	}
}
