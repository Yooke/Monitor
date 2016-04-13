package apis

import (
	"config"
	"encoding/json"
	"express"
	"logger"
	"models"
	"net/http"
	"tools"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// UrlsDir 获取监控列表 GET /urls
func UrlsDir(w *express.Response, r *express.Request) {
	selfInfo := w.GetLocals("UserInfo").(*models.User)
	body := make(map[string][]models.Url)
	selfUrls := []models.Url{}
	if err := models.FindAll(config.CollUrls, bson.M{"UserId": selfInfo.UserId.Hex()}, &selfUrls); err != nil {
		logger.Errorf("UrlsDir mongodb error: %s\n", err.Error())
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	body["self"] = selfUrls
	if len(selfInfo.Groups) > 0 {
		groupsInfo := []models.User{}
		if err := models.FindAll(config.CollUsers, bson.M{"WxID": bson.M{"$in": selfInfo.Groups}}, &groupsInfo); err != nil {
			logger.Errorf("UrlsDir mongodb error: %s\n", err.Error())
			w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
			return
		}
		for _, group := range groupsInfo {
			data := []models.Url{}
			if err := models.FindAll(config.CollUrls, bson.M{"UserId": group.UserId.Hex()}, &data); err != nil {
				logger.Errorf("UrlsDir mongodb error: %s\n", err.Error())
				w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
				return
			}
			body[group.NickName] = data
		}
	}
	w.Json(express.MustToJson(body))
}

// UrlsInfo 获取监控的详细信息  GET /urls/{UrlId}
func UrlsInfo(w *express.Response, r *express.Request) {
	userInfo := w.GetLocals("UserInfo").(*models.User)
	urlID := r.PathParam["UrlId"]
	url := new(models.Url)
	if bson.IsObjectIdHex(urlID) == false {
		w.Status(http.StatusBadRequest).Json(NewJSONErrorf("Invalid url id for %s", urlID))
		return
	}
	if err := models.FindOne(config.CollUrls, bson.M{"_id": bson.ObjectIdHex(urlID)}, url); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONErrorf("Mongodb error: %s", err.Error()))
		return
	}
	if url.UserId != userInfo.UserId.Hex() {
		w.Status(http.StatusForbidden).Json(NewJSONError("403 Forbidden"))
		return
	}
	w.Json(express.MustToJson(url))
}

// UrlsAdd 添加url，需要参数：AliasName Address Interval   POST /urls
func UrlsAdd(w *express.Response, r *express.Request) {
	urlInfo := models.Url{}
	defer r.Request.Body.Close()
	if err := json.NewDecoder(r.Request.Body).Decode(&urlInfo); err != nil {
		w.Status(http.StatusBadRequest).Json(NewJSONErrorf("Decode error: %s", err.Error()))
		return
	}
	if tools.ValidURL(urlInfo.Address) == false || urlInfo.AliasName == "" {
		w.Status(http.StatusBadRequest).Json(NewJSONError("400 Bad Request"))
		return
	}
	if urlInfo.Interval != "60" && urlInfo.Interval != "300" && urlInfo.Interval != "600" {
		w.Status(http.StatusBadRequest).Json(NewJSONError("400 Bad Resquest"))
		return
	}
	urlInfo.UrlId = bson.NewObjectId()
	urlInfo.UserId = w.GetLocals("UserInfo").(*models.User).UserId.Hex()
	urlInfo.IsOk = true
	logger.Debugf("%#v\n", urlInfo)
	if err := models.Insert(config.CollUrls, urlInfo); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONErrorf("Mongodb error: %s", err.Error()))
		return
	}
	w.Status(http.StatusCreated)
}

// UrlsUpdate 更新url， 可选参数：AliasName Address Remark Interval   POST /urls/{UrlId}
func UrlsUpdate(w *express.Response, r *express.Request) {
	urlID := r.PathParam["UrlId"]
	urlInfo := models.Url{}
	if bson.IsObjectIdHex(urlID) == false {
		w.Status(http.StatusBadRequest).Json(NewJSONError("Invalid url id"))
		return
	}
	defer r.Request.Body.Close()
	if err := json.NewDecoder(r.Request.Body).Decode(&urlInfo); err != nil {
		w.Status(http.StatusBadRequest).Json(NewJSONErrorf("Decode error: %s", err.Error()))
		return
	}
	selector := bson.M{"_id": bson.ObjectIdHex(urlID), "UserId": w.GetLocals("UserInfo").(*models.User).UserId.Hex()}
	doc := bson.M{}
	if urlInfo.AliasName != "" {
		doc["AliasName"] = urlInfo.AliasName
	}
	if urlInfo.Remark != "" {
		doc["Remark"] = urlInfo.Remark
	}
	if tools.ValidURL(urlInfo.Address) {
		doc["Address"] = urlInfo.Address
	}
	if urlInfo.Interval == "60" || urlInfo.Interval == "300" || urlInfo.Interval == "600" {
		doc["Interval"] = urlInfo.Interval
	}
	if len(doc) == 0 {
		w.Status(http.StatusBadRequest).Json(NewJSONError("Bad Request"))
		return
	}
	if err := models.Update(config.CollUrls, selector, bson.M{"$set": doc}); err != nil {
		if err == mgo.ErrNotFound {
			w.Status(http.StatusNotFound).Json(NewJSONError("Resource not found or forbidden"))
		} else {
			w.Status(http.StatusInternalServerError).Json(NewJSONErrorf("Mongodb error: %s", err.Error()))
		}
		return
	}
	w.Status(http.StatusCreated)
}

// UrlsDelete 删除url   DELETE /urls/{UrlId}
func UrlsDelete(w *express.Response, r *express.Request) {
	urlID := r.PathParam["UrlId"]
	userInfo := w.GetLocals("UserInfo").(*models.User)
	if bson.IsObjectIdHex(urlID) == false {
		w.Status(http.StatusBadRequest).Json(NewJSONError("Invalid url id"))
		return
	}
	if err := models.Remove(config.CollUrls, bson.M{"_id": bson.ObjectIdHex(urlID), "UserId": userInfo.UserId.Hex()}); err != nil {
		if err == mgo.ErrNotFound {
			w.Status(http.StatusNotFound).Json(NewJSONError("Resource not found or forbidden"))
		} else {
			w.Status(http.StatusInternalServerError).Json(NewJSONErrorf("Mongodb error: %s", err.Error()))
		}
		return
	}
	w.Status(http.StatusNoContent)
}
