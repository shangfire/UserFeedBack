/*
 * @Author: shanghanjin
 * @Date: 2024-08-20 15:00:52
 * @LastEditTime: 2024-08-29 11:39:05
 * @FilePath: \UserFeedBack\dto\dto.go
 * @Description: 公共结构体
 */
package dto

type FeedbackFile struct {
	FileName      string `json:"fileName"`
	FilePathOnOss string `json:"filePathOnOss"`
	FileSize      int64  `json:"fileSize"`
}

type FeedbackUpload struct {
	AppVersion         string         `json:"appVersion"`
	ImpactedModule     string         `json:"impactedModule"`
	OccurringFrequency int            `json:"occurringFrequency"`
	BugDescription     string         `json:"bugDescription"`
	ReproduceSteps     string         `json:"reproduceSteps"`
	UserInfo           string         `json:"userInfo"`
	Email              string         `json:"email"`
	Files              []FeedbackFile `json:"files"`
}

type FeedbackQueryAll struct {
	TotalSize        int                `json:"totalSize"`
	CurrentPageIndex int                `json:"currentPageIndex"`
	PageData         []FeedbackQueryOne `json:"pageData"`
}

type FeedbackQueryOne struct {
	FeedbackID         int            `json:"feedbackID"`
	AppVersion         string         `json:"appVersion"`
	TimeStamp          int64          `json:"timeStamp"`
	ImpactedModule     string         `json:"impactedModule"`
	OccurringFrequency int            `json:"occurringFrequency"`
	BugDescription     string         `json:"bugDescription"`
	ReproduceSteps     string         `json:"reproduceSteps"`
	UserInfo           string         `json:"userInfo"`
	ProcessInfo        string         `json:"processInfo"`
	Email              string         `json:"email"`
	Files              []FeedbackFile `json:"files"`
}
