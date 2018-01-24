package handlers

import (
	"encoding/json"
	"fmt"

	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/model"
)

func (Ctx *AppContext) TopicList(frame Frame) {
	var topics []model.Topic
	//Ctx.Db.Joins("LEFT JOIN categories ON (topics.category_id = categories.id)").Select("*").Find(&topics)
	Ctx.Db.Preload("Category").Find(&topics)
	ts, err := json.Marshal(topics)
	if err == nil {
		Ctx.Ws.WriteJSON(Frame{Type: TopicListFrame, Data: string(ts)})
	} else {
		Ctx.Logger.WithError(err).Error()
	}
}

func (Ctx *AppContext) TopicSave(frame Frame) {
	if Ctx.IsAuthorized() {
		var topic model.Topic
		err := frame.BindJSON(&topic)
		Ctx.Logger.Info("Decoded topic", topic)
		Ctx.Logger.Info("User", Ctx.User)
		if err == nil {
			if topic.ID > 0 {
				// find topic in DB and update it
				var oldTopic model.Topic
				Ctx.Db.Where("id = ?", topic.ID).First(&oldTopic)
				if oldTopic.ID > 0 {
					oldTopic.Title = topic.Title
					//oldTopic.Acl = topic.Acl
					oldTopic.Body = topic.Body
					err = Ctx.Db.Set("gorm:save_associations", false).Save(&oldTopic).Error
				}
			} else {
				// this is new topic
				topic.ID = 0
				fmt.Println(frame.Data)
				err = Ctx.Db.Omit("acl.email").Save(&topic).Error
			}
			if err != nil {
				Ctx.Logger.WithError(err).Error("Unable to save topic")
			}
		} else {
			Ctx.Logger.WithError(err).Error("can't unmarshall topic")
		}
	} else {
		Ctx.Logger.WithField("msg", "Unauthorized topic save call").Info()
	}
}
