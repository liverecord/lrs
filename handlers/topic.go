package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/model"
)

func (Ctx *AppContext) Topic(frame Frame) {
	var topic model.Topic

	var data map[string]string
	frame.BindJSON(&data)

	if slug, ok := data["slug"]; ok {
		Ctx.Db.
			Preload("Category").
			Preload("User").
			Where("slug = ?", slug).First(&topic)
		if topic.ID > 0 {
			ts, err := json.Marshal(topic)
			if err == nil {
				topicFrame := Frame{Type: TopicFrame, Data: string(ts)}
				Ctx.Ws.WriteJSON(topicFrame)
				Ctx.CommentList(topicFrame)
				Ctx.Db.Model(&topic).UpdateColumn("total_views", gorm.Expr("total_views + ?", 1))
			} else {
				Ctx.Logger.WithError(err).Error()
			}
		}
	}
}

func (Ctx *AppContext) TopicList(frame Frame) {
	var topics []model.Topic
	var category model.Category
	var data map[string]string
	page := 0
	frame.BindJSON(&data)
	var query *gorm.DB
	query = Ctx.Db.Preload("Category")
	Ctx.Logger.Println(data)
	if catSlug, ok := data["category"]; ok {
		Ctx.Db.Where("slug = ?", catSlug).First(&category)
		if category.ID > 0 {
			query = query.Where("category_id = ?", category.ID)
		}
	}
	if searchTerm, ok := data["term"]; ok {
		if len(searchTerm) > 1 {
			query = query.
				Where(
				"title LIKE ? OR body LIKE ?",
					fmt.Sprint("%", searchTerm, "%"),
					fmt.Sprint("%", searchTerm, "%"),
				)
		}
	}
	if section, ok := data["section"]; ok {
		switch section {
		case "newTopics":
		case "recentlyViewed":
		case "participated":
		case "bookmarks":

		}
	}

	if rp, ok := data["page"]; ok {
		page, _ := strconv.Atoi(rp)
		if page <= 0 {
			page = 1;
		}
	}

	query.
		Select("id,title,slug,created_at,updated_at,category_id").
		Order("updated_at DESC,created_at DESC").
		Offset((page-1) * 100).
		Limit(100).
		Find(&topics)
	ts, err := json.Marshal(topics)
	if err == nil {
		Ctx.Ws.WriteJSON(Frame{Type: TopicListFrame, Data: string(ts)})
	} else {
		Ctx.Logger.WithError(err).Error()
	}
}

func (Ctx *AppContext) TopicDelete(frame Frame) {
	if Ctx.IsAuthorized() {
		var topic model.Topic
		err := frame.BindJSON(&topic)
		if err == nil {
			if topic.ID > 0 {
				var found model.Topic
				Ctx.Db.First(&found, topic.ID)
				if found.ID > 0 && found.User.ID == Ctx.User.ID {
					Ctx.Db.Delete(found)
				}
			}
		}
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
					err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&oldTopic).Error
				}
			} else {
				// this is new topic
				topic.ID = 0
				topic.User.ID = Ctx.User.ID
				fmt.Println(frame.Data)
				err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&topic).Error
				Ctx.Ws.WriteJSON(Frame{Type: TopicSaveFrame, Data: topic.ToJSON()})
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
