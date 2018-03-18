package handlers

import (
	"encoding/json"

	"fmt"
	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/model"
)

func (Ctx *AppContext) CommentList(frame Frame) {
	var comments []model.Comment
	comments = make([]model.Comment, 0, 1)
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
		topic_id    uint
		topic_slug  string
		topic_title string
	}

	type CommentCategory struct {
		category_id   uint
		category_slug string
		category_name string
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
				comm.Topic.Slug = commTopic.topic_slug
				comm.Topic.Title = commTopic.topic_title
				Ctx.Logger.Debugln(commTopic)
			}

			if err := Ctx.Db.ScanRows(rows, &commCat); err == nil {
				comm.Topic.Category.ID = commCat.category_id
				comm.Topic.Category.Slug = commCat.category_slug
				comm.Topic.Category.Name = commCat.category_name
				Ctx.Logger.Debugln(commCat)
			}

			comments = append(comments, comm)
		}
		defer rows.Close()
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
