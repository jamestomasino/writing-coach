package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadConfig(t *testing.T) {
	root := t.TempDir()
	cfg := Default(root)
	cfg.WriterName = "Tomasino"
	cfg.PromptModel = "test-prompt"
	cfg.ReviewModel = "test-review"
	cfg.ValeBinary = "/tmp/vale"
	cfg.LanguageToolURL = "http://localhost:8081"
	cfg.NLPAnalyzerURL = "http://localhost:8020"

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
	if loaded.ValeBinary != "/tmp/vale" || loaded.LanguageToolURL != "http://localhost:8081" || loaded.NLPAnalyzerURL != "http://localhost:8020" {
		t.Fatalf("unexpected analyzer config: %+v", loaded)
	}
	if _, err := os.Stat(filepath.Join(root, ".writing-coach", "config.json")); err != nil {
		t.Fatalf("config file missing: %v", err)
	}
}

func TestLoadReadsAPITokenFromEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WRITING_COACH_API_TOKEN", "secret")
	t.Setenv("WRITING_COACH_KRATOS_PUBLIC_URL", "http://kratos:4433")
	t.Setenv("WRITING_COACH_WRITER_NAME", "Coach Writer")
	t.Setenv("WRITING_COACH_DEFAULT_USER_SLUG", "coach")
	t.Setenv("WRITING_COACH_DEFAULT_TREE_SLUG", "youth-writing-foundations")
	t.Setenv("WRITING_COACH_ADMIN_EMAILS", "writer@example.com, coach@example.com")

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.APIToken != "secret" {
		t.Fatalf("api token = %q", loaded.APIToken)
	}
	if loaded.KratosPublicURL != "http://kratos:4433" {
		t.Fatalf("kratos public url = %q", loaded.KratosPublicURL)
	}
	if loaded.WriterName != "Coach Writer" {
		t.Fatalf("writer name = %q", loaded.WriterName)
	}
	if loaded.DefaultUserSlug != "coach" {
		t.Fatalf("default user slug = %q", loaded.DefaultUserSlug)
	}
	if loaded.DefaultTreeSlug != "youth-writing-foundations" {
		t.Fatalf("default tree slug = %q", loaded.DefaultTreeSlug)
	}
	if len(loaded.AdminEmails) != 2 || loaded.AdminEmails[0] != "writer@example.com" {
		t.Fatalf("admin emails = %#v", loaded.AdminEmails)
	}
}

func TestLoadReadsAIValidationLimitsFromEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WRITING_COACH_AI_VALIDATE_LIMIT_PER_MINUTE", "3")
	t.Setenv("WRITING_COACH_AI_VALIDATE_GLOBAL_LIMIT_PER_MINUTE", "12")
	t.Setenv("WRITING_COACH_AI_PROVIDER_EVENT_RETENTION_DAYS", "21")
	t.Setenv("WRITING_COACH_CALIBRATION_MAINTENANCE_ENABLED", "false")
	t.Setenv("WRITING_COACH_CALIBRATION_MAINTENANCE_INTERVAL", "168h")
	t.Setenv("WRITING_COACH_CALIBRATION_MIN_SAMPLES", "25")
	t.Setenv("WRITING_COACH_CALIBRATION_LIMIT_PER_TRACK", "150")

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.AIValidateLimitPerMinute != 3 {
		t.Fatalf("per-user validation limit = %d", loaded.AIValidateLimitPerMinute)
	}
	if loaded.AIValidateGlobalLimitPerMinute != 12 {
		t.Fatalf("global validation limit = %d", loaded.AIValidateGlobalLimitPerMinute)
	}
	if loaded.AIProviderEventRetentionDays != 21 {
		t.Fatalf("provider event retention days = %d", loaded.AIProviderEventRetentionDays)
	}
	if loaded.CalibrationMaintenanceEnabled {
		t.Fatal("expected calibration maintenance to be disabled by env")
	}
	if loaded.CalibrationMaintenanceInterval != 168*time.Hour {
		t.Fatalf("calibration maintenance interval = %s", loaded.CalibrationMaintenanceInterval)
	}
	if loaded.CalibrationMinSamples != 25 {
		t.Fatalf("calibration min samples = %d", loaded.CalibrationMinSamples)
	}
	if loaded.CalibrationLimitPerTrack != 150 {
		t.Fatalf("calibration limit per track = %d", loaded.CalibrationLimitPerTrack)
	}
}

func TestLoadReadsPromptGenerationTimeoutFromEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WRITING_COACH_PROMPT_GENERATION_TIMEOUT", "75s")

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.PromptGenerationTimeout != 75*time.Second {
		t.Fatalf("prompt generation timeout = %s", loaded.PromptGenerationTimeout)
	}
}
