package config

import (
	"encoding/json"
	"io/ioutil"
	"logger"
	"os"
	"strconv"
)

const (
	CollUsers    string = "Users"
	CollUrls     string = "Urls"
	CollTTLQueue string = "TTLQueue"
)

var (
	Listener           string // 监听地址和端口
	WxAppID            string // 微信AppID
	WxAppSecret        string // 微信AppSecret
	MongoUrl           string // Mongodb 连接地址
	GoNumber           int    // 启动Go routine个数
	WxServerToken      string // 微信接入服务器配置Token
	CheckFalseTemplate string // 故障报警 模板ID
	CheckTrueTemplate  string // 故障恢复 模板ID
)

func init() {
	LoadConfig()
}

// 解析配置文件
func LoadConfig() {
	if len(os.Args) < 2 {
		logger.Fatal("Required the config file.")
	}
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		logger.Fatal(err)
	}
	result := make(map[string]string)
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Fatal(err)
	}
	Listener = verifyKey(result, "Listener")
	MongoUrl = verifyKey(result, "MongoUrl")
	WxAppID = verifyKey(result, "WxAppID")
	WxAppSecret = verifyKey(result, "WxAppSecret")
	WxServerToken = verifyKey(result, "WxServerToken")
	CheckFalseTemplate = verifyKey(result, "CheckFalseTemplate")
	CheckTrueTemplate = verifyKey(result, "CheckTrueTemplate")
	GoNumber, err = strconv.Atoi(verifyKey(result, "GoNumber"))
	if err != nil {
		logger.Fatalf("The config for GoNumber must can be convert to int")
	}
}

// 验证key是否存在
func verifyKey(keyMap map[string]string, key string) string {
	if value, ok := keyMap[key]; ok {
		return value
	}
	logger.Fatalf("Required the config key < %s >.", key)
	return ""
}
