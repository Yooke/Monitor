package models

import (
	"config"
	"gopkg.in/mgo.v2"
	"logger"
	"time"
)

var sessionMongo *mgo.Session

func init() {
	session, err := getSession()
	if err != nil {
		logger.Fatalf("Get mongodb session error: %s\n", err.Error())
	}
	err = session.DB("").C(config.CollTTLQueue).EnsureIndex(mgo.Index{Key: []string{"Date"}, ExpireAfter: 10 * time.Second})
	if err != nil {
		logger.Fatalf("Create mongodb index error: %s\n", err.Error())
	}
}

func getSession() (*mgo.Session, error) {
	if sessionMongo == nil {
		session, err := mgo.DialWithTimeout(config.MongoUrl, 3*time.Second)
		if err != nil {
			return nil, err
		}
		session.SetPoolLimit(30)
		return session.Clone(), nil
	}
	return sessionMongo.Clone(), nil
}

func FindAll(collection string, selector, result interface{}) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.DB("").C(collection).Find(selector).All(result)
}

func FindOne(collection string, selector, result interface{}) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.DB("").C(collection).Find(selector).One(result)
}

func Insert(collection string, doc interface{}) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.DB("").C(collection).Insert(doc)
}

func Update(collection string, selector, doc interface{}) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.DB("").C(collection).Update(selector, doc)
}

func Upsert(collection string, selector, doc interface{}) (*mgo.ChangeInfo, error) {
	session, err := getSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.DB("").C(collection).Upsert(selector, doc)
}

func UpdateAll(collection string, selector, doc interface{}) (*mgo.ChangeInfo, error) {
	session, err := getSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.DB("").C(collection).UpdateAll(selector, doc)
}

func Remove(collection string, selector interface{}) error {
	session, err := getSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.DB("").C(collection).Remove(selector)
}

func RemoveAll(collection string, selector interface{}) (*mgo.ChangeInfo, error) {
	session, err := getSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.DB("").C(collection).RemoveAll(selector)
}