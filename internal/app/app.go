package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/api"
	"github.com/tomasino/writing-coach/internal/cli"
	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/openai"
	"github.com/tomasino/writing-coach/internal/prompt"
	"github.com/tomasino/writing-coach/internal/review"
	"github.com/tomasino/writing-coach/internal/session"
)

type App struct {
	api   api.Server
	cli   cli.CLI
	store *db.Store
}

func New(ctx context.Context) (*App, error) {
	projectRoot, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}

	store, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := store.SQL.PingContext(ctx); err != nil {
		_ = store.Close()
		return nil, err
	}
	if err := store.Migrate(ctx, filepath.Join(projectRoot, "migrations")); err != nil {
		_ = store.Close()
		return nil, err
	}
	if err := store.EnsureSeedData(ctx, cfg.WriterName); err != nil {
		_ = store.Close()
		return nil, err
	}
	if err := store.EnsureAdminEmails(ctx, cfg.AdminEmails); err != nil {
		_ = store.Close()
		return nil, err
	}

	openAIClient := openai.NewClient(cfg)
	valeBinary := cfg.ValeBinary
	if valeBinary == "" {
		localVale := filepath.Join(projectRoot, ".writing-coach", "bin", "vale")
		if _, err := os.Stat(localVale); err == nil {
			valeBinary = localVale
		}
	}
	if valeBinary == "" {
		if found, err := exec.LookPath("vale"); err == nil {
			valeBinary = found
		}
	}
	analyzerService := analyzer.NewService(
		analyzer.Heuristic{},
		analyzer.Vale{
			Binary:        valeBinary,
			ConfigPath:    filepath.Join(projectRoot, ".vale.ini"),
			WorkingDir:    projectRoot,
			StylesPath:    filepath.Join(projectRoot, "styles"),
			MinAlertLevel: "suggestion",
			BaseStyles:    []string{"WritingCoachCore"},
		},
		analyzer.LanguageTool{BaseURL: cfg.LanguageToolURL},
		analyzer.NLP{BaseURL: cfg.NLPAnalyzerURL},
	)
	appContext, err := session.Resolve(ctx, store, cfg)
	if err != nil {
		_ = store.Close()
		return nil, err
	}

	return &App{
		api: api.Server{
			Config:     cfg,
			Store:      store,
			Prompts:    prompt.NewService(openAIClient).WithGenerationTimeout(cfg.PromptGenerationTimeout),
			Reviews:    review.NewService(openAIClient, analyzerService),
			Curriculum: curriculum.NewService(),
		},
		cli: cli.CLI{
			Config:     cfg,
			AppContext: appContext,
			Store:      store,
			Prompts:    prompt.NewService(openAIClient).WithGenerationTimeout(cfg.PromptGenerationTimeout),
			Reviews:    review.NewService(openAIClient, analyzerService),
			Curriculum: curriculum.NewService(),
		},
		store: store,
	}, nil
}

func (a *App) Run(ctx context.Context, args []string) error {
	defer a.store.Close()
	if len(args) > 0 && args[0] == "serve" {
		return a.api.Serve(ctx)
	}
	return a.cli.Run(ctx, args)
}
