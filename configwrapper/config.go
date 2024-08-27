/*
 * @Author: shanghanjin
 * @Date: 2024-08-19 17:51:57
 * @LastEditTime: 2024-08-19 19:03:25
 * @FilePath: \UserFeedBack\configwrapper\config.go
 * @Description:
 */
package configwrapper

import (
	logger "UserFeedBack/logwrapper"
	"encoding/json"
	"os"
)

type Oss struct {
	OssAccessKeyId          string `json:"accessKeyId"`
	OssAccessKeySecret      string `json:"accessKeySecret"`
	OssEndpoint             string `json:"ossEndpoint"`
	StsEndpoint             string `json:"stsEndpoint"`
	FeedbackRole            string `json:"feedbackRole"`
	BucketName              string `json:"bucketName"`
	DirFeedback             string `json:"dirFeedback"`
	RoleSessionName         string `json:"roleSessionName"`
	AdminOssAccessKeyId     string `json:"adminAccessKeyId"`
	AdminOssAccessKeySecret string `json:"adminAccessKeySecret"`
}

type Database struct {
	User     string `json:"user"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Schema   string `json:"schema"`
	Password string `json:"password"`
}

type Config struct {
	Oss      Oss      `json:"oss"`
	Database Database `json:"database"`
}

var Cfg *Config

func Init(configFilePath string) error {
	Cfg = &Config{}

	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		logger.Logger.Fatalf("Error reading config file: %v", err)
	}

	err = json.Unmarshal(configData, Cfg)
	if err != nil {
		logger.Logger.Fatalf("Error parsing config file: %v", err)
	}

	return nil
}
