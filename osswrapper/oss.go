/*
 * @Author: shanghanjin
 * @Date: 2024-08-13 16:46:39
 * @LastEditTime: 2024-09-05 17:40:27
 * @FilePath: \UserFeedBack\osswrapper\oss.go
 * @Description: oss操作封装
 */
package osswrapper

import (
	zgconfig "UserFeedBack/configwrapper"
	logger "UserFeedBack/logwrapper"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// sts客户端
var stsClient *sts.Client

// oss客户端
var ossClient *oss.Client

/**
 * @description: 初始化OSS并生成sts和oss客户端
 * @return {*}
 */
func Init() error {
	// 生成sts客户端
	config := &openapi.Config{
		AccessKeyId:     tea.String(zgconfig.Cfg.Oss.OssAccessKeyId),
		AccessKeySecret: tea.String(zgconfig.Cfg.Oss.OssAccessKeySecret),
	}

	config.Endpoint = tea.String(zgconfig.Cfg.Oss.StsEndpoint)
	_stsClient, _err := sts.NewClient(config)
	if _err != nil {
		logger.Logger.Error("error initializing OSS client:", _err)
		return _err
	}

	stsClient = _stsClient

	// 生成oss客户端
	var _ossClient *oss.Client
	_ossClient, _err = oss.New(zgconfig.Cfg.Oss.OssEndpoint, zgconfig.Cfg.Oss.AdminOssAccessKeyId, zgconfig.Cfg.Oss.AdminOssAccessKeySecret)
	if _err != nil {
		return _err
	}
	ossClient = _ossClient

	return nil
}

type OssPathReflect struct {
	RawPath string `json:"rawPath"`
	OssPath string `json:"ossPath"`
}

type GenrateResult struct {
	OssEndpoint     string           `json:"ossEndpoint"`
	BucketName      string           `json:"bucketName"`
	AccessKeyId     string           `json:"accessKeyId"`
	AccessKeySecret string           `json:"accessKeySecret"`
	Expiration      string           `json:"expiration"`
	SecurityToken   string           `json:"securityToken"`
	OssPathReflect  []OssPathReflect `json:"ossPathReflect"`
}

/**
 * @description: 生成sts的上传token
 * @param {[]string} originalPaths 原始文件路径数组
 * @return {*} 生成结果
 */
func GenerateSecurityToken(originalPaths []string) (*GenrateResult, error) {
	if len(originalPaths) == 0 {
		logger.Logger.Error("empty file names provided")
		return nil, errors.New("empty file names provided")
	}

	// 声明返回结果
	result := &GenrateResult{}
	result.OssEndpoint = zgconfig.Cfg.Oss.OssEndpoint
	result.BucketName = zgconfig.Cfg.Oss.BucketName

	// 待填充的新的文件路径集合
	resourcePaths := make([]string, 0, len(originalPaths))

	// 遍历原始文件名数组生成新的文件名
	timestamp := time.Now().UnixMilli()
	for _, originalPath := range originalPaths {
		// 只使用文件名
		originalFileName := filepath.Base(originalPath)

		// 提取文件名和扩展名
		fileName := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
		extension := filepath.Ext(originalFileName)

		// 生成oss上的存放路径
		pathOnOss := fmt.Sprintf("%s/%d/%s.%s",
			zgconfig.Cfg.Oss.DirFeedback,
			timestamp,
			fileName,
			extension,
		)
		result.OssPathReflect = append(result.OssPathReflect, OssPathReflect{RawPath: originalPath, OssPath: pathOnOss})

		// 生成resouce字段对应的路径
		pathInResource := fmt.Sprintf("acs:oss:*:*:%s/%s",
			zgconfig.Cfg.Oss.BucketName,
			pathOnOss,
		)
		resourcePaths = append(resourcePaths, pathInResource)
	}

	// 定义结构体来表示 Policy
	type Statement struct {
		Effect   string   `json:"Effect"`
		Action   string   `json:"Action"`
		Resource []string `json:"Resource"`
	}

	type Policy struct {
		Version   string      `json:"Version"`
		Statement []Statement `json:"Statement"`
	}

	// 创建 Policy 对象
	policy := Policy{
		Version: "1",
		Statement: []Statement{
			{
				Effect:   "Allow",
				Action:   "oss:PutObject",
				Resource: resourcePaths,
			},
		},
	}

	// 转化Policy为json字符串
	_, err := json.Marshal(policy)
	if err != nil {
		logger.Logger.Error("error marshalling policy:", err)
		return nil, err
	}

	// AssumeRole请求
	assumeRoleRequest := &sts.AssumeRoleRequest{
		RoleArn:         tea.String(zgconfig.Cfg.Oss.FeedbackRole),
		RoleSessionName: tea.String(zgconfig.Cfg.Oss.RoleSessionName),
		// Policy:          tea.String(string(policyBytes)),
	}

	// 解析请求结果
	stsResult := &sts.AssumeRoleResponse{}
	if stsResult, err = stsClient.AssumeRoleWithOptions(assumeRoleRequest, &util.RuntimeOptions{}); err != nil {
		logger.Logger.Error("error generating signed URL:", err)
		return nil, err
	}

	// 填充返回结果
	result.AccessKeyId = *stsResult.Body.Credentials.AccessKeyId
	result.AccessKeySecret = *stsResult.Body.Credentials.AccessKeySecret
	result.Expiration = *stsResult.Body.Credentials.Expiration
	result.SecurityToken = *stsResult.Body.Credentials.SecurityToken

	return result, nil
}

/**
 * @description:删除oss上的文件
 * @param {[]string} path oss路径
 * @return {*}
 */
func DeleteFileOnOssByPath(path []string) error {
	if len(path) == 0 {
		logger.Logger.Error("empty file names provided")
		return errors.New("empty file names provided")
	}

	bucket, err := ossClient.Bucket(zgconfig.Cfg.Oss.BucketName)
	if err != nil {
		return err
	}

	for _, ossPath := range path {
		err = bucket.DeleteObject(ossPath)
		if err != nil {
			logger.Logger.Error("error deleting file:", err)
			return err
		}
	}

	return nil
}
