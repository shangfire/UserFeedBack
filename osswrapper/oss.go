/*
 * @Author: shanghanjin
 * @Date: 2024-08-13 16:46:39
 * @LastEditTime: 2024-08-16 14:08:44
 * @FilePath: \UserFeedBack\osswrapper\oss.go
 * @Description:
 */
package osswrapper

import (
	logger "UserFeedBack/logwrapper"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var client *sts.Client

func Init() error {
	config := &openapi.Config{
		AccessKeyId:     tea.String(""),
		AccessKeySecret: tea.String(""),
	}

	config.Endpoint = tea.String("oss-cn-beijing.aliyuncs.com")
	_client, _err := sts.NewClient(config)
	if _err != nil {
		logger.Logger.Error("Error initializing OSS client:", _err)
		return _err
	}

	client = _client
	return nil
}

func GenerateUploadUrl(originalFileName string) (string, error) {
	// 获取当前时间
	currentTime := time.Now()

	// 提取文件名和扩展名
	fileName := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
	extension := filepath.Ext(originalFileName)

	// 生成时间戳
	timestamp := currentTime.Unix()

	// 生成路径，日期格式为 "年-月-日"
	path := fmt.Sprintf("%d-%02d-%02d/%s_%d%s",
		currentTime.Year(),
		currentTime.Month(),
		currentTime.Day(),
		fileName,
		timestamp,
		extension,
	)

	assumeRoleRequest := &sts.AssumeRoleRequest{
		RoleArn: tea.String("acs:ram::1652022578600526:role/rolests"),
	}

	// 生成临时上传URL
	signedURL, err := bucket.SignURL(path, oss.HTTPPut, 3600) // 3600秒=1小时
	if err != nil {
		logger.Logger.Error("Error generating signed URL:", err)
		return "", err
	}

	// 打印生成的URL
	logger.Logger.Debug("Signed URL:", signedURL)

	return signedURL, nil
}
