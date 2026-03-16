package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	root := t.TempDir()
	cfg := Default(root)
	cfg.WriterName = "Tomasino"
	cfg.PromptModel = "test-prompt"
	cfg.ReviewModel = "test-review"
	cfg.ValeBinary = "/tmp/vale"
	cfg.LanguageToolURL = "http://localhost:8081"

	if err := Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if loaded.WriterName != "Tomasino" {
		t.Fatalf("writer name = %q", loaded.WriterName)
	}
	if loaded.PromptModel != "test-prompt" || loaded.ReviewModel != "test-review" {
		t.Fatalf("unexpected models: %+v", loaded)
	}
	if loaded.ValeBinary != "/tmp/vale" || loaded.LanguageToolURL != "http://localhost:8081" {
		t.Fatalf("unexpected analyzer config: %+v", loaded)
	}
	if _, err := os.Stat(filepath.Join(root, ".writing-coach", "config.json")); err != nil {
		t.Fatalf("config file missing: %v", err)
	}
}

func TestLoadReadsAPITokenFromEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WRITING_COACH_API_TOKEN", "secret")

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.APIToken != "secret" {
		t.Fatalf("api token = %q", loaded.APIToken)
	}
}
