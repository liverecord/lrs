package handlers

import (
	"fmt"
	"github.com/liverecord/lrs"
)

// CommentList returns list of comments
func (Ctx *ConnCtx) CommentList(frame lrs.Frame) {
	var comments []lrs.Comment
	comments = make([]lrs.Comment, 0, 1)
	var topic lrs.Topic
	err := frame.BindJSON(&topic)
	if err != nil {
		Ctx.Logger.WithError(err)
		return
	}
	rows, err := Ctx.Db.
		Table("comments").
		Joins("JOIN users ON users.id = comments.user_id ").
		Joins("JOIN topics ON topics.id = comments.topic_id ").
		Joins("LEFT JOIN categories ON topics.category_id = categories.id ").
		Where("topic_id = ?", topic.ID).
		Group("comments.id").
		Order("comments.created_at DESC").
		Select("comments.*, " +
			"users.name as user_name, " +
			"users.slug as user_slug, " +
			"users.picture as user_picture, " +
			"users.rank as user_rank, " +
			"users.online as user_online, " +
			"topics.title as topic_title, " +
			"topics.slug as topic_slug, " +
			"topics.id as topic_id, " +
			"topics.category_id as category_id, " +
			"categories.slug as category_slug, " +
			"categories.name as category_name ").
		Rows()

	type CommentUser struct {
		UserSlug    string
		UserName    string
		UserRank    float32
		UserOnline  bool
		UserPicture string
	}

	type CommentTopic struct {
		TopicID    uint64
		TopicSlug  string
		TopicTitle string
	}

	type CommentCategory struct {
		CategoryID   uint64
		CategorySlug string
		CategoryName string
	}

	// comments

	if err == nil {
		for rows.Next() {
			var comm lrs.Comment
			var commTopic CommentTopic
			var commCat CommentCategory
			var commUser CommentUser

			if err := Ctx.Db.ScanRows(rows, &comm); err != nil {
				Ctx.Logger.Errorf("should get no error, but got %v", err)
			}

			if err := Ctx.Db.ScanRows(rows, &commUser); err == nil {
				comm.User.ID = comm.UserID
				comm.User.Slug = commUser.UserSlug
				comm.User.Name = commUser.UserName
				comm.User.Online = commUser.UserOnline
				comm.User.Picture = commUser.UserPicture
				comm.User.Rank = commUser.UserRank
				comm.User = comm.User.SafePluck()
			}

			if err := Ctx.Db.ScanRows(rows, &commTopic); err == nil {
				comm.Topic.ID = comm.TopicID
				comm.Topic.Slug = commTopic.TopicSlug
				comm.Topic.Title = commTopic.TopicTitle
				Ctx.Logger.Debugln(commTopic)
			}

			if err := Ctx.Db.ScanRows(rows, &commCat); err == nil {
				comm.Topic.CategoryID = commCat.CategoryID
				comm.Topic.Category.ID = commCat.CategoryID
				comm.Topic.Category.Slug = commCat.CategorySlug
				comm.Topic.Category.Name = commCat.CategoryName
				Ctx.Logger.Debugln(commCat)
			}

			comments = append(comments, comm)
		}
		defer rows.Close()
		Ctx.Pool.Write(Ctx.Ws, lrs.NewFrame(lrs.CommentListFrame, comments, frame.RequestID))
	} else {
		Ctx.Logger.WithError(err)
	}
}

// BroadcastComment sends comment to topic subscribers
func (Ctx *ConnCtx) BroadcastComment() {

}

// CommentSave saves the comment
func (Ctx *ConnCtx) CommentSave(frame lrs.Frame) {
	if Ctx.IsAuthorized() {
		var comment lrs.Comment
		err := frame.BindJSON(&comment)
		Ctx.Logger.Info("Decoded comment", comment)
		Ctx.Logger.Info("User", Ctx.User)
		if err == nil {
			comment.User.ID = Ctx.User.ID
			comment.User = *Ctx.User

			if comment.TopicID > 0 {
				var topic lrs.Topic
				Ctx.Db.First(&topic, comment.TopicID)
				if topic.ID > 0 {
					Ctx.Logger.Debugf("%v", topic.ACL)
					if topic.Private {
						// broadcast only to people from acl or author
						fr := lrs.NewFrame(lrs.CommentSaveFrame, comment, frame.RequestID)
						Ctx.Pool.Send(&topic.User, fr)
						for u := range topic.ACL {
							Ctx.Pool.Send(&topic.ACL[u], fr)
						}

					} else {
						// broadcast to everyone, who has any connection to this topic
						// find everyone, who viewed or commented this topic

					}
				}

				if comment.ID > 0 {
					Ctx.Logger.WithField("msg", "Comment updates not supported yet").Info()
				} else {
					comment.ID = 0
					fmt.Println(frame.Data)
					err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&comment).Error
					if err == nil {

						Ctx.Pool.Broadcast(lrs.NewFrame(lrs.CommentSaveFrame, comment, frame.RequestID))
					}
				}
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
