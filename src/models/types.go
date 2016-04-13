package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Url struct {
	UrlId      bson.ObjectId `bson:"_id" json:"UrlId"`             // ObjectId
	AliasName  string        `bson:"AliasName" json:"AliasName"`   // 别名
	Remark     string        `bson:"Remark" json:"Remark"`         // 备注
	Address    string        `bson:"Address" json:"Address"`       // 测试的url地址
	Interval   string        `bson:"Interval" json:"Interval"`     // 检测间隔单位秒
	UserId     string        `bson:"UserId" json:"-"`              // 所有者的id
	Disable    time.Time     `bson:"Disable" json:"Disable"`       // 禁用检测的过期时间
	UpdateTime time.Time     `bson:"UpdateTime" json:"UpdateTime"` // 检测状态更新时间
	IsOk       bool          `bson:"IsOk" json:"IsOk"`             // 当前的状态
	FailedTime time.Time     `bson:"FailedTime" json:"-"`          // 检测失败开始时间
}

type User struct {
	UserId    bson.ObjectId `bson:"_id" json:"-"`             // ObjectId
	WxID      string        `bson:"WxID" json:"WxID"`         // 用户的微信ID
	Members   []string      `bson:"Members" json:"-"`         // 共享的用户
	NickName  string        `bson:"NickName" json:"NickName"` // 昵称
	Remark    string        `bson:"Remark" json:"Remark"`     // 备注
	Token     string        `bson:"Token" json:"-"`           // 用户认证token
	Groups    []string      `bson:"Groups" json:"-"`          // 加入的组
	QRCodeUrl string        `bson:"QRCodeUrl" json:"-"`       // 邀请加入的二维码地址
}

type Queue struct {
	WxID   string    `bson:"WxID" json:"WxID"`     // 微信ID
	UserID string    `bson:"UserID" json:"UserID"` // 用户ID
	Date   time.Time `bson:"Date" json:"-"`        // 插入时间
}
