package handlers

import (
		"github.com/liverecord/lrs"
	"github.com/liverecord/lrs/common"
	"strings"
	"time"
)

// CommentList returns list of comments
func (Ctx *ConnCtx) CommentList(frame lrs.Frame) {
	var comments []lrs.Comment
	comments = make([]lrs.Comment, 0, 1)
	var creq	 lrs.CommentListRequest
	err := frame.BindJSON(&creq)
	if err != nil {
		Ctx.Logger.WithError(err)
		return
	}
	if creq.Page < 1 {
		creq.Page = 1
	}
	rows, err := Ctx.Db.
		Table("comments").
		Joins("JOIN users ON users.id = comments.user_id ").
		Joins("JOIN topics ON topics.id = comments.topic_id ").
		Joins("LEFT JOIN categories ON topics.category_id = categories.id ").
		Where("comments.spam = 0 AND comments.topic_id = ?", creq.TopicID).
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
		Offset(Ctx.Cfg.CommentsPerPage * (creq.Page - 1)).
		Limit(Ctx.Cfg.CommentsPerPage).
		Rows()

	total := uint(0)
	Ctx.Db.
		Table("comments").
		Where("comments.spam = 0 AND topic_id = ?", creq.TopicID).
		Count(&total)

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
		pagination := lrs.Pagination{Page: 1, Total: total, Limit: Ctx.Cfg.CommentsPerPage}
		resp := lrs.CommentListResponse{TopicID: creq.TopicID, Comments: comments, Pagination: pagination}
		Ctx.Pool.Write(Ctx.Ws, lrs.NewFrame(lrs.CommentListFrame, resp, frame.RequestID))
	} else {
		Ctx.Logger.WithError(err)
	}
}

func (Ctx *ConnCtx) CommentTyping(frame lrs.Frame) {
	if !Ctx.IsAuthorized() {
		return
	}
	type Typing struct {
		TopicID    uint64
	}
	var typing Typing
	err := frame.BindJSON(&typing)
	if err != nil {
		return
	}
	var topic lrs.Topic
	Ctx.Db.First(&topic, typing.TopicID)
	if topic.ID == 0 {
		return
	}
	if !topic.IsAccessibleBy(Ctx.User) {
		return
	}
	type TypingBroadcast struct {
		TopicID    uint64 `json:"topicId"`
		User	   lrs.User `json:"user"`
	}
	fr := lrs.NewFrame(lrs.CommentTypingFrame, TypingBroadcast{TopicID:topic.ID, User: Ctx.User.SafePluck()}, frame.RequestID)
	Ctx.BroadcastFrameToTopicConnections(topic, fr)
}

// CommentSave saves the comment
func (Ctx *ConnCtx) CommentSave(frame lrs.Frame) {
	if !Ctx.IsAuthorized() {
		Ctx.Logger.WithField("msg", "Unauthorized comment save call").Info()
		return
	}
	var comment lrs.Comment
	err := frame.BindJSON(&comment)
	Ctx.Logger.Debug("Decoded comment", comment)
	if err != nil {
		Ctx.Logger.WithError(err).Error("can't unmarshall comment")
		return
	}
	if comment.TopicID == 0 {
		return
	}
	var topic lrs.Topic
	Ctx.Db.First(&topic, comment.TopicID)
	if topic.ID == 0 {
		return
	}
	if !topic.IsAccessibleBy(Ctx.User) {
		return
	}
	if comment.ID > 0 {
		var existingComment lrs.Comment
		Ctx.Db.First(&existingComment, comment.ID)
		if existingComment.ID > 0 && existingComment.User.ID != Ctx.User.ID {
			// something is not right going on here
			Ctx.Logger.Warning("Unauthorized try to save a comment", comment, Ctx.User)
			return
		}
	}
	comment.User.ID = Ctx.User.ID
	comment.User = *Ctx.User
	comment.Body = common.SanitizeHtml(comment.Body)
	comment.Body = strings.TrimSpace(comment.Body)
	if len(comment.Body) < 1 {
		return
	}
	err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&comment).Error
	if err != nil {
		Ctx.Logger.WithError(err)
		return
	}
	Ctx.Db.Model(&topic).UpdateColumn("commented_at", time.Now())
	topic.MarkAsRead(Ctx.Db, Ctx.User)
	fr := lrs.NewFrame(lrs.CommentSaveFrame, comment, frame.RequestID)
	Ctx.Logger.Debugf("%v", topic.ACL)
	Ctx.BroadcastFrameToTopicConnections(topic, fr)
}

// BroadcastFrameToTopicConnections sends frame to topic subscribers
func (Ctx *ConnCtx) BroadcastFrameToTopicConnections(topic lrs.Topic, frame lrs.Frame) {
	if topic.Private {
		// broadcast only to people from acl or author
		Ctx.Pool.Send(&topic.User, frame)
		for u := range topic.ACL {
			Ctx.Pool.Send(&topic.ACL[u], frame)
		}
	} else {
		// broadcast to everyone, who has any connection to this topic
		// find everyone, who viewed or commented this topic
		Ctx.Pool.Broadcast(frame, nil)
	}
}