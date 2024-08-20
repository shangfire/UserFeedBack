package dto

type FeedBackFile struct {
	Filename      string
	FilePathOnOss string
	FileSize      int64
}

type FeedbackDTO struct {
	ImpactedModule     string         `json:"impactedModule"`
	OccurringFrequency string         `json:"occurringFrequency"`
	BugDescription     string         `json:"bugDescription"`
	ReproduceSteps     string         `json:"reproduceSteps"`
	UserInfo           string         `json:"userInfo"`
	ProcessInfo        *string        `json:"processInfo,omitempty"`
	Email              string         `json:"email"`
	Files              []FeedBackFile `json:"files"`
}
