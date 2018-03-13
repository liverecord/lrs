package handlers

import (
	"encoding/json"

	"fmt"
	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/model"
)

func (Ctx *AppContext) CommentList(frame Frame) {
	var comments []model.Comment
	var topic model.Topic
	err := frame.BindJSON(&topic)
	if err != nil {
		Ctx.Logger.WithError(err)
		return
	}
	rows, err := Ctx.Db.
		Table("comments").
		Preload("users").
		Joins("JOIN topics ON topics.id = comments.topic_id ").
		Joins("LEFT JOIN categories ON topics.category_id = categories.id ").
		Where("topic_id = ?", topic.ID).
		Group("comments.id").
		Order("comments.created_at DESC").
		Select("comments.*, " +
			"topics.title as topic_title, " +
			"topics.slug as topic_slug, " +
			"topics.id as topic_id, " +
			"topics.category_id as category_id, " +
			"categories.slug as category_slug, " +
			"categories.name as category_name ").
		Rows()

	type CommentTopic struct {
		topic_id   uint
		topicSlug  string
		topicTitle string
	}

	type CommentCategory struct {
		categoryId uint
		categorySlug string
		categoryName string
	}

	// comments

	if err == nil {
		for rows.Next() {
			var comm model.Comment
			var commTopic CommentTopic
			var commCat CommentCategory

			if err := Ctx.Db.ScanRows(rows, &comm); err != nil {
				Ctx.Logger.Errorf("should get no error, but got %v", err)
			}

			if err := Ctx.Db.ScanRows(rows, &commTopic); err == nil {
				comm.Topic.ID = commTopic.topic_id
				comm.Topic.Slug = commTopic.topicSlug
				comm.Topic.Title = commTopic.topicTitle
				Ctx.Logger.Debugln(commTopic)
			}

			if err := Ctx.Db.ScanRows(rows, &commCat); err == nil {
				comm.Topic.Category.ID = commCat.categoryId
				comm.Topic.Category.Slug = commCat.categorySlug
				comm.Topic.Category.Name = commCat.categoryName
				Ctx.Logger.Debugln(commCat)
			}

			comments = append(comments, comm)
		}
		defer rows.Close()

	Ctx.Logger.Debugf("%v", comments)
		cats, _ := json.Marshal(comments)
		Ctx.Ws.WriteJSON(Frame{Type: CommentListFrame, Data: string(cats)})
	} else {
		Ctx.Logger.WithError(err)
	}
}

func (Ctx *AppContext) CommentSave(frame Frame) {
	if Ctx.IsAuthorized() {
		var comment model.Comment
		err := frame.BindJSON(&comment)
		Ctx.Logger.Info("Decoded comment", comment)
		Ctx.Logger.Info("User", Ctx.User)
		if err == nil {
			comment.User.ID = Ctx.User.ID
			if comment.ID > 0 {
				Ctx.Logger.WithField("msg", "Comment updates not supported yet").Info()
			} else {
				comment.ID = 0
				fmt.Println(frame.Data)
				err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&comment).Error
				Ctx.Ws.WriteJSON(NewFrame(CommentSaveFrame, comment, frame.RequestID))
			}
			if err != nil {
				Ctx.Logger.WithError(err).Error("Unable to save comment")
			}
		} else {
			Ctx.Logger.WithError(err).Error("can't unmarshall comment")
		}
	} else {
		Ctx.Logger.WithField("msg", "Unauthorized comment save call").Info()
	}
}
