package tools

import (
	"config"
	"express"
	"logger"
	"models"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// FilterLogging 全局记录日志过滤器
func FilterLogging(w *express.Response, r *express.Request, f *express.Filter) {
	now := time.Now()
	f.Handle(w, r)
	logger.Infof("%s %s %d %s\n", r.Request.Method, r.Request.URL.Path, w.StatusCode, time.Since(now))
}

// FilterAuthUser 全局cookie认证过滤器
func FilterAuthUser(w *express.Response, r *express.Request, f *express.Filter) {
	if strings.HasPrefix(r.Request.URL.Path, "/weixin") {
		f.Handle(w, r)
		return
	}
	cookie, err := r.Request.Cookie("token")
	if err != nil {
		w.Status(http.StatusUnauthorized)
		return
	}

	userInfo := new(models.User)
	if err := models.FindOne(config.CollUsers, bson.M{"Token": cookie.Value}, userInfo); err != nil {
		//if err := models.FindOne(config.CollUsers, bson.M{"Token": "test"}, userInfo); err != nil {
		if err == mgo.ErrNotFound {
			w.Status(http.StatusForbidden)
		} else {
			logger.Errorf("Mongodb error: %s\n", err.Error())
			w.Status(http.StatusInternalServerError)
		}
		return
	}
	w.SetLocals("UserInfo", userInfo)
	f.Handle(w, r)
}
