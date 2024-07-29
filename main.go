package main

import (
	"UserFeedBack/dbwrapper"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	// 检查请求是否为multipart/form-data
	if r.Method != "POST" || !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		log.Printf("Invalid request type or content type")
		// 回传400错误以及提示信息给client，下不赘述
		http.Error(w, "Invalid request type or content type", http.StatusBadRequest)
		return
	}

	// 解析表单
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 最大32MB
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	submission := dbwrapper.FormSubmission{}

	// 获取title和content字段
	title := r.FormValue("title")
	content := r.FormValue("content")

	// 检查title和content字段是否存在
	if title == "" || content == "" {
		log.Printf("Missing title or content fields")
		http.Error(w, "Missing title or content fields", http.StatusBadRequest)
		return
	}

	submission.Title = title
	submission.Content = content
	submission.SubmitTime = time.Now()

	// 获取文件字段
	files, ok := r.MultipartForm.File["files"]
	if !ok || len(files) == 0 {
		// 如果不需要文件也可以继续处理，这里只是记录日志
		log.Printf("No files were uploaded")
		// 但你可能想直接返回或继续处理其他逻辑
		// ...
	}

	// 设置文件保存目录（确保这个目录存在）
	saveDir := "./uploads"
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
		return
	}

	// 遍历文件并保存
	for _, file := range files {
		// 生成唯一文件名
		timestamp := time.Now().Format("20060102150405")
		fileExt := filepath.Ext(file.Filename)
		newFileName := fmt.Sprintf("%s-%d%s", timestamp, time.Now().UnixNano(), fileExt)
		filePath := filepath.Join(saveDir, newFileName)

		// 打开文件以进行保存
		dst, err := os.Create(filePath)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// 从请求中读取文件内容并保存到本地
		src, err := file.Open()
		if err != nil {
			log.Printf("Error opening uploaded file: %v", err)
			http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
			return
		}
		defer src.Close()

		if _, err := io.Copy(dst, src); err != nil {
			log.Printf("Error saving file: %v", err)
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}

		fileInfo := dbwrapper.FileInfo{}
		fileInfo.OriginalName = file.Filename
		fileInfo.FileSize = file.Size
		fileInfo.FileType = file.Header.Get("Content-Type")
		fileInfo.ServerPath = filePath

		submission.FileInfos = append(submission.FileInfos, fileInfo)
	}

	dbwrapper.InsertFileSubmission(submission)

	// 响应
	fmt.Fprintf(w, "Files uploaded successfully")
}

func queryFileSubmission(w http.ResponseWriter, r *http.Request) {
	feedbacks, err := dbwrapper.QueryFileSubmission()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedbacks)
}

func main() {
	// 初始化数据库
	dbwrapper.InitDB()
	defer dbwrapper.CloseDB()

	// 启动HTTP服务器
	http.HandleFunc("/feedback", queryFileSubmission)
	http.HandleFunc("/upload", uploadFile)
	fmt.Println("Server is running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
