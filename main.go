/*
 * @Author: shanghanjin
 * @Date: 2024-08-12 11:38:02
 * @LastEditTime: 2024-09-05 10:18:23
 * @FilePath: \UserFeedBack\main.go
 * @Description:main
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
	"strconv"

	"github.com/sirupsen/logrus"
)

/**
 * @description: 汇报反馈接口
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func reportFeedback(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析body
	var reqBody dto.FeedbackUpload
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
	dbwrapper.InsertFeedback(reqBody)

	// 响应客户端已完成
	fmt.Fprintf(w, "Files uploaded successfully")
}

/**
 * @description: 查询反馈接口
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func queryFeedback(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pageIndexStr := r.URL.Query().Get("pageIndex")
	pageSizeStr := r.URL.Query().Get("pageSize")

	var pageIndex int
	if pageIndexStr == "" {
		pageIndex = 0
	} else {
		pageIndex, _ = strconv.Atoi(pageIndexStr)
		if pageIndex < 0 {
			pageIndex = 0
		}
	}

	var pageSize int
	if pageSizeStr == "" {
		pageSize = 10
	} else {
		pageSize, _ = strconv.Atoi(pageSizeStr)
		if pageSize < 10 {
			pageSize = 10
		}
	}

	// 查询数据库
	feedbacks, err := dbwrapper.QueryFeedback(pageIndex, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 写入查询结果
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(feedbacks)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

/**
 * @description: 查询上传文件保存路径
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func queryUploadSavePath(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析body
	type RequestBody struct {
		Files []string `json:"files"`
	}
	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// OSS生成上传路径
	respBody, err := osswrapper.GenerateSecurityToken(reqBody.Files)
	if err != nil {
		http.Error(w, "Failed to generate security token", http.StatusInternalServerError)
		return
	}

	// 将响应数据返回给客户端
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(respBody)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

/**
 * @description: 删除反馈接口
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func deleteFeedback(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析body
	type RequestBody struct {
		FeedBackIDs []int `json:"feedbackID"`
	}
	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// 查询关联的文件
	queryResult := dbwrapper.QueryRelatedFilesByFeedbackID(reqBody.FeedBackIDs)
	var feedbackIDs []int
	var ossFiles []string
	for _, item := range queryResult {
		feedbackIDs = append(feedbackIDs, item.FeedbackID)
		ossFiles = append(ossFiles, item.FileOssPath...)
	}

	// 在oss上删除文件
	osswrapper.DeleteFileOnOssByPath(ossFiles)

	// 数据库删除记录
	dbwrapper.DeleteFeedbackByID(feedbackIDs)

	// 响应客户端已完成
	fmt.Fprintf(w, "Feedback delete successfully")
}

func main() {
	// 初始化日志库
	if err := logwrapper.Init("./log/log.log", logrus.DebugLevel); err != nil {
		logwrapper.Logger.Fatal(err)
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
	}

	// 提供浏览页面的服务
	queryFS := http.FileServer(http.Dir("./html/query"))
	http.Handle("/query/", http.StripPrefix("/query", queryFS))

	// 提供vue测试项目页面的服务
	vueFS := http.FileServer(http.Dir("./dist"))
	http.Handle("/", vueFS)

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
