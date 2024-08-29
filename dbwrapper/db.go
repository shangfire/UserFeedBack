/*
 * @Author: shanghanjin
 * @Date: 2024-08-25 20:51:47
 * @LastEditTime: 2024-08-27 19:22:11
 * @FilePath: \UserFeedBack\dbwrapper\db.go
 * @Description:1
 */
package dbwrapper

import (
	"UserFeedBack/configwrapper"
	"UserFeedBack/dto"
	"UserFeedBack/logwrapper"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	instance *sql.DB
	once     sync.Once
)

// InitDB 初始化数据库连接
// 方法名需要大写才能被外部包调用
func InitDB() {
	once.Do(func() {
		var err error
		// 连接到 MySQL 数据库
		address := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			configwrapper.Cfg.Database.User,
			configwrapper.Cfg.Database.Password,
			configwrapper.Cfg.Database.Host,
			configwrapper.Cfg.Database.Port,
			configwrapper.Cfg.Database.Schema)
		instance, err = sql.Open("mysql", address)
		if err != nil {
			logwrapper.Logger.Fatalf("Failed to connect to database: %v", err)
		}

		// 检查连接是否成功
		if err = instance.Ping(); err != nil {
			logwrapper.Logger.Fatalf("Failed to ping database: %v", err)
		}

		// 检查 FeedBack 表是否存在，如果不存在则创建它
		createTabFeedback := `
		CREATE TABLE IF NOT EXISTS feedback (
			feedback_id INT AUTO_INCREMENT PRIMARY KEY,
			bug_description TEXT NOT NULL,
			impacted_module TEXT NOT NULL,
			occurring_frequency INT NOT NULL,
			reproduce_steps TEXT NOT NULL,
			user_info TEXT,
			process_info TEXT,
			email TEXT,
			app_version TEXT,
			time_stamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`

		if _, err := instance.Exec(createTabFeedback); err != nil {
			logwrapper.Logger.Fatalf("Failed to create table: %v", err)
		}

		// 检查 File 表是否存在，如果不存在则创建它
		createTabFile := `
		CREATE TABLE IF NOT EXISTS file (
	    	file_id INT AUTO_INCREMENT PRIMARY KEY,
			feedback_id INT,
			file_name VARCHAR(255) NOT NULL,
			file_path VARCHAR(255) NOT NULL,
			file_size BIGINT,
			FOREIGN KEY (feedback_id) REFERENCES feedback(feedback_id) ON DELETE CASCADE
		);
		`

		if _, err := instance.Exec(createTabFile); err != nil {
			logwrapper.Logger.Fatalf("Failed to create table: %v", err)
		}
	})
}

// GetDB 返回数据库连接的实例
func GetDB() *sql.DB {
	if instance == nil {
		logwrapper.Logger.Fatal("Database not initialized")
	}
	return instance
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	return instance.Close()
}

// 提交数据到数据库
func InsertFileSubmission(feedback dto.FeedbackUpload) error {
	// Start a transaction
	tx, err := instance.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert feedback data
	result, err := tx.Exec("INSERT INTO feedback (bug_description, impacted_module, occurring_frequency, reproduce_steps, user_info, process_info, email, app_version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		feedback.BugDescription,
		feedback.ImpactedModule,
		feedback.OccurringFrequency,
		feedback.ReproduceSteps,
		feedback.UserInfo,
		"",
		feedback.Email,
		feedback.AppVersion)
	if err != nil {
		return err
	}

	// Get the last inserted ID
	feedbackID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Insert file data
	for _, fileInfo := range feedback.Files {
		_, err = tx.Exec("INSERT INTO file (feedback_id, file_name, file_path, file_size) VALUES (?, ?, ?, ?)",
			feedbackID, fileInfo.FileName, fileInfo.FilePathOnOss, fileInfo.FileSize)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// 查询所有记录信息
func QueryFeedback(pageIndex int, pageSize int) (dto.FeedbackQueryAll, error) {
	var (
		query      string
		realResult dto.FeedbackQueryAll
		result     []dto.FeedbackQueryOne
		resultMap  map[int]*dto.FeedbackQueryOne
	)

	// 查询feedback表的总条数
	var totalCount int
	err := instance.QueryRow("SELECT COUNT(*) FROM feedback").Scan(&totalCount)
	if err != nil {
		return realResult, err
	}

	query = `
	    SELECT
	        f.feedback_id, f.bug_description, f.impacted_module, f.occurring_frequency, f.reproduce_steps, f.user_info, f.process_info, f.email, f.app_version, f.time_stamp,
	        fl.file_name, fl.file_path, fl.file_size
	    FROM
	        feedback f
	    LEFT JOIN
	        file fl ON f.feedback_id = fl.feedback_id
	    ORDER BY
	        f.feedback_id
	`

	// 如果pageIndex不是-1，则添加LIMIT和OFFSET进行分页
	if pageIndex != -1 {
		offset := pageIndex * pageSize
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	rows, err := instance.Query(query)
	if err != nil {
		return realResult, err
	}
	defer rows.Close()

	result = []dto.FeedbackQueryOne{}
	resultMap = make(map[int]*dto.FeedbackQueryOne)

	// 处理查询结果
	for rows.Next() {
		var (
			feedbackID         int
			bugDescription     string
			impactedModule     string
			occurringFrequency int
			reproduceSteps     string
			userInfo           string
			processInfo        string
			email              string
			appVersion         string
			timeStamp          time.Time
			filename           sql.NullString
			filePathOnOss      sql.NullString
			fileSize           sql.NullInt64
		)

		err = rows.Scan(
			&feedbackID,
			&impactedModule,
			&occurringFrequency,
			&bugDescription,
			&reproduceSteps,
			&userInfo,
			&processInfo,
			&email,
			&appVersion,
			&timeStamp,
			&filename,
			&filePathOnOss,
			&fileSize,
		)
		if err != nil {
			return realResult, err
		}

		// 如果已经有了该feedbackID的记录，则追加文件信息
		if feedback, exists := resultMap[feedbackID]; exists {
			if filename.Valid {
				feedback.Files = append(feedback.Files, dto.FeedbackFile{
					FileName:      filename.String,
					FilePathOnOss: "https://" + configwrapper.Cfg.Oss.BucketName + "." + configwrapper.Cfg.Oss.OssEndpoint + "/" + filePathOnOss.String,
					FileSize:      fileSize.Int64,
				})
			}
		} else { // 如果还没有该feedbackID的记录，则创建新的记录
			files := []dto.FeedbackFile{}
			if filename.Valid {
				files = append(files, dto.FeedbackFile{
					FileName:      filename.String,
					FilePathOnOss: "https://" + configwrapper.Cfg.Oss.BucketName + "." + configwrapper.Cfg.Oss.OssEndpoint + "/" + filePathOnOss.String,
					FileSize:      fileSize.Int64,
				})
			}

			resultMap[feedbackID] = &dto.FeedbackQueryOne{
				FeedbackID:         feedbackID,
				AppVersion:         appVersion,
				TimeStamp:          timeStamp.UnixMilli(),
				ImpactedModule:     impactedModule,
				OccurringFrequency: occurringFrequency,
				BugDescription:     bugDescription,
				ReproduceSteps:     reproduceSteps,
				UserInfo:           userInfo,
				ProcessInfo:        processInfo,
				Email:              email,
				Files:              files,
			}
		}
	}

	for _, feedback := range resultMap {
		result = append(result, *feedback)
	}

	realResult.PageData = result
	realResult.TotalSize = totalCount
	realResult.CurrentPageIndex = pageIndex

	return realResult, nil
}

type FeedbackRelatedFile struct {
	FeedbackID  int
	FileOssPath []string
}

func QueryRelatedFilesByFeedbackID(feedbackIDs []int) []FeedbackRelatedFile {
	var feedbackRelatedFiles []FeedbackRelatedFile

	if len(feedbackIDs) == 0 {
		return feedbackRelatedFiles
	}

	query := "SELECT file_path FROM file WHERE feedback_id = ?"

	for _, feedbackID := range feedbackIDs {
		feedbackRelatedFiles = append(feedbackRelatedFiles, FeedbackRelatedFile{
			FeedbackID:  feedbackID,
			FileOssPath: []string{},
		})

		rows, err := instance.Query(query, feedbackID)
		if err != nil {
			logwrapper.Logger.Error(err)
			return feedbackRelatedFiles
		}
		defer rows.Close()

		for rows.Next() {
			var filePathOnOss string

			err = rows.Scan(&filePathOnOss)
			if err != nil {
				logwrapper.Logger.Error(err)
				return feedbackRelatedFiles
			}

			feedbackRelatedFiles[len(feedbackRelatedFiles)-1].FileOssPath = append(
				feedbackRelatedFiles[len(feedbackRelatedFiles)-1].FileOssPath,
				filePathOnOss,
			)
		}
	}

	return feedbackRelatedFiles
}

func DeleteFeedbackByID(feedbackIDs []int) {
	var builder strings.Builder
	builder.WriteString("DELETE FROM feedback WHERE feedback_id IN (")

	for i, id := range feedbackIDs {
		builder.WriteString(fmt.Sprintf("%d", id))
		if i < len(feedbackIDs)-1 {
			builder.WriteString(",")
		}
	}

	builder.WriteString(")")

	instance.Exec(builder.String())
}
