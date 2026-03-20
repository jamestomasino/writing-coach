package api

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/session"
)

type treeResponse struct {
	ID             int64         `json:"id"`
	Slug           string        `json:"slug"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	SeedCodes      []string      `json:"seed_codes,omitempty"`
	PrioritySkills []string      `json:"priority_skills,omitempty"`
	TGOs           []tgoResponse `json:"tgos,omitempty"`
	CreatedAt      string        `json:"created_at,omitempty"`
}

type treeVersionResponse struct {
	ID          int64  `json:"id"`
	TreeID      int64  `json:"tree_id"`
	TreeSlug    string `json:"tree_slug"`
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func (s Server) handleTreesList(w http.ResponseWriter, r *http.Request) {
	trees, err := s.Store.ListTrees(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	includeTGOs := r.URL.Query().Get("include_tgos") == "1"
	writeJSON(w, http.StatusOK, map[string]any{"trees": s.toTreeResponses(r.Context(), trees, includeTGOs)})
}

func (s Server) handleTreeCreate(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var payload struct {
		Slug           string        `json:"slug"`
		Title          string        `json:"title"`
		Description    string        `json:"description"`
		SeedCodes      []string      `json:"seed_codes"`
		PrioritySkills []string      `json:"priority_skills"`
		TGOs           []tgoResponse `json:"tgos"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	def, err := buildTreeDefinition(payload.Slug, payload.Title, payload.Description, payload.SeedCodes, payload.PrioritySkills, payload.TGOs)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.Store.SaveTreeDefinition(r.Context(), def); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tree, err := s.Store.TreeBySlug(r.Context(), def.Slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"tree": s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0]})
}

func (s Server) handleTreeGet(w http.ResponseWriter, r *http.Request) {
	tree, err := s.Store.TreeBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	response := s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0]
	if appContext, err := s.resolveSession(r.Context(), r); err == nil {
		response = s.applyGeneratedTreeProfileDisplay(r.Context(), appContext, tree, response)
	}
	writeJSON(w, http.StatusOK, map[string]any{"tree": response})
}

func (s Server) handleTreeUpdate(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var payload struct {
		Title          string        `json:"title"`
		Description    string        `json:"description"`
		SeedCodes      []string      `json:"seed_codes"`
		PrioritySkills []string      `json:"priority_skills"`
		TGOs           []tgoResponse `json:"tgos"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	def, err := buildTreeDefinition(r.PathValue("slug"), payload.Title, payload.Description, payload.SeedCodes, payload.PrioritySkills, payload.TGOs)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.Store.SaveTreeDefinition(r.Context(), def); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tree, err := s.Store.TreeBySlug(r.Context(), def.Slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tree": s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0]})
}

func (s Server) handleTreeVersionsList(w http.ResponseWriter, r *http.Request) {
	versions, err := s.Store.ListTreeVersions(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]treeVersionResponse, 0, len(versions))
	for _, version := range versions {
		items = append(items, treeVersionResponse{
			ID:          version.ID,
			TreeID:      version.TreeID,
			TreeSlug:    version.TreeSlug,
			Version:     version.Version,
			Title:       version.Title,
			Description: version.Description,
			CreatedAt:   db.Since(version.CreatedAt),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": items})
}

func (s Server) handleTreeVersionGet(w http.ResponseWriter, r *http.Request) {
	version, err := strconv.Atoi(r.PathValue("version"))
	if err != nil || version <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid version"))
		return
	}
	meta, def, err := s.Store.TreeVersionByNumber(r.Context(), r.PathValue("slug"), version)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version": treeVersionResponse{
			ID:          meta.ID,
			TreeID:      meta.TreeID,
			TreeSlug:    meta.TreeSlug,
			Version:     meta.Version,
			Title:       meta.Title,
			Description: meta.Description,
			CreatedAt:   db.Since(meta.CreatedAt),
		},
		"tree": treeResponse{
			Slug:           def.Slug,
			Title:          def.Title,
			Description:    def.Description,
			SeedCodes:      append([]string(nil), def.SeedCodes...),
			PrioritySkills: append([]string(nil), def.PrioritySkills...),
			TGOs:           toTGOResponses(def.TGOs),
		},
	})
}

func (s Server) handleTreeDiff(w http.ResponseWriter, r *http.Request) {
	fromVersion, err := strconv.Atoi(r.URL.Query().Get("from"))
	if err != nil || fromVersion <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid from version"))
		return
	}
	toVersion, err := strconv.Atoi(r.URL.Query().Get("to"))
	if err != nil || toVersion <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid to version"))
		return
	}
	_, fromDef, err := s.Store.TreeVersionByNumber(r.Context(), r.PathValue("slug"), fromVersion)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	_, toDef, err := s.Store.TreeVersionByNumber(r.Context(), r.PathValue("slug"), toVersion)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"from_version": fromVersion,
		"to_version":   toVersion,
		"diff":         treeDefinitionDiff(fromDef, toDef),
	})
}

func (s Server) handleTreeVersionRestore(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	version, err := strconv.Atoi(r.PathValue("version"))
	if err != nil || version <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid version"))
		return
	}
	_, def, err := s.Store.TreeVersionByNumber(r.Context(), r.PathValue("slug"), version)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if err := s.Store.SaveTreeDefinition(r.Context(), def); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tree, err := s.Store.TreeBySlug(r.Context(), def.Slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"restored_version": version,
		"tree":             s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0],
	})
}

func buildTreeDefinition(slug, title, description string, seedCodes, prioritySkills []string, items []tgoResponse) (domain.TGOTreeDefinition, error) {
	if strings.TrimSpace(slug) == "" || strings.TrimSpace(title) == "" || len(items) == 0 {
		return domain.TGOTreeDefinition{}, fmt.Errorf("slug, title, and at least one TGO are required")
	}
	def := domain.TGOTreeDefinition{
		Slug:           slug,
		Title:          title,
		Description:    description,
		SeedCodes:      append([]string(nil), seedCodes...),
		PrioritySkills: append([]string(nil), prioritySkills...),
	}
	for _, item := range items {
		def.TGOs = append(def.TGOs, domain.TGO{
			Code:          item.Code,
			Title:         item.Title,
			Description:   item.Description,
			Stage:         item.Stage,
			StageOrder:    item.StageOrder,
			Prerequisites: append([]string(nil), item.Prerequisites...),
			MasteryHint:   item.MasteryHint,
			ProgressMode:  item.ProgressMode,
		})
	}
	if len(def.SeedCodes) == 0 && len(def.TGOs) >= 3 {
		def.SeedCodes = []string{def.TGOs[0].Code, def.TGOs[1].Code, def.TGOs[2].Code}
	}
	return def, nil
}

func treeDefinitionDiff(fromDef, toDef domain.TGOTreeDefinition) map[string]any {
	fromMap := map[string]domain.TGO{}
	toMap := map[string]domain.TGO{}
	for _, tgo := range fromDef.TGOs {
		fromMap[tgo.Code] = tgo
	}
	for _, tgo := range toDef.TGOs {
		toMap[tgo.Code] = tgo
	}
	var added []string
	var removed []string
	var changed []string
	for code, tgo := range toMap {
		if _, ok := fromMap[code]; !ok {
			added = append(added, code)
			continue
		}
		if fromMap[code].Title != tgo.Title || fromMap[code].Description != tgo.Description || fromMap[code].Stage != tgo.Stage || fromMap[code].StageOrder != tgo.StageOrder || strings.Join(fromMap[code].Prerequisites, ",") != strings.Join(tgo.Prerequisites, ",") || fromMap[code].MasteryHint != tgo.MasteryHint {
			changed = append(changed, code)
		}
	}
	for code := range fromMap {
		if _, ok := toMap[code]; !ok {
			removed = append(removed, code)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)
	return map[string]any{
		"title_changed":           fromDef.Title != toDef.Title,
		"description_changed":     fromDef.Description != toDef.Description,
		"seed_codes_changed":      strings.Join(fromDef.SeedCodes, ",") != strings.Join(toDef.SeedCodes, ","),
		"priority_skills_changed": strings.Join(fromDef.PrioritySkills, ",") != strings.Join(toDef.PrioritySkills, ","),
		"added_tgos":              added,
		"removed_tgos":            removed,
		"changed_tgos":            changed,
	}
}

func (s Server) toTreeResponses(ctx context.Context, trees []domain.TGOTree, includeTGOs bool) []treeResponse {
	out := make([]treeResponse, 0, len(trees))
	for _, tree := range trees {
		item := treeResponse{
			ID:          tree.ID,
			Slug:        tree.Slug,
			Title:       tree.Title,
			Description: tree.Description,
			CreatedAt:   db.Since(tree.CreatedAt),
		}
		if includeTGOs {
			if def, err := s.Store.TreeDefinitionBySlug(ctx, tree.Slug); err == nil {
				item.SeedCodes = append([]string(nil), def.SeedCodes...)
				item.PrioritySkills = append([]string(nil), def.PrioritySkills...)
				item.TGOs = toTGOResponses(def.TGOs)
			}
		}
		out = append(out, item)
	}
	return out
}

func (s Server) applyGeneratedTreeProfileDisplay(ctx context.Context, appContext session.Context, tree domain.TGOTree, response treeResponse) treeResponse {
	profile, err := s.Store.OnboardingProfileByEnrollmentID(ctx, appContext.EnrollmentID)
	if err != nil || profile.GeneratedTreeSlug != tree.Slug {
		return response
	}
	user, err := s.Store.UserBySlug(ctx, appContext.UserSlug)
	if err != nil {
		return response
	}
	response.Title, response.Description = domain.GeneratedTreeDisplay(user.Name, profile, tree.Title, tree.Description)
	response.PrioritySkills = nil
	return response
}
