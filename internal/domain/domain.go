package domain

import "time"

type User struct {
	ID        int64
	Slug      string
	Name      string
	CreatedAt time.Time
}

type TGOTree struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	CreatedAt   time.Time
}

type Enrollment struct {
	ID        int64
	UserID    int64
	TreeID    int64
	UserSlug  string
	TreeSlug  string
	CreatedAt time.Time
}

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
	ID                 int64
	UserID             int64
	TreeID             int64
	Title              string
	Brief              string
	Constraints        []string
	FocusSkills        []string
	TGOCodes           []string
	SuccessCriteria    []string
	GenerationKind     string
	ProviderNote       string
	SourceSubmissionID int64
	CreatedAt          time.Time
}

type Submission struct {
	ID                 int64
	UserID             int64
	TreeID             int64
	ExerciseID         int64
	ParentSubmissionID int64
	DraftNumber        int
	Content            string
	WordCount          int
	CreatedAt          time.Time
}

type Review struct {
	ID               int64
	UserID           int64
	TreeID           int64
	SubmissionID     int64
	Summary          string
	Strengths        []string
	Weaknesses       []string
	AnalyzerFindings []string
	TGOAssessments   []TGOAssessment
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
