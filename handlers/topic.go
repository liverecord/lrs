package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	"errors"

	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/model"
)

func Topic(ctx *AppContext, frame Frame) (Frame, error) {
	var topic model.Topic

	var data map[string]string
	frame.BindJSON(&data)

	slug, ok := data["slug"]
	if !ok {
		return Frame{}, errors.New("can not process topic without slug")
	}

	ctx.Db.
		Preload("Category").
		Preload("User").
		Where("slug = ?", slug).First(&topic)

	if topic.ID == 0 {
		return Frame{}, fmt.Errorf("could not find topic %s", slug)
	}

	ts, err := json.Marshal(topic)
	if err != nil {
		return Frame{}, err
	}

	ctx.Db.Model(&topic).UpdateColumn("total_views", gorm.Expr("total_views + ?", 1))

	return Frame{Type: TopicFrame, Data: string(ts)}, nil
}

func TopicList(ctx *AppContext, frame Frame) (Frame, error) {
	var topics []model.Topic
	var category model.Category
	var data map[string]string
	page := 0
	frame.BindJSON(&data)

	var query *gorm.DB
	query = ctx.Db.Preload("Category")
	ctx.Logger.Println(data)

	if catSlug, ok := data["category"]; ok {
		ctx.Db.Where("slug = ?", catSlug).First(&category)
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
		Select("id,title,slug,created_at,updated_at,category_id").
		Order("updated_at DESC,created_at DESC").
		Offset((page - 1) * 100).
		Limit(100).
		Find(&topics)

	ts, err := json.Marshal(topics)
	if err != nil {
		return Frame{}, err
	}

	return Frame{Type: TopicListFrame, Data: string(ts)}, nil
}

func TopicDelete(ctx *AppContext, frame Frame) (Frame, error) {
	if !ctx.IsAuthorized() {
		return Frame{}, errors.New("we can delete topic only with authorized request")
	}

	var topic model.Topic
	err := frame.BindJSON(&topic)
	if err != nil {
		return Frame{}, fmt.Errorf("could not parse data: %s", frame.Data)
	}

	if topic.ID == 0 {
		return Frame{}, errors.New("we do not have topic id in data")
	}

	var found model.Topic
	ctx.Db.First(&found, topic.ID)

	if found.ID == 0 || found.User.ID != ctx.User.ID {
		return Frame{}, errors.New("we didn't find topic or we got request not from author")
	}

	ctx.Db.Delete(found)
	return Frame{}, err
}

func TopicSave(ctx *AppContext, frame Frame) (Frame, error) {
	if !ctx.IsAuthorized() {
		return Frame{}, errors.New("we can not save unauthorized user's topic")
	}

	var topic model.Topic
	err := frame.BindJSON(&topic)
	if err != nil {
		return Frame{}, errors.New("could not unmarshall topic")
	}

	ctx.Logger.Info("Decoded topic", topic)
	ctx.Logger.Info("User", ctx.User)

	// existing topic
	if topic.ID > 0 {
		return Frame{}, updateTopic(ctx, topic)
	}

	// new topic
	topic.ID = 0
	topic.User.ID = ctx.User.ID
	err = ctx.Db.Set("gorm:association_autoupdate", false).Save(&topic).Error
	return Frame{Type: TopicUpdateFrame, Data: topic.ToJSON()}, nil
}

func updateTopic(ctx *AppContext, topic model.Topic) error {
	// find topic in DB and update it
	var oldTopic model.Topic
	ctx.Db.Where("id = ?", topic.ID).First(&oldTopic)

	if oldTopic.ID == 0 {
		return fmt.Errorf("could not find topic with id: %d", topic.ID)
	}

	oldTopic.Title = topic.Title
	//oldTopic.Acl = topic.Acl
	oldTopic.Body = topic.Body
	return ctx.Db.Set("gorm:association_autoupdate", false).Save(&oldTopic).Error
}
