package main

import (
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

type treeDoc struct {
	Tree           domain.TGOTreeDefinition
	StageOrder     []string
	StageCounts    map[string]int
	PrereqCount    int
	SeedSet        map[string]bool
	IsPublic       bool
	DiagramRelPath string
	SourceRelPath  string
}

var stageColors = map[string]string{
	"foundation": "#E3F2FD",
	"core":       "#E8F5E9",
	"paragraph":  "#FFF8E1",
	"story":      "#FCE4EC",
	"dialogue":   "#F3E5F5",
	"forms":      "#E0F7FA",
	"revision":   "#ECEFF1",
	"scene":      "#FFF3E0",
	"character":  "#FBE9E7",
	"world":      "#E8EAF6",
	"style":      "#F3E5F5",
	"structure":  "#E0F2F1",
	"insight":    "#F1F8E9",
	"evidence":   "#E8F5E9",
	"voice":      "#FCE4EC",
	"audience":   "#FFFDE7",
	"advanced":   "#EDE7F6",
	"tone":       "#FDE2E4",
	"analysis":   "#E6FCF5",
	"scanning":   "#FFF1E6",
	"examples":   "#E7F5FF",
	"reference":  "#F3F0FF",
	"support":    "#FFF9DB",
	"research":   "#E9FAC8",
}

var fallbackPalette = []string{
	"#E3F2FD",
	"#E8F5E9",
	"#FFF8E1",
	"#FCE4EC",
	"#E0F7FA",
	"#F3E5F5",
	"#E8EAF6",
	"#F1F8E9",
	"#FFF3E0",
	"#ECEFF1",
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fail(err)
	}

	docsDir := filepath.Join(root, "docs")
	diagramDir := filepath.Join(docsDir, "diagrams", "skill-paths")
	if err := os.MkdirAll(diagramDir, 0o755); err != nil {
		fail(err)
	}

	publicSet := domain.PublicBuiltInTreeSlugs()
	var docs []treeDoc
	for _, tree := range domain.BuiltInTrees {
		doc := buildTreeDoc(tree, publicSet[tree.Slug])
		if err := os.WriteFile(filepath.Join(diagramDir, tree.Slug+".puml"), []byte(renderPlantUML(doc)), 0o644); err != nil {
			fail(err)
		}
		docs = append(docs, doc)
	}

	if err := renderSVGs(diagramDir); err != nil {
		fail(err)
	}

	indexPath := filepath.Join(docsDir, "skill-paths.md")
	if err := os.WriteFile(indexPath, []byte(renderIndex(docs)), 0o644); err != nil {
		fail(err)
	}
}

func buildTreeDoc(tree domain.TGOTreeDefinition, isPublic bool) treeDoc {
	stageOrder := make([]string, 0)
	stageSeen := map[string]bool{}
	stageCounts := map[string]int{}
	seedSet := map[string]bool{}
	for _, code := range tree.SeedCodes {
		seedSet[code] = true
	}
	prereqCount := 0
	for _, tgo := range tree.TGOs {
		if !stageSeen[tgo.Stage] {
			stageSeen[tgo.Stage] = true
			stageOrder = append(stageOrder, tgo.Stage)
		}
		stageCounts[tgo.Stage]++
		prereqCount += len(tgo.Prerequisites)
	}
	return treeDoc{
		Tree:           tree,
		StageOrder:     stageOrder,
		StageCounts:    stageCounts,
		PrereqCount:    prereqCount,
		SeedSet:        seedSet,
		IsPublic:       isPublic,
		DiagramRelPath: filepath.ToSlash(filepath.Join("diagrams", "skill-paths", tree.Slug+".svg")),
		SourceRelPath:  filepath.ToSlash(filepath.Join("diagrams", "skill-paths", tree.Slug+".puml")),
	}
}

func renderSVGs(diagramDir string) error {
	paths, err := filepath.Glob(filepath.Join(diagramDir, "*.puml"))
	if err != nil {
		return err
	}
	args := []string{"-tsvg"}
	for _, path := range paths {
		args = append(args, filepath.Base(path))
	}
	cmd := exec.Command("plantuml", args...)
	cmd.Dir = diagramDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func renderIndex(docs []treeDoc) string {
	publicBySlug := make(map[string]treeDoc, len(docs))
	var internalDocs []treeDoc
	for _, doc := range docs {
		if doc.IsPublic {
			publicBySlug[doc.Tree.Slug] = doc
			continue
		}
		internalDocs = append(internalDocs, doc)
	}

	var b strings.Builder
	b.WriteString("# Skill Path UML\n\n")
	b.WriteString("This document is generated from `internal/domain/tree_catalog.go`.\n\n")
	b.WriteString("Each diagram is emitted as PlantUML source plus a rendered SVG so the skill DAGs can be reviewed visually in the repo.\n\n")
	b.WriteString("Regenerate with `go run ./cmd/generate-skill-path-docs`.\n\n")
	b.WriteString("## Legend\n\n")
	b.WriteString("- Node color reflects the node stage within that path.\n")
	b.WriteString("- Seed nodes use a stronger border.\n")
	b.WriteString("- Arrows point from prerequisite to unlocked node.\n")
	b.WriteString("- All diagrams are generated from the validated built-in tree definitions, so they reflect the current DAG structure exactly.\n\n")
	b.WriteString("## Public Tracks\n\n")
	for _, tree := range domain.PublicBuiltInTrees {
		if doc, ok := publicBySlug[tree.Slug]; ok {
			writeTreeSection(&b, doc)
		}
	}
	if len(internalDocs) > 0 {
		b.WriteString("## Internal Templates\n\n")
		for _, doc := range internalDocs {
			writeTreeSection(&b, doc)
		}
	}
	return b.String()
}

func writeTreeSection(b *strings.Builder, doc treeDoc) {
	fmt.Fprintf(b, "### %s\n\n", doc.Tree.Title)
	fmt.Fprintf(b, "- Slug: `%s`\n", doc.Tree.Slug)
	fmt.Fprintf(b, "- Nodes: `%d`\n", len(doc.Tree.TGOs))
	fmt.Fprintf(b, "- Prerequisite edges: `%d`\n", doc.PrereqCount)
	fmt.Fprintf(b, "- Seeds: `%s`\n", joinCodeList(doc.Tree.SeedCodes))
	fmt.Fprintf(b, "- Priority skills: `%s`\n", joinCodeList(doc.Tree.PrioritySkills))
	fmt.Fprintf(b, "- Stage mix: `%s`\n", stageSummary(doc))
	fmt.Fprintf(b, "- UML source: [%s](%s)\n", filepath.Base(doc.SourceRelPath), doc.SourceRelPath)
	fmt.Fprintf(b, "- SVG: [%s](%s)\n\n", filepath.Base(doc.DiagramRelPath), doc.DiagramRelPath)
	fmt.Fprintf(b, "![%s](%s)\n\n", doc.Tree.Title, doc.DiagramRelPath)
}

func joinCodeList(values []string) string {
	return strings.Join(values, "`, `")
}

func stageSummary(doc treeDoc) string {
	parts := make([]string, 0, len(doc.StageOrder))
	for _, stage := range doc.StageOrder {
		parts = append(parts, fmt.Sprintf("%s=%d", stage, doc.StageCounts[stage]))
	}
	return strings.Join(parts, ", ")
}

func renderPlantUML(doc treeDoc) string {
	var b strings.Builder
	b.WriteString("@startuml\n")
	b.WriteString("left to right direction\n")
	b.WriteString("skinparam backgroundColor #FFFFFF\n")
	b.WriteString("skinparam shadowing false\n")
	b.WriteString("skinparam packageStyle rectangle\n")
	b.WriteString("skinparam defaultTextAlignment center\n")
	b.WriteString("skinparam dpi 150\n")
	b.WriteString("skinparam ArrowColor #5C6773\n")
	b.WriteString("skinparam ArrowThickness 1.4\n")
	b.WriteString("skinparam package {\n")
	b.WriteString("  BorderColor #D5DCE3\n")
	b.WriteString("  FontColor #2C3E50\n")
	b.WriteString("  FontStyle bold\n")
	b.WriteString("}\n")
	b.WriteString("skinparam rectangle {\n")
	b.WriteString("  RoundCorner 10\n")
	b.WriteString("  BorderColor #5C6773\n")
	b.WriteString("  FontColor #1F2933\n")
	b.WriteString("}\n")
	b.WriteString("skinparam rectangle<<seed>> {\n")
	b.WriteString("  BorderColor #A61E4D\n")
	b.WriteString("  BorderThickness 3\n")
	b.WriteString("}\n")
	b.WriteString("hide stereotype\n\n")
	fmt.Fprintf(&b, "title %s\\n<color:#5C6773>%s</color>\n\n", escape(doc.Tree.Title), escape(doc.Tree.Slug))

	b.WriteString("legend left\n")
	b.WriteString("|= Marker |= Meaning |\n")
	b.WriteString("| thicker border | seed node |\n")
	for _, stage := range doc.StageOrder {
		fmt.Fprintf(&b, "|<back:%s> %s | stage |\n", stageColor(stage), escape(stage))
	}
	b.WriteString("endlegend\n\n")

	stageGroups := make(map[string][]domain.TGO, len(doc.StageOrder))
	for _, tgo := range doc.Tree.TGOs {
		stageGroups[tgo.Stage] = append(stageGroups[tgo.Stage], tgo)
	}

	for _, stage := range doc.StageOrder {
		fmt.Fprintf(&b, "package \"%s\" %s {\n", escape(stageLabel(stage, doc.StageCounts[stage])), stageColor(stage))
		for _, tgo := range stageGroups[stage] {
			stereotype := ""
			if doc.SeedSet[tgo.Code] {
				stereotype = " <<seed>>"
			}
			fmt.Fprintf(&b, "  rectangle \"%s\" as %s%s %s\n", nodeLabel(tgo, doc.SeedSet[tgo.Code]), aliasFor(tgo.Code), stereotype, stageColor(stage))
		}
		b.WriteString("}\n\n")
	}

	for _, tgo := range doc.Tree.TGOs {
		for _, prereq := range tgo.Prerequisites {
			fmt.Fprintf(&b, "%s -[#5C6773]-> %s\n", aliasFor(prereq), aliasFor(tgo.Code))
		}
	}
	b.WriteString("@enduml\n")
	return b.String()
}

func nodeLabel(tgo domain.TGO, isSeed bool) string {
	title := escape(tgo.Title)
	if isSeed {
		title = "<b>" + title + "</b>"
	}
	return title + "\\n" + fmt.Sprintf("<size:10><color:#5C6773>%s</color></size>", escape(tgo.Code))
}

func stageLabel(stage string, count int) string {
	return fmt.Sprintf("%s (%d)", stage, count)
}

func stageColor(stage string) string {
	if color := stageColors[stage]; color != "" {
		return color
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(stage))
	return fallbackPalette[int(h.Sum32())%len(fallbackPalette)]
}

func aliasFor(code string) string {
	var b strings.Builder
	for _, r := range code {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

func escape(value string) string {
	replacer := strings.NewReplacer(
		`"`, `'`,
		"\n", " ",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(value)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
