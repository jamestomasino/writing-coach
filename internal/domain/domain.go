package domain

import "time"

type WriterProfile struct {
	ID              int64
	Name            string
	AestheticTarget string
	CreatedAt       time.Time
}

type CurriculumState struct {
	ID              int64
	CurrentFocus    string
	DifficultyLevel int
	LastReviewID    int64
	UpdatedAt       time.Time
}

type Exercise struct {
	ID              int64
	Title           string
	Brief           string
	Constraints     []string
	FocusSkills     []string
	SuccessCriteria []string
	GenerationKind  string
	ProviderNote    string
	CreatedAt       time.Time
}

type Submission struct {
	ID         int64
	ExerciseID int64
	Content    string
	WordCount  int
	CreatedAt  time.Time
}

type Review struct {
	ID               int64
	SubmissionID     int64
	Summary          string
	Strengths        []string
	Weaknesses       []string
	AnalyzerFindings []string
	NextFocus        string
	ReviewKind       string
	ProviderNote     string
	CreatedAt        time.Time
	MetricWordCount  int
}

type SkillScore struct {
	SubmissionID int64
	Skill        string
	Score        int
}
