package dbwrapper

import (
	"UserFeedBack/logwrapper"
	"database/sql"
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
		instance, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/new_schema")
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
			title VARCHAR(255) NOT NULL,  
			content TEXT
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
    		file_type VARCHAR(50),  
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

// 提交表单数据到数据库
func InsertFileSubmission(submission FormSubmission) error {
	// Start a transaction
	tx, err := instance.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert feedback data
	result, err := tx.Exec("INSERT INTO feedback (title, content) VALUES (?, ?)", submission.Title, submission.Content)
	if err != nil {
		return err
	}

	// Get the last inserted ID
	feedbackID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Insert file data
	for _, fileInfo := range submission.FileInfos {
		_, err = tx.Exec("INSERT INTO file (feedback_id, file_name, file_path, file_type, file_size) VALUES (?, ?, ?, ?, ?)",
			feedbackID, fileInfo.OriginalName, fileInfo.ServerPath, fileInfo.FileType, fileInfo.FileSize)
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
func QueryFileSubmission() ([]FormSubmission, error) {
	rows, err := instance.Query(`
        SELECT
            f.feedback_id, f.title, f.content,
            fl.file_name, fl.file_path, fl.file_type, fl.file_size
        FROM
            feedback f
        LEFT JOIN
            file fl ON f.feedback_id = fl.feedback_id
        ORDER BY
            f.feedback_id
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feedbacks := []FormSubmission{}
	feedbackMap := make(map[int]*FormSubmission)

	for rows.Next() {
		var feedbackID int
		var title, content, fileName, filePath, fileType sql.NullString
		var fileSize sql.NullInt64

		if err := rows.Scan(&feedbackID, &title, &content, &fileName, &filePath, &fileType, &fileSize); err != nil {
			return nil, err
		}

		if feedback, exists := feedbackMap[feedbackID]; exists {
			if fileName.Valid {
				feedback.FileInfos = append(feedback.FileInfos, FileInfo{
					OriginalName: fileName.String,
					ServerPath:   filePath.String,
					FileType:     fileType.String,
					FileSize:     fileSize.Int64,
				})
			}
		} else {
			files := []FileInfo{}
			if fileName.Valid {
				files = append(files, FileInfo{
					OriginalName: fileName.String,
					ServerPath:   filePath.String,
					FileType:     fileType.String,
					FileSize:     fileSize.Int64,
				})
			}

			feedbackMap[feedbackID] = &FormSubmission{
				Title:     title.String,
				Content:   content.String,
				FileInfos: files,
			}
		}
	}

	for _, feedback := range feedbackMap {
		feedbacks = append(feedbacks, *feedback)
	}

	return feedbacks, nil
}
