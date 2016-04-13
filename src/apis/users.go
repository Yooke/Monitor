package apis

import (
	"config"
	"encoding/json"
	"express"
	"models"
	"net/http"

	"gopkg.in/mgo.v2/bson"
	"io"
	"logger"
	"tools"
)

// UsersInfo 用户基本信息 GET /users/self
func UsersInfo(w *express.Response, r *express.Request) {
	selfInfo := w.GetLocals("UserInfo").(*models.User)
	w.Json(express.MustToJson(*selfInfo))
}

// UsersQRCode 用户邀请的二维码 GET /users/qrcode
func UsersQRCode(w *express.Response, r *express.Request) {
	selfInfo := w.GetLocals("UserInfo").(*models.User)
	if selfInfo.QRCodeUrl == "" {
		url, err := tools.GetQRUrl(selfInfo.UserId.Hex())
		if err != nil {
			logger.Errorf("GetQRUrl error: %s\n", err.Error())
			w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
			return
		}
		selfInfo.QRCodeUrl = url
		models.Update(config.CollUsers, bson.M{"_id": selfInfo.UserId}, bson.M{"$set": bson.M{"QRCodeUrl": url}})
	}
	code, err := tools.CreateQRCode(selfInfo.QRCodeUrl)
	if err != nil {
		logger.Errorf("CreateQRCode error: %s\n", err.Error())
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	w.Header().Set("Content-Type", "image/png")
	io.Copy(w.ResponseWriter, code)
}

// UsersUpdate 更新用户信息 POST /users/self
func UsersUpdate(w *express.Response, r *express.Request) {
	selfInfo := models.User{}
	if err := json.NewDecoder(r.Request.Body).Decode(&selfInfo); err != nil {
		w.Status(http.StatusBadRequest).Json(NewJSONErrorf("Decode error: %s", err.Error()))
		return
	}
	selector := bson.M{"_id": w.GetLocals("UserInfo").(*models.User).UserId}
	doc := bson.M{}
	if selfInfo.NickName != "" {
		doc["NickName"] = selfInfo.NickName
	}
	if selfInfo.Remark != "" {
		doc["Remark"] = selfInfo.Remark
	}
	if err := models.Update(config.CollUsers, selector, bson.M{"$set": doc}); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONErrorf("Mongodb error: %s", err.Error()))
		return
	}
	w.Status(http.StatusCreated)
}

// UsersMemberAdd 所有者添加成员 PUT /users/member
func UsersMemberAdd(w *express.Response, r *express.Request) {
	selfInfo := w.GetLocals("UserInfo").(*models.User)
	wxID := r.Request.FormValue("WxID")
	// 加入自己的Members
	if err := models.Update(config.CollUsers, bson.M{"_id": selfInfo.UserId}, bson.M{"$addToSet": bson.M{"Members": wxID}}); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	// 更新成员的Groups
	if err := models.Update(config.CollUsers, bson.M{"WxID": wxID}, bson.M{"$addToSet": bson.M{"Groups": selfInfo.WxID}}); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	// 移除TTLQueue
	models.Remove(config.CollTTLQueue, bson.M{"UserID": selfInfo.UserId.Hex()})
	w.Status(http.StatusCreated)
}

// UsersMemberList 成员列表  GET /users/member
func UsersMemerList(w *express.Response, r *express.Request) {
	selfInfo := w.GetLocals("UserInfo").(*models.User)
	selector := bson.M{"WxID": bson.M{"$in": selfInfo.Members}}
	memList := []models.User{}
	if err := models.FindAll(config.CollUsers, selector, &memList); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	w.Json(express.MustToJson(memList))
}

// UsersMemberDel 所有者移除成员 DELETE /users/member
func UsersMemberDel(w *express.Response, r *express.Request) {
	selfInfo := w.GetLocals("UserInfo").(*models.User)
	wxID := r.Request.FormValue("WxID")
	// 从自己的Members中移除成员
	if err := models.Update(config.CollUsers, bson.M{"_id": selfInfo.UserId}, bson.M{"$pull": bson.M{"Members": wxID}}); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Interval server error"))
		return
	}
	// 从成员的的Groups中移除自己
	if err := models.Update(config.CollUsers, bson.M{"WxID": wxID}, bson.M{"$pull": bson.M{"Groups": selfInfo.WxID}}); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	w.Status(http.StatusNoContent)
}

// UsersGroupList 加入组列表 GET /users/group
func UsersGroupList(w *express.Response, r *express.Request) {
	selfInfo := w.GetLocals("UserInfo").(*models.User)
	selector := bson.M{"WxID": bson.M{"$in": selfInfo.Groups}}
	grpList := []models.User{}
	if err := models.FindAll(config.CollUsers, selector, &grpList); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	w.Json(express.MustToJson(grpList))
}

// UsersGroupDel 成员自己退出组 DELETE /users/group
func UsersGroupDel(w *express.Response, r *express.Request) {
	userInfo := w.GetLocals("UserInfo").(*models.User)
	wxID := r.Request.FormValue("WxID")
	// 从所有者的Members中移除自己
	if err := models.Update(config.CollUsers, bson.M{"WxID": wxID}, bson.M{"$pull": bson.M{"Members": userInfo.WxID}}); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	// 从自己的Groups中移除该组
	if err := models.Update(config.CollUsers, bson.M{"_id": userInfo.UserId}, bson.M{"$pull": bson.M{"Groups": wxID}}); err != nil {
		w.Status(http.StatusInternalServerError).Json(NewJSONError("Internal server error"))
		return
	}
	w.Status(http.StatusNoContent)
}
