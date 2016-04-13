package apis

import (
	"config"
	"express"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"logger"
	"models"
)

func QueueJoin(w *express.Response, r *express.Request) {
	selfInfo := w.GetLocals("UserInfo").(*models.User)
	body := models.Queue{}
	if err := models.FindOne(config.CollTTLQueue, bson.M{"UserID": selfInfo.UserId.Hex()}, &body); err != nil {
		if err != mgo.ErrNotFound {
			logger.Errorf("Get TTLQueue error: %s\n", err.Error())
		}
		w.Send(NewJSONMessage("null"))
		return
	}
	w.Send(NewJSONMessage(body.WxID))
}
