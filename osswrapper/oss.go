/*
 * @Author: shanghanjin
 * @Date: 2024-08-13 16:46:39
 * @LastEditTime: 2024-08-20 11:50:15
 * @FilePath: \UserFeedBack\osswrapper\oss.go
 * @Description:
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
)

var client *sts.Client

func Init() error {
	config := &openapi.Config{
		AccessKeyId:     tea.String(zgconfig.Cfg.Oss.OssAccessKeyId),
		AccessKeySecret: tea.String(zgconfig.Cfg.Oss.OssAccessKeySecret),
	}

	config.Endpoint = tea.String(zgconfig.Cfg.Oss.StsEndpoint)
	_client, _err := sts.NewClient(config)
	if _err != nil {
		logger.Logger.Error("error initializing OSS client:", _err)
		return _err
	}

	client = _client
	return nil
}

type GenrateResult struct {
	AccessKeyId     string
	AccessKeySecret string
	Expiration      string
	SecurityToken   string
	OssPaths        []string
}

func GenerateSecurityToken(originalFileNames []string) (*GenrateResult, error) {
	if len(originalFileNames) == 0 {
		logger.Logger.Error("empty file names provided")
		return nil, errors.New("empty file names provided")
	}

	// 返回结果
	result := &GenrateResult{}

	// 待填充的新的文件路径集合
	resourcePaths := make([]string, 0, len(originalFileNames))

	// 遍历原始文件名数组生成新的文件名
	for _, originalFileName := range originalFileNames {
		// 提取文件名和扩展名
		fileName := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
		extension := filepath.Ext(originalFileName)

		// 生成oss上的存放路径
		pathOnOss := fmt.Sprintf("%s/%s_%d%s",
			zgconfig.Cfg.Oss.DirFeedback,
			fileName,
			time.Now().Unix(),
			extension,
		)
		result.OssPaths = append(result.OssPaths, pathOnOss)

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
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		logger.Logger.Error("error marshalling policy:", err)
		return nil, err
	}

	assumeRoleRequest := &sts.AssumeRoleRequest{
		RoleArn:         tea.String(zgconfig.Cfg.Oss.FeedbackRole),
		RoleSessionName: tea.String(zgconfig.Cfg.Oss.RoleSessionName),
		Policy:          tea.String(string(policyBytes)),
	}

	stsResult := &sts.AssumeRoleResponse{}
	if stsResult, err = client.AssumeRoleWithOptions(assumeRoleRequest, &util.RuntimeOptions{}); err != nil {
		logger.Logger.Error("error generating signed URL:", err)
		return nil, err
	}

	result.AccessKeyId = *stsResult.Body.Credentials.AccessKeyId
	result.AccessKeySecret = *stsResult.Body.Credentials.AccessKeySecret
	result.Expiration = *stsResult.Body.Credentials.Expiration
	result.SecurityToken = *stsResult.Body.Credentials.SecurityToken

	return result, nil
}
