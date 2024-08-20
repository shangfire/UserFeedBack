package dbwrapper

import (
	"UserFeedBack/configwrapper"
	"UserFeedBack/dto"
	"UserFeedBack/logwrapper"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// FileInfo 结构体，用于存储文件信息
type FileInfo struct {
	OriginalName string
	FileType     string
	FileSize     int64
	ServerPath   string
}

// FormSubmission 结构体，用于存储表单提交信息
type FormSubmission struct {
	Title      string
	Content    string
	FileInfos  []FileInfo
	SubmitTime time.Time
}

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
		createTabFeedBack := `  
		CREATE TABLE IF NOT EXISTS feedback (  
			feedback_id INT AUTO_INCREMENT PRIMARY KEY,  
			bug_description TEXT NOT NULL, 
			impacted_module TEXT NOT NULL, 
			occurring_frequency TEXT NOT NULL, 
			reproduce_steps TEXT NOT NULL,  
			user_info TEXT, 
			process_info TEXT, 
			email TEXT 
		);  
		`

		if _, err := instance.Exec(createTabFeedBack); err != nil {
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
    		FOREIGN KEY (feedback_id) REFERENCES Feedback(feedback_id) 
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
func InsertFileSubmission(feedBack dto.FeedbackDTO) error {
	// Start a transaction
	tx, err := instance.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert feedback data
	result, err := tx.Exec("INSERT INTO feedback (bug_description, impacted_module, occurring_frequency, reproduce_steps, user_info, email) VALUES (?, ?, ?, ?, ?, ?)",
		feedBack.BugDescription,
		feedBack.ImpactedModule,
		feedBack.OccurringFrequency,
		feedBack.ReproduceSteps,
		feedBack.UserInfo,
		feedBack.Email)
	if err != nil {
		return err
	}

	// Get the last inserted ID
	feedbackID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Insert file data
	for _, fileInfo := range feedBack.Files {
		_, err = tx.Exec("INSERT INTO file (feedback_id, file_name, file_path, file_size) VALUES (?, ?, ?, ?)",
			feedbackID, fileInfo.Filename, fileInfo.FilePathOnOss, fileInfo.FileSize)
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
func QueryFeedback(pageIndex int, pageSize int) ([]dto.FeedbackDTO, error) {
	var (
		query     string
		result    []dto.FeedbackDTO
		resultMap map[int]*dto.FeedbackDTO
	)

	query = `  
        SELECT  
            f.feedback_id, f.bug_description, f.impacted_module, f.occurring_frequency, f.reproduce_steps, f.user_info, f.process_info, f.email,  
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
		return nil, err
	}
	defer rows.Close()

	result = []dto.FeedbackDTO{}
	resultMap = make(map[int]*dto.FeedbackDTO)

	// 处理查询结果
	for rows.Next() {
		var (
			feedbackID         int
			bugDescription     string
			impactedModule     string
			occurringFrequency string
			reproduceSteps     string
			userInfo           string
			processInfo        string
			email              string
			filename           string
			filePathOnOss      string
			fileSize           int64
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
			&filename,
			&filePathOnOss,
			&fileSize,
		)
		if err != nil {
			return nil, err
		}

		if feedback, exists := resultMap[feedbackID]; exists {
			feedback.Files = append(feedback.Files, dto.FeedBackFile{
				Filename:      filename,
				FilePathOnOss: filePathOnOss,
				FileSize:      fileSize,
			})
		} else {
			files := []dto.FeedBackFile{}
			files = append(files, dto.FeedBackFile{
				Filename:      filename,
				FilePathOnOss: filePathOnOss,
				FileSize:      fileSize,
			})

			resultMap[feedbackID] = &dto.FeedbackDTO{
				ImpactedModule:     impactedModule,
				OccurringFrequency: occurringFrequency,
				BugDescription:     bugDescription,
				ReproduceSteps:     reproduceSteps,
				UserInfo:           userInfo,
				ProcessInfo:        &processInfo,
				Email:              email,
				Files:              files,
			}
		}
	}

	for _, feedback := range resultMap {
		result = append(result, *feedback)
	}

	return result, nil
}
