package tools

import (
	"bytes"
	"config"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"logger"
	"models"
	"net/http"
	"strings"
	"time"
)

var accessToken struct {
	Token     string
	ExpiresAt time.Time
}

type accessTokenReturn struct {
	AccessToken string `json:"access_token"`
	Expires_In  int    `json:"expires_in"`
	ErrorMsg    string `json:"errmsg"`
}

type createQRCodeReturn struct {
	Url      string `json:"url"`
	ErrorMsg string `json:"errmsg"`
}

type templateMessage struct {
	ToUser     string                       `json:"touser"`
	TemplateID string                       `json:"template_id"`
	Url        string                       `json:"url"`
	Data       map[string]map[string]string `json:"data"`
}

func GetAccessToken() (string, error) {
	if time.Now().Before(accessToken.ExpiresAt) {
		return accessToken.Token, nil
	}
	apiUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", config.WxAppID, config.WxAppSecret)
	res, err := http.Get(apiUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body := new(accessTokenReturn)
	json.NewDecoder(res.Body).Decode(body)
	if body.ErrorMsg != "" {
		return "", errors.New(body.ErrorMsg)
	}
	accessToken.Token = body.AccessToken
	accessToken.ExpiresAt = time.Now().Add(time.Duration(body.Expires_In-500) * time.Second)
	return accessToken.Token, nil
}

func GetQRUrl(id string) (string, error) {
	token, err := GetAccessToken()
	if err != nil {
		return "", err
	}
	apiUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s", token)
	res, err := http.Post(apiUrl, "application/json", strings.NewReader(fmt.Sprintf(`{"action_name": "QR_LIMIT_STR_SCENE", "action_info": {"scene": {"scene_str": "%s"}}}`, id)))
	if err != nil {
		return "", err
	}
	body := new(createQRCodeReturn)
	json.NewDecoder(res.Body).Decode(body)
	if body.ErrorMsg != "" {
		return "", errors.New(body.ErrorMsg)
	}
	return body.Url, nil
}

func CheckFalseMessage(urlInfo models.Url) {
	userInfo := new(models.User)
	if err := models.FindOne(config.CollUsers, bson.M{"_id": bson.ObjectIdHex(urlInfo.UserId)}, userInfo); err != nil {
		return
	}
	userInfo.Members = append(userInfo.Members, userInfo.WxID)
	for _, wxID := range userInfo.Members {
		body := templateMessage{
			ToUser:     wxID,
			TemplateID: config.CheckFalseTemplate,
			//Url:        fmt.Sprintf("http://monitor.qbangmang.com/urls/" + urlInfo.UrlId.Hex()),
			Data:       make(map[string]map[string]string),
		}
		body.Data["first"] = map[string]string{"value": "警告！！！连续3次检测失败。\n", "color": "#ff0000"}
		body.Data["title"] = map[string]string{"value": urlInfo.AliasName, "color":"#173177"}
		body.Data["info"] = map[string]string{"value": urlInfo.Remark, "color":"#173177"}
		body.Data["remark"] = map[string]string{"value": "\n请及时处理，故障恢复后会通知你。", "color":"#696969"}
		go body.Send()
	}

}

func CheckTrueMessage(urlInfo models.Url) {
	userInfo := new(models.User)
	if err := models.FindOne(config.CollUsers, bson.M{"_id": bson.ObjectIdHex(urlInfo.UserId)}, userInfo); err != nil {
		return
	}
	userInfo.Members = append(userInfo.Members, userInfo.WxID)
	for _, wxID := range userInfo.Members {
		body := templateMessage{
			ToUser:     wxID,
			TemplateID: config.CheckTrueTemplate,
			//Url:        fmt.Sprintf("http://monitor.qbangmang.com/urls/" + urlInfo.UrlId.Hex()),
			Data:       make(map[string]map[string]string),
		}
		body.Data["first"] = map[string]string{"value": "故障已恢复正常。\n", "color": "#32cd32"}
		body.Data["title"] = map[string]string{"value": urlInfo.AliasName, "color":"#173177"}
		body.Data["info"] = map[string]string{"value": urlInfo.Remark, "color":"#173177"}
		body.Data["remark"] = map[string]string{"value": fmt.Sprintf("\n故障持续时间，%.0f 分钟！", time.Now().Sub(urlInfo.FailedTime).Minutes()), "color":"#696969"}
		go body.Send()
	}
}

func (msg templateMessage) Send() {
	token, err := GetAccessToken()
	if err != nil {
		logger.Errorf("Get access token error: %s\n", err.Error())
		return
	}
	data, _ := json.Marshal(msg)
	apiUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", token)
	_, err = http.Post(apiUrl, "application/json", bytes.NewReader(data))
	if err != nil {
		logger.Errorf("Request template message error: %s\n", err.Error())
		return
	}
}
