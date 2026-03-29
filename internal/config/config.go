package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

const dirName = ".writing-coach"

type Config struct {
	ProjectRoot                    string        `json:"-"`
	ConfigPath                     string        `json:"-"`
	DataDir                        string        `json:"data_dir"`
	DatabaseURL                    string        `json:"database_url"`
	WriterName                     string        `json:"writer_name"`
	DefaultUserSlug                string        `json:"default_user_slug"`
	DefaultTreeSlug                string        `json:"default_tree_slug"`
	HTTPAddr                       string        `json:"http_addr"`
	APIToken                       string        `json:"-"`
	AllowInsecureAuth              bool          `json:"allow_insecure_auth"`
	AdminEmails                    []string      `json:"admin_emails"`
	KratosPublicURL                string        `json:"kratos_public_url"`
	OpenAIAPIKey                   string        `json:"-"`
	OpenAIBaseURL                  string        `json:"openai_base_url"`
	AIKeySecret                    string        `json:"-"`
	PromptModel                    string        `json:"prompt_model"`
	ReviewModel                    string        `json:"review_model"`
	PromptGenerationTimeout        time.Duration `json:"-"`
	AIValidateLimitPerMinute       int           `json:"ai_validate_limit_per_minute"`
	AIValidateGlobalLimitPerMinute int           `json:"ai_validate_global_limit_per_minute"`
	AIProviderEventRetentionDays   int           `json:"ai_provider_event_retention_days"`
	CalibrationMaintenanceEnabled  bool          `json:"calibration_maintenance_enabled"`
	CalibrationMaintenanceInterval time.Duration `json:"-"`
	CalibrationMinSamples          int           `json:"calibration_min_samples"`
	CalibrationLimitPerTrack       int           `json:"calibration_limit_per_track"`
	ValeBinary                     string        `json:"vale_binary"`
	LanguageToolURL                string        `json:"languagetool_url"`
	NLPAnalyzerURL                 string        `json:"nlp_analyzer_url"`
}

func Default(projectRoot string) Config {
	dataDir := filepath.Join(projectRoot, dirName)
	return Config{
		ProjectRoot:                    projectRoot,
		ConfigPath:                     filepath.Join(dataDir, "config.json"),
		DataDir:                        dataDir,
		DatabaseURL:                    filepath.Join(dataDir, "writing-coach.db"),
		WriterName:                     "Writer",
		DefaultUserSlug:                "default",
		DefaultTreeSlug:                domain.GlobalSkillGraphSlug,
		HTTPAddr:                       ":8080",
		KratosPublicURL:                "",
		OpenAIBaseURL:                  "https://api.openai.com/v1",
		PromptModel:                    "gpt-5-mini",
		ReviewModel:                    "gpt-5-mini",
		PromptGenerationTimeout:        45 * time.Second,
		AIValidateLimitPerMinute:       6,
		AIValidateGlobalLimitPerMinute: 60,
		AIProviderEventRetentionDays:   30,
		CalibrationMaintenanceEnabled:  true,
		CalibrationMaintenanceInterval: 30 * 24 * time.Hour,
		CalibrationMinSamples:          50,
		CalibrationLimitPerTrack:       200,
	}
}

func Load(projectRoot string) (Config, error) {
	cfg := Default(projectRoot)

	bytes, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
	} else {
		if err := json.Unmarshal(bytes, &cfg); err != nil {
			return Config{}, err
		}
	}

	cfg.ProjectRoot = projectRoot
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(projectRoot, dirName)
	}
	cfg.ConfigPath = filepath.Join(cfg.DataDir, "config.json")
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = filepath.Join(cfg.DataDir, "writing-coach.db")
	}
	if cfg.OpenAIBaseURL == "" {
		cfg.OpenAIBaseURL = "https://api.openai.com/v1"
	}
	if cfg.PromptModel == "" {
		cfg.PromptModel = "gpt-5-mini"
	}
	if cfg.ReviewModel == "" {
		cfg.ReviewModel = "gpt-5-mini"
	}
	if cfg.PromptGenerationTimeout <= 0 {
		cfg.PromptGenerationTimeout = 45 * time.Second
	}
	if cfg.AIValidateLimitPerMinute <= 0 {
		cfg.AIValidateLimitPerMinute = 6
	}
	if cfg.AIValidateGlobalLimitPerMinute <= 0 {
		cfg.AIValidateGlobalLimitPerMinute = 60
	}
	if cfg.AIProviderEventRetentionDays <= 0 {
		cfg.AIProviderEventRetentionDays = 30
	}
	if cfg.CalibrationMaintenanceInterval <= 0 {
		cfg.CalibrationMaintenanceInterval = 30 * 24 * time.Hour
	}
	if cfg.CalibrationMinSamples <= 0 {
		cfg.CalibrationMinSamples = 50
	}
	if cfg.CalibrationLimitPerTrack <= 0 {
		cfg.CalibrationLimitPerTrack = 200
	}
	if cfg.DefaultUserSlug == "" {
		cfg.DefaultUserSlug = "default"
	}
	if cfg.DefaultTreeSlug == "" {
		cfg.DefaultTreeSlug = domain.GlobalSkillGraphSlug
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	if value := os.Getenv("OPENAI_API_KEY"); value != "" {
		cfg.OpenAIAPIKey = value
	}
	if value := os.Getenv("OPENAI_BASE_URL"); value != "" {
		cfg.OpenAIBaseURL = value
	}
	if value := os.Getenv("WRITING_COACH_AI_KEY_SECRET"); value != "" {
		cfg.AIKeySecret = value
	}
	if value := os.Getenv("WRITING_COACH_PROMPT_MODEL"); value != "" {
		cfg.PromptModel = value
	}
	if value := os.Getenv("WRITING_COACH_REVIEW_MODEL"); value != "" {
		cfg.ReviewModel = value
	}
	if value := os.Getenv("WRITING_COACH_PROMPT_GENERATION_TIMEOUT"); value != "" {
		cfg.PromptGenerationTimeout = parseDuration(value, cfg.PromptGenerationTimeout)
	}
	if value := os.Getenv("WRITING_COACH_AI_VALIDATE_LIMIT_PER_MINUTE"); value != "" {
		cfg.AIValidateLimitPerMinute = parsePositiveInt(value, cfg.AIValidateLimitPerMinute)
	}
	if value := os.Getenv("WRITING_COACH_AI_VALIDATE_GLOBAL_LIMIT_PER_MINUTE"); value != "" {
		cfg.AIValidateGlobalLimitPerMinute = parsePositiveInt(value, cfg.AIValidateGlobalLimitPerMinute)
	}
	if value := os.Getenv("WRITING_COACH_AI_PROVIDER_EVENT_RETENTION_DAYS"); value != "" {
		cfg.AIProviderEventRetentionDays = parsePositiveInt(value, cfg.AIProviderEventRetentionDays)
	}
	if value := os.Getenv("WRITING_COACH_CALIBRATION_MAINTENANCE_ENABLED"); value != "" {
		cfg.CalibrationMaintenanceEnabled = parseBool(value, cfg.CalibrationMaintenanceEnabled)
	}
	if value := os.Getenv("WRITING_COACH_CALIBRATION_MAINTENANCE_INTERVAL"); value != "" {
		cfg.CalibrationMaintenanceInterval = parseDuration(value, cfg.CalibrationMaintenanceInterval)
	}
	if value := os.Getenv("WRITING_COACH_CALIBRATION_MIN_SAMPLES"); value != "" {
		cfg.CalibrationMinSamples = parsePositiveInt(value, cfg.CalibrationMinSamples)
	}
	if value := os.Getenv("WRITING_COACH_CALIBRATION_LIMIT_PER_TRACK"); value != "" {
		cfg.CalibrationLimitPerTrack = parsePositiveInt(value, cfg.CalibrationLimitPerTrack)
	}
	if value := os.Getenv("VALE_BINARY"); value != "" {
		cfg.ValeBinary = value
	}
	if value := os.Getenv("LANGUAGETOOL_URL"); value != "" {
		cfg.LanguageToolURL = value
	}
	if value := os.Getenv("WRITING_COACH_NLP_ANALYZER_URL"); value != "" {
		cfg.NLPAnalyzerURL = value
	}
	if value := os.Getenv("WRITING_COACH_HTTP_ADDR"); value != "" {
		cfg.HTTPAddr = value
	}
	if value := os.Getenv("WRITING_COACH_DATA_DIR"); value != "" {
		cfg.DataDir = value
		cfg.ConfigPath = filepath.Join(cfg.DataDir, "config.json")
	}
	if value := os.Getenv("WRITING_COACH_DATABASE_URL"); value != "" {
		cfg.DatabaseURL = value
	}
	if value := os.Getenv("WRITING_COACH_WRITER_NAME"); value != "" {
		cfg.WriterName = value
	}
	if value := os.Getenv("WRITING_COACH_DEFAULT_USER_SLUG"); value != "" {
		cfg.DefaultUserSlug = value
	}
	if value := os.Getenv("WRITING_COACH_DEFAULT_TREE_SLUG"); value != "" {
		cfg.DefaultTreeSlug = value
	}
	if value := os.Getenv("WRITING_COACH_API_TOKEN"); value != "" {
		cfg.APIToken = value
	}
	if value := os.Getenv("WRITING_COACH_ALLOW_INSECURE_AUTH"); value != "" {
		cfg.AllowInsecureAuth = parseBool(value, cfg.AllowInsecureAuth)
	}
	if value := os.Getenv("WRITING_COACH_ADMIN_EMAILS"); value != "" {
		cfg.AdminEmails = splitCSV(value)
	}
	if value := os.Getenv("WRITING_COACH_KRATOS_PUBLIC_URL"); value != "" {
		cfg.KratosPublicURL = value
	}

	return cfg, nil
}

func Save(cfg Config) error {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return err
	}

	payload := struct {
		DataDir                        string   `json:"data_dir"`
		DatabaseURL                    string   `json:"database_url"`
		WriterName                     string   `json:"writer_name"`
		DefaultUserSlug                string   `json:"default_user_slug"`
		DefaultTreeSlug                string   `json:"default_tree_slug"`
		HTTPAddr                       string   `json:"http_addr"`
		AllowInsecureAuth              bool     `json:"allow_insecure_auth"`
		AdminEmails                    []string `json:"admin_emails"`
		KratosPublicURL                string   `json:"kratos_public_url"`
		OpenAIBaseURL                  string   `json:"openai_base_url"`
		PromptModel                    string   `json:"prompt_model"`
		ReviewModel                    string   `json:"review_model"`
		PromptGenerationTimeout        string   `json:"prompt_generation_timeout"`
		AIValidateLimitPerMinute       int      `json:"ai_validate_limit_per_minute"`
		AIValidateGlobalLimitPerMinute int      `json:"ai_validate_global_limit_per_minute"`
		AIProviderEventRetentionDays   int      `json:"ai_provider_event_retention_days"`
		CalibrationMaintenanceEnabled  bool     `json:"calibration_maintenance_enabled"`
		CalibrationMaintenanceInterval string   `json:"calibration_maintenance_interval"`
		CalibrationMinSamples          int      `json:"calibration_min_samples"`
		CalibrationLimitPerTrack       int      `json:"calibration_limit_per_track"`
		ValeBinary                     string   `json:"vale_binary"`
		LanguageToolURL                string   `json:"languagetool_url"`
		NLPAnalyzerURL                 string   `json:"nlp_analyzer_url"`
	}{
		DataDir:                        cfg.DataDir,
		DatabaseURL:                    cfg.DatabaseURL,
		WriterName:                     cfg.WriterName,
		DefaultUserSlug:                cfg.DefaultUserSlug,
		DefaultTreeSlug:                cfg.DefaultTreeSlug,
		HTTPAddr:                       cfg.HTTPAddr,
		AllowInsecureAuth:              cfg.AllowInsecureAuth,
		AdminEmails:                    cfg.AdminEmails,
		KratosPublicURL:                cfg.KratosPublicURL,
		OpenAIBaseURL:                  cfg.OpenAIBaseURL,
		PromptModel:                    cfg.PromptModel,
		ReviewModel:                    cfg.ReviewModel,
		PromptGenerationTimeout:        cfg.PromptGenerationTimeout.String(),
		AIValidateLimitPerMinute:       cfg.AIValidateLimitPerMinute,
		AIValidateGlobalLimitPerMinute: cfg.AIValidateGlobalLimitPerMinute,
		AIProviderEventRetentionDays:   cfg.AIProviderEventRetentionDays,
		CalibrationMaintenanceEnabled:  cfg.CalibrationMaintenanceEnabled,
		CalibrationMaintenanceInterval: cfg.CalibrationMaintenanceInterval.String(),
		CalibrationMinSamples:          cfg.CalibrationMinSamples,
		CalibrationLimitPerTrack:       cfg.CalibrationLimitPerTrack,
		ValeBinary:                     cfg.ValeBinary,
		LanguageToolURL:                cfg.LanguageToolURL,
		NLPAnalyzerURL:                 cfg.NLPAnalyzerURL,
	}

	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.ConfigPath, append(bytes, '\n'), 0o644)
}

func splitCSV(raw string) []string {
	var items []string
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			items = append(items, value)
		}
	}
	return items
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseDuration(raw string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseBool(raw string, fallback bool) bool {
	value, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return value
}
