/*
 * @Author: shanghanjin
 * @Date: 2024-08-20 15:00:52
 * @LastEditTime: 2024-08-27 14:44:23
 * @FilePath: \UserFeedBack\dto\dto.go
 * @Description:
 */
package dto

type FeedbackFile struct {
	FileName      string `json:"fileName"`
	FilePathOnOss string `json:"filePathOnOss"`
	FileSize      int64  `json:"fileSize"`
}

type QueryFeedback struct {
	TotalSize        int             `json:"totalSize"`
	CurrentPageIndex int             `json:"currentPageIndex"`
	PageData         []FeedbackQuery `json:"pageData"`
}

type FeedbackDTO struct {
	ImpactedModule     string         `json:"impactedModule"`
	OccurringFrequency int            `json:"occurringFrequency"`
	BugDescription     string         `json:"bugDescription"`
	ReproduceSteps     string         `json:"reproduceSteps"`
	UserInfo           string         `json:"userInfo"`
	ProcessInfo        *string        `json:"processInfo,omitempty"`
	Email              string         `json:"email"`
	Files              []FeedbackFile `json:"files"`
}

type FeedbackQuery struct {
	FeedbackID         int            `json:"feedbackID"`
	ImpactedModule     string         `json:"impactedModule"`
	OccurringFrequency int            `json:"occurringFrequency"`
	BugDescription     string         `json:"bugDescription"`
	ReproduceSteps     string         `json:"reproduceSteps"`
	UserInfo           string         `json:"userInfo"`
	ProcessInfo        string         `json:"processInfo"`
	Email              string         `json:"email"`
	Files              []FeedbackFile `json:"files"`
}
