package handlers

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	server "github.com/liverecord/lrs"
)

func (Ctx *AppContext) Topic(frame server.Frame) {
	var topic server.Topic

	var data map[string]string
	frame.BindJSON(&data)

	if slug, ok := data["slug"]; ok {
		Ctx.Db.
			Preload("Category").
			Preload("User").
			Where("slug = ?", slug).First(&topic)
		if topic.ID > 0 {
			f := server.NewFrame(server.TopicFrame, topic, frame.RequestID)
			Ctx.Pool.Write(Ctx.Ws, f)
			Ctx.CommentList(f)
			Ctx.Db.Model(&topic).UpdateColumn("total_views", gorm.Expr("total_views + ?", 1))
		}
	}
}

func (Ctx *AppContext) TopicList(frame server.Frame) {
	var topics []server.Topic
	var category server.Category
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
			page = 1
		}
	}

	query.
		Preload("User").
		//Select("id,title,slug,created_at,updated_at,category_id").
		Order("updated_at DESC,created_at DESC").
		Offset((page - 1) * 100).
		Limit(100).
		Find(&topics)
	f := server.NewFrame(server.TopicListFrame, topics, frame.RequestID)
	Ctx.Pool.Write(Ctx.Ws, f)
}

func (Ctx *AppContext) TopicDelete(frame server.Frame) {
	if Ctx.IsAuthorized() {
		var topic server.Topic
		err := frame.BindJSON(&topic)
		if err == nil {
			if topic.ID > 0 {
				var found server.Topic
				Ctx.Db.First(&found, topic.ID)
				if found.ID > 0 && found.User.ID == Ctx.User.ID {
					Ctx.Db.Delete(found)
				}
			}
		}
	}
}

func (Ctx *AppContext) TopicSave(frame server.Frame) {
	if Ctx.IsAuthorized() {
		var topic server.Topic
		err := frame.BindJSON(&topic)
		Ctx.Logger.Info("Decoded topic", topic)
		Ctx.Logger.Info("User", Ctx.User)
		if err == nil {
			topic.Private = len(topic.Acl) > 0
			if topic.ID > 0 {
				// find topic in DB and update it
				var oldTopic server.Topic
				Ctx.Db.Where("id = ?", topic.ID).First(&oldTopic)
				if oldTopic.ID > 0 {
					oldTopic.Title = topic.Title
					oldTopic.Acl = topic.Acl
					oldTopic.Body = topic.Body
					err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&oldTopic).Error
					f := server.NewFrame(server.TopicSaveFrame, topic, frame.RequestID)
					Ctx.Pool.Write(Ctx.Ws, f)
					if topic.Private == false {
						Ctx.Pool.Broadcast(server.NewFrame(server.TopicSaveFrame, topic, ""))
					}

				}
			} else {
				// this is new topic
				topic.ID = 0
				topic.User.ID = Ctx.User.ID
				//topic.
				fmt.Println(frame.Data)
				err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&topic).Error
				f := server.NewFrame(server.TopicSaveFrame, topic, frame.RequestID)
				Ctx.Pool.Write(Ctx.Ws, f)
				if topic.Private == false {
					Ctx.Pool.Broadcast(server.NewFrame(server.TopicSaveFrame, topic, ""))
				}
			}
			if err != nil {
				Ctx.Logger.WithError(err).Error("Unable to save topic")
			}
		} else {
			Ctx.Logger.WithError(err).Error("Can't unmarshall topic")
		}
	} else {
		Ctx.Logger.WithField("msg", "Unauthorized topic save call").Info()
	}
}
