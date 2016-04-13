package loop

import (
	"config"
	"gopkg.in/mgo.v2/bson"
	"logger"
	"models"
	"net/http"
	"time"
	"tools"
)

type Task struct {
	Url     models.Url
	IsOk    bool
	Counter int
}

var (
	counter     int
	taskChannel chan *Task = make(chan *Task, config.GoNumber*2)
)

func Loop() {
	for i := 0; i < config.GoNumber; i++ {
		go taskExec(taskChannel)
	}
	tick := time.Tick(60 * time.Second)
	for {
		<-tick
		counter++
		logger.Debug("In main loop")
		go fetchUrls()
	}
}

func fetchUrls() {
	selector := bson.M{"Disable": bson.M{"$lte": time.Now()}}
	switch counter {
	case 5:
		selector["$or"] = []bson.M{bson.M{"Interval": "60"}, bson.M{"Interval": "300"}}
	case 10:
		selector["$or"] = []bson.M{bson.M{"Interval": "60"}, bson.M{"Interval": "300"}, bson.M{"Interval": "600"}}
		counter = 0
	default:
		selector["Interval"] = "60"
	}
	// 查询本次要操作的对象
	result := new([]models.Url)
	if err := models.FindAll(config.CollUrls, selector, result); err != nil {
		logger.Errorf("Mongodb error: %s\n", err.Error())
		return
	}
	// 更新本次操作对象的UpdateTime字段
	if len(*result) > 0 {
		_, err := models.UpdateAll(config.CollUrls, selector, bson.M{"$set": bson.M{"UpdateTime": time.Now()}})
		if err != nil {
			logger.Warnf("Mongodb error: %s\n", err.Error())
		}
	}
	// 添加任务到队列
	for _, url := range *result {
		logger.Debugf("%-15s Put %s in taskChannel\n", "taskSchedule", url.Address)
		taskChannel <- &Task{url, url.IsOk, 0}
	}
}

// 检测失败后的处理
func (task *Task) handleFail() {
	// 如果已经处于故障状态，则不做任何操作
	if task.IsOk == false {
		return
	}
	task.Counter++
	switch {
	// 检测失败，后提升检测间隔至10秒钟
	case task.Counter < 3:
		go func() {
			<-time.After(10 * time.Second)
			taskChannel <- task
		}()
	// 失败3次后，通知用户，并更新mongodb的 IsOk: false, FailedTime: time.Now()
	case task.Counter == 3:
		logger.Info("Check failed, send bad message, update mongodb!!!")
		models.Update(config.CollUrls, bson.M{"_id": task.Url.UrlId}, bson.M{"$set": bson.M{"IsOk": false, "FailedTime": time.Now()}})
		tools.CheckFalseMessage(task.Url)
	}
}

// 检测成功后的处理
func (task *Task) handleSuccess() {
	// 如果当前状态正常，则不做任何处理
	if task.IsOk {
		return
	}
	// 如果当前状态IsOk: false, 则通用户故障恢复，并更新mongodb的 IsOk: true
	logger.Info("Check success, send ok message, update mongodb IsOk: true")
	models.Update(config.CollUrls, bson.M{"_id": task.Url.UrlId}, bson.M{"$set": bson.M{"IsOk": true}})
	tools.CheckTrueMessage(task.Url)
}

func taskExec(channel <-chan *Task) {
	client := &http.Client{Timeout: 5 * time.Second}
	for task := range channel {
		res, err := client.Head(task.Url.Address)
		if err != nil {
			logger.Debugf("%-15s Exec %s error: %s\n", "taskExec", task.Url.Address, err.Error())
			task.handleFail()
		} else {
			logger.Debugf("%-15s Exec %s success: %d\n", "taskExec", task.Url.Address, res.StatusCode)
			task.handleSuccess()
		}
	}
}
