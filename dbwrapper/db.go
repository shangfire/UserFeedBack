/*
 * @Author: shanghanjin
 * @Date: 2024-08-25 20:51:47
 * @LastEditTime: 2024-09-05 10:21:22
 * @FilePath: \UserFeedBack\dbwrapper\db.go
 * @Description: 数据库操作封装
 */
package dbwrapper

import (
	"UserFeedBack/configwrapper"
	"UserFeedBack/dto"
	"UserFeedBack/logwrapper"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// 数据库单例
	db *sql.DB
	// 单例标志
	once sync.Once
)

/**
 * @description: 初始化数据库连接
 * @return {*}
 */
func InitDB() {
	once.Do(func() {
		var err error
		// 连接到 MySQL 数据库
		address := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			configwrapper.Cfg.Database.User,
			configwrapper.Cfg.Database.Password,
			configwrapper.Cfg.Database.Host,
			configwrapper.Cfg.Database.Port,
			configwrapper.Cfg.Database.Schema)
		db, err = sql.Open("mysql", address)
		if err != nil {
			logwrapper.Logger.Fatalf("Failed to connect to database: %v", err)
		}

		// 检查连接是否成功
		if err = db.Ping(); err != nil {
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

		if _, err := db.Exec(createTabFeedback); err != nil {
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

		if _, err := db.Exec(createTabFile); err != nil {
			logwrapper.Logger.Fatalf("Failed to create table: %v", err)
		}
	})
}

/**
 * @description: 关闭数据库连接
 * @return {*}
 */
func CloseDB() error {
	return db.Close()
}

/**
 * @description: 提交反馈数据到数据库
 * @param {dto.FeedbackUpload} feedback
 * @return {*}
 */
func InsertFeedback(feedback dto.FeedbackUpload) error {
	// 开启事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// 确保失败时能正确回滚
	defer tx.Rollback()

	// 插入反馈数据
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

	// 获取到最后插入的主键ID，也就是feedbackID
	feedbackID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// 插入文件数据
	for _, fileInfo := range feedback.Files {
		_, err = tx.Exec("INSERT INTO file (feedback_id, file_name, file_path, file_size) VALUES (?, ?, ?, ?)",
			feedbackID, fileInfo.FileName, fileInfo.FilePathOnOss, fileInfo.FileSize)
		if err != nil {
			return err
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

/**
 * @description: 查询所有反馈信息
 * @param {int} pageIndex 分页索引
 * @param {int} pageSize 分页大小
 * @return {*}
 */
func QueryFeedback(pageIndex int, pageSize int) (dto.FeedbackQueryAll, error) {
	var (
		query      string
		realResult dto.FeedbackQueryAll
		result     []dto.FeedbackQueryOne
		resultMap  map[int]*dto.FeedbackQueryOne
	)

	// 查询feedback表的总条数
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM feedback").Scan(&totalCount)
	if err != nil {
		return realResult, err
	}

	// 越界时修正为最后的index
	if pageIndex*pageSize > totalCount {
		pageIndex = (totalCount+pageSize-1)/pageSize - 1
	}

	// 按分页大小和索引查询对应的反馈信息
	var limit string
	if pageIndex >= 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, pageIndex*pageSize)
	}

	query = fmt.Sprintf(`  
        SELECT  
            f.feedback_id, f.bug_description, f.impacted_module, f.occurring_frequency, f.reproduce_steps, f.user_info, f.process_info, f.email, f.app_version, f.time_stamp,  
            fl.file_name, fl.file_path, fl.file_size  
        FROM  
            (SELECT feedback_id FROM feedback ORDER BY feedback_id%s) AS sub  
        JOIN  
            feedback f ON sub.feedback_id = f.feedback_id  
        LEFT JOIN  
            file fl ON f.feedback_id = fl.feedback_id  
        ORDER BY  
            f.feedback_id;  
    `, limit)

	rows, err := db.Query(query)
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
			&bugDescription,
			&impactedModule,
			&occurringFrequency,
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

	// map是无序的，这里需要对key排序后遍历
	keys := make([]int, 0, len(resultMap))
	for k := range resultMap {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	for _, k := range keys {
		result = append(result, *resultMap[k])
	}

	// 填充反馈结果
	realResult.PageData = result
	realResult.TotalSize = totalCount
	realResult.CurrentPageIndex = pageIndex

	return realResult, nil
}

type FeedbackRelatedFile struct {
	FeedbackID  int
	FileOssPath []string
}

/**
 * @description: 查询feedbackid相关的文件
 * @param {[]int} feedbackIDs 要查询的feedbackid数组
 * @return {*}
 */
func QueryRelatedFilesByFeedbackID(feedbackIDs []int) []FeedbackRelatedFile {
	var feedbackRelatedFiles []FeedbackRelatedFile

	if len(feedbackIDs) == 0 {
		return feedbackRelatedFiles
	}

	// 查询feedbackid相关的文件信息
	query := "SELECT file_path FROM file WHERE feedback_id = ?"

	for _, feedbackID := range feedbackIDs {
		feedbackRelatedFiles = append(feedbackRelatedFiles, FeedbackRelatedFile{
			FeedbackID:  feedbackID,
			FileOssPath: []string{},
		})

		rows, err := db.Query(query, feedbackID)
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

			// 填充结果到返回值
			feedbackRelatedFiles[len(feedbackRelatedFiles)-1].FileOssPath = append(
				feedbackRelatedFiles[len(feedbackRelatedFiles)-1].FileOssPath,
				filePathOnOss,
			)
		}
	}

	return feedbackRelatedFiles
}

/**
 * @description: 删除feedback表
 * @param {[]int} feedbackIDs feedbackid数组
 * @return {*}
 */
func DeleteFeedbackByID(feedbackIDs []int) {
	// 构造删除语句
	var builder strings.Builder
	builder.WriteString("DELETE FROM feedback WHERE feedback_id IN (")

	for i, id := range feedbackIDs {
		builder.WriteString(fmt.Sprintf("%d", id))
		if i < len(feedbackIDs)-1 {
			builder.WriteString(",")
		}
	}

	builder.WriteString(")")

	// 执行删除语句
	db.Exec(builder.String())
}
