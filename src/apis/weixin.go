package apis

import (
	"config"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"express"
	"fmt"
	"github.com/docker/distribution/uuid"
	"gopkg.in/mgo.v2/bson"
	"logger"
	"models"
	"net/http"
	"sort"
	"strings"
	"time"
)

type wxEvent struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Event        string `xml:"Event"`
	EventKey     string `xml:"EventKey"`
}

type wxOAuth struct {
	OpenID   string `json:"openid"`
	ErrorMsg string `json:"errmsg"`
}

// WeiXinGet 微信接入服务器的验证  GET /weixin
func WeiXinGet(w *express.Response, r *express.Request) {
	signature := r.Request.FormValue("signature")
	timestamp := r.Request.FormValue("timestamp")
	nonce := r.Request.FormValue("nonce")
	echostr := r.Request.FormValue("echostr")

	temp := []string{config.WxServerToken, timestamp, nonce}
	sort.Strings(temp)
	result := fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(temp, ""))))
	logger.Debug(result)
	logger.Debug(signature)
	if result == signature {
		w.Send(echostr)
		return
	}
	w.Status(http.StatusForbidden)
}

// WeiXinPost 微信推送信息 POST /weixin
func WeiXinPost(w *express.Response, r *express.Request) {
	event := new(wxEvent)
	xml.NewDecoder(r.Request.Body).Decode(event)
	logger.Debugf("%#v", event)
	switch event.Event {
	case "subscribe":
		if _, err := models.Upsert(config.CollUsers, bson.M{"WxID": event.FromUserName}, models.User{WxID: event.FromUserName, UserId: bson.NewObjectId()}); err != nil {
			logger.Errorf("Registry user error: %s\n", err.Error())
		}
		fallthrough
	case "SCAN":
		if event.EventKey != "" {
			if err := models.Insert(config.CollTTLQueue, models.Queue{WxID: event.FromUserName, UserID: strings.TrimLeft(event.EventKey, "qrscene_"), Date: time.Now()}); err != nil {
				logger.Errorf("Insert to TTLQueue error: %s\n", err.Error())
			}
		}
	case "unsubscribe":
		selfInfo := new(models.User)
		if err := models.FindOne(config.CollUsers, bson.M{"WxID": event.FromUserName}, selfInfo); err != nil {
			logger.Errorf("Find user error: %s\n", err.Error())
			return
		}
		models.Remove(config.CollUrls, bson.M{"UserId": selfInfo.UserId.Hex()})
	}
}

// WeiXinOAuth 微信OAuth认证 GET /weixin/oauth
func WeiXinOAuth(w *express.Response, r *express.Request) {
	code := r.Request.FormValue("code")
	if code == "" {
		logger.Error("Weixin OAuth error: Response code is null")
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Response code is null"))
		return
	}
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", config.WxAppID, config.WxAppSecret, code)
	res, err := http.Get(apiURL)
	if err != nil {
		logger.Errorf("Weixin OAuth error: %s\n", err.Error())
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	body := new(wxOAuth)
	if err := json.NewDecoder(res.Body).Decode(body); err != nil {
		logger.Errorf("Weixin OAuth error: %s\n", err.Error())
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	if body.ErrorMsg != "" {
		logger.Errorf("Weixin OAuth error: %s\n", body.ErrorMsg)
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	token := uuid.Generate().String()
	models.Update(config.CollUsers, bson.M{"WxID": body.OpenID}, bson.M{"$set": bson.M{"Token": token}})
	cookie := http.Cookie{Name: "token", Value: token, Path: "/"}
	http.SetCookie(w, &cookie)
	http.Redirect(w.ResponseWriter, r.Request, "/#/users/info", http.StatusFound)
}