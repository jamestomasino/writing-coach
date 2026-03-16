package session

import (
	"context"

	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/db"
)

type Context struct {
	UserID       int64
	TreeID       int64
	EnrollmentID int64
	UserSlug     string
	TreeSlug     string
}

func Resolve(ctx context.Context, store *db.Store, cfg config.Config) (Context, error) {
	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, cfg.DefaultUserSlug, cfg.WriterName, cfg.DefaultTreeSlug)
	if err != nil {
		return Context{}, err
	}
	return Context{
		UserID:       userID,
		TreeID:       treeID,
		EnrollmentID: enrollmentID,
		UserSlug:     cfg.DefaultUserSlug,
		TreeSlug:     cfg.DefaultTreeSlug,
	}, nil
}
