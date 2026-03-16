package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const dirName = ".writing-coach"

type Config struct {
	ProjectRoot     string `json:"-"`
	ConfigPath      string `json:"-"`
	DataDir         string `json:"data_dir"`
	DatabaseURL     string `json:"database_url"`
	WriterName      string `json:"writer_name"`
	DefaultUserSlug string `json:"default_user_slug"`
	DefaultTreeSlug string `json:"default_tree_slug"`
	HTTPAddr        string `json:"http_addr"`
	APIToken        string `json:"-"`
	OpenAIAPIKey    string `json:"-"`
	OpenAIBaseURL   string `json:"openai_base_url"`
	PromptModel     string `json:"prompt_model"`
	ReviewModel     string `json:"review_model"`
	ValeBinary      string `json:"vale_binary"`
	LanguageToolURL string `json:"languagetool_url"`
}

func Default(projectRoot string) Config {
	dataDir := filepath.Join(projectRoot, dirName)
	return Config{
		ProjectRoot:     projectRoot,
		ConfigPath:      filepath.Join(dataDir, "config.json"),
		DataDir:         dataDir,
		DatabaseURL:     filepath.Join(dataDir, "writing-coach.db"),
		WriterName:      "Writer",
		DefaultUserSlug: "default",
		DefaultTreeSlug: "mythic-tragedy-apprenticeship",
		HTTPAddr:        ":8080",
		OpenAIBaseURL:   "https://api.openai.com/v1",
		PromptModel:     "gpt-5-mini",
		ReviewModel:     "gpt-5-mini",
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
	if cfg.DefaultUserSlug == "" {
		cfg.DefaultUserSlug = "default"
	}
	if cfg.DefaultTreeSlug == "" {
		cfg.DefaultTreeSlug = "mythic-tragedy-apprenticeship"
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
	if value := os.Getenv("WRITING_COACH_PROMPT_MODEL"); value != "" {
		cfg.PromptModel = value
	}
	if value := os.Getenv("WRITING_COACH_REVIEW_MODEL"); value != "" {
		cfg.ReviewModel = value
	}
	if value := os.Getenv("VALE_BINARY"); value != "" {
		cfg.ValeBinary = value
	}
	if value := os.Getenv("LANGUAGETOOL_URL"); value != "" {
		cfg.LanguageToolURL = value
	}
	if value := os.Getenv("WRITING_COACH_HTTP_ADDR"); value != "" {
		cfg.HTTPAddr = value
	}
	if value := os.Getenv("WRITING_COACH_API_TOKEN"); value != "" {
		cfg.APIToken = value
	}

	return cfg, nil
}

func Save(cfg Config) error {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return err
	}

	payload := struct {
		DataDir         string `json:"data_dir"`
		DatabaseURL     string `json:"database_url"`
		WriterName      string `json:"writer_name"`
		DefaultUserSlug string `json:"default_user_slug"`
		DefaultTreeSlug string `json:"default_tree_slug"`
		HTTPAddr        string `json:"http_addr"`
		OpenAIBaseURL   string `json:"openai_base_url"`
		PromptModel     string `json:"prompt_model"`
		ReviewModel     string `json:"review_model"`
		ValeBinary      string `json:"vale_binary"`
		LanguageToolURL string `json:"languagetool_url"`
	}{
		DataDir:         cfg.DataDir,
		DatabaseURL:     cfg.DatabaseURL,
		WriterName:      cfg.WriterName,
		DefaultUserSlug: cfg.DefaultUserSlug,
		DefaultTreeSlug: cfg.DefaultTreeSlug,
		HTTPAddr:        cfg.HTTPAddr,
		OpenAIBaseURL:   cfg.OpenAIBaseURL,
		PromptModel:     cfg.PromptModel,
		ReviewModel:     cfg.ReviewModel,
		ValeBinary:      cfg.ValeBinary,
		LanguageToolURL: cfg.LanguageToolURL,
	}

	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.ConfigPath, append(bytes, '\n'), 0o644)
}
