package domain

import "time"

type User struct {
	ID             int64
	Slug           string
	Name           string
	ActiveTreeSlug string
	CreatedAt      time.Time
}

type AIProviderSettings struct {
	UserID              int64
	Provider            string
	APIKeyEncrypted     string
	APIKeyLast4         string
	BaseURLOverride     string
	PromptModelOverride string
	ReviewModelOverride string
	Enabled             bool
	ValidatedAt         time.Time
	LastValidationError string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type AIProviderEvent struct {
	ID         int64
	UserID     int64
	UserSlug   string
	Provider   string
	Event      string
	Category   string
	StatusCode int
	DetailJSON string
	CreatedAt  time.Time
}

type AIProviderEventCount struct {
	Label string
	Count int
}

type AIProviderEventSummary struct {
	Since               time.Time
	Total               int
	ValidationFailures  int
	ValidationRateLimit int
	Fallbacks           int
	ProviderCounts      []AIProviderEventCount
	EventCounts         []AIProviderEventCount
	CategoryCounts      []AIProviderEventCount
}

type TGOTree struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	CreatedAt   time.Time
}

type TreeVersion struct {
	ID          int64
	TreeID      int64
	TreeSlug    string
	Version     int
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

type UserTrack struct {
	EnrollmentID int64
	TreeID       int64
	TreeSlug     string
	Title        string
	Description  string
	IsActive     bool
	CreatedAt    time.Time
}

type TrackActivitySummary struct {
	AssignmentCount      int
	CurrentAssignment    string
	LatestAssignmentTime time.Time
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
	ClosedAt           time.Time
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
	ID                 int64
	UserID             int64
	TreeID             int64
	SubmissionID       int64
	Summary            string
	Strengths          []string
	Weaknesses         []string
	AnalyzerFindings   []string
	TGOAssessments     []TGOAssessment
	CompletedTGOChecks []TGOAssessment
	Annotations        []ReviewAnnotation
	NextFocus          string
	ReviewKind         string
	ProviderNote       string
	CreatedAt          time.Time
	MetricWordCount    int
	SkillScores        []SkillScore
}

type ReviewAnnotation struct {
	Quote    string
	TGOCode  string
	Category string
	Comment  string
	Severity string
}

type ReviewArtifacts struct {
	ReviewID           int64
	AnalyzerReportJSON string
	RecommendationJSON string
	ComparisonJSON     string
	AnnotationsJSON    string
	CreatedAt          time.Time
}

type SkillScore struct {
	SubmissionID int64
	Skill        string
	Score        int
}

type AIJob struct {
	ID           int64
	UserID       int64
	TreeID       int64
	EnrollmentID int64
	Kind         string
	ResourceKey  string
	ExerciseID   int64
	SubmissionID int64
	ReviewID     int64
	Status       string
	AttemptCount int
	MaxAttempts  int
	LastError    string
	PayloadJSON  string
	ResultJSON   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PlaygroundSession struct {
	ID               int64
	UserID           int64
	TreeID           int64
	Title            string
	Content          string
	WritingLanguage  string
	WritingType      string
	AssignmentFormat string
	CoachingBrief    string
	LatestDraftID    int64
	LatestReviewID   int64
	LatestReviewAt   time.Time
	DraftCount       int
	ReviewCount      int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PlaygroundDraft struct {
	ID            int64
	SessionID     int64
	UserID        int64
	TreeID        int64
	ParentDraftID int64
	Content       string
	WordCount     int
	CreatedAt     time.Time
}

type PlaygroundReview struct {
	ID                 int64
	SessionID          int64
	DraftID            int64
	UserID             int64
	TreeID             int64
	Review             Review
	AnalyzerReportJSON string
	ComparisonJSON     string
	CreatedAt          time.Time
}

type ReviewJob struct {
	ID           int64
	UserID       int64
	TreeID       int64
	EnrollmentID int64
	SubmissionID int64
	ReviewID     int64
	Status       string
	AttemptCount int
	MaxAttempts  int
	LastError    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
