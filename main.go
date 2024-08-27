/*
 * @Author: shanghanjin
 * @Date: 2024-08-12 11:38:02
 * @LastEditTime: 2024-08-27 19:11:12
 * @FilePath: \UserFeedBack\main.go
 * @Description:
 */
package main

import (
	"UserFeedBack/configwrapper"
	"UserFeedBack/dbwrapper"
	"UserFeedBack/dto"
	"UserFeedBack/logwrapper"
	"UserFeedBack/osswrapper"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

/**
 * @description: 汇报反馈接口
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func reportFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析body
	var reqBody dto.FeedbackDTO
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	if reqBody.ImpactedModule == "" || reqBody.BugDescription == "" || reqBody.ReproduceSteps == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// 相关内容写入数据库
	dbwrapper.InsertFileSubmission(reqBody)

	// 响应客户端已完成
	fmt.Fprintf(w, "Files uploaded successfully")
}

/**
 * @description: 查询文件接口
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func queryFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		PageIndex int `json:"pageIndex"`
		PageSize  int `json:"pageSize"`
	}

	// 解析body
	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	feedbacks, err := dbwrapper.QueryFeedback(reqBody.PageIndex, reqBody.PageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedbacks)
}

/**
 * @description: 查询上传文件保存路径
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func queryUploadSavePath(w http.ResponseWriter, r *http.Request) {
	// RequestBody 是从客户端接收的数据结构
	type RequestBody struct {
		Files []string `json:"files"`
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	respBody, err := osswrapper.GenerateSecurityToken(reqBody.Files)
	if err != nil {
		http.Error(w, "Failed to generate security token", http.StatusInternalServerError)
		return
	}

	// 将响应数据返回给客户端
	err = json.NewEncoder(w).Encode(respBody)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

/**
 * @description:
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func deleteFeedback(w http.ResponseWriter, r *http.Request) {
	// RequestBody 是从客户端接收的数据结构
	type RequestBody struct {
		FeedBackIDs []int `json:"feedbackID"`
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	queryResult := dbwrapper.QueryRelatedFilesByFeedbackID(reqBody.FeedBackIDs)
	var feedbackIDs []int
	var ossFiles []string
	for _, item := range queryResult {
		feedbackIDs = append(feedbackIDs, item.FeedbackID)
		for _, itemPath := range item.FileOssPath {
			ossFiles = append(ossFiles, itemPath)
		}
	}

	osswrapper.DeleteFileOnOssByPath(ossFiles)
	dbwrapper.DeleteFeedbackByID(feedbackIDs)

	// 响应客户端已完成
	fmt.Fprintf(w, "Feedback delete successfully")
}

func main() {
	// 初始化日志库
	if err := logwrapper.Init("./log/log.log", logrus.DebugLevel); err != nil {
		logwrapper.Logger.Fatal(err)
		return
	}

	// 初始化配置
	if err := configwrapper.Init("./config"); err != nil {
		logwrapper.Logger.Fatal(err)
	}

	// 初始化数据库
	dbwrapper.InitDB()
	defer dbwrapper.CloseDB()

	// 初始化oss
	if err := osswrapper.Init(); err != nil {
		logwrapper.Logger.Fatal(err)
		return
	}

	// 提供上传页面的服务
	uploadFS := http.FileServer(http.Dir("./html/upload"))
	http.Handle("/upload/", http.StripPrefix("/upload", uploadFS))

	// 提供浏览页面的服务
	queryFS := http.FileServer(http.Dir("./html/query"))
	http.Handle("/query/", http.StripPrefix("/query", queryFS))

	// 设置各接口响应函数
	http.HandleFunc("/api/queryFeedback", queryFeedback)
	http.HandleFunc("/api/reportFeedback", reportFeedback)
	http.HandleFunc("/api/queryUploadSavePath", queryUploadSavePath)
	http.HandleFunc("/api/deleteFeedback", deleteFeedback)

	logwrapper.Logger.Info("Server is running")

	// 启动服务
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
