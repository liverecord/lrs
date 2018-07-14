package handlers

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	server "github.com/liverecord/lrs"
	"time"
)

// Topic sends topic data
func (Ctx *ConnCtx) Topic(frame server.Frame) {
	var topic server.Topic
	var data map[string]string
	frame.BindJSON(&data)
	if slug, ok := data["slug"]; ok {
		Ctx.Db.
			Preload("Category").
			Preload("User").
			Preload("ACL").
			Where("slug = ?", slug).First(&topic)
		if topic.ID > 0 {
			Ctx.ViewTopic(&topic)
			topic.SafeTopic()
			f := server.NewFrame(server.TopicFrame, topic, frame.RequestID)
			Ctx.Pool.Write(Ctx.Ws, f)
			Ctx.CommentList(f)

		}
	}
}

func (Ctx *ConnCtx) ViewTopic(topic *server.Topic) {

	if Ctx.IsAuthorized() {
		var topicStatus server.TopicStatus
		now := time.Now()
		Ctx.Db.
			Where(server.TopicStatus{UserID: Ctx.User.ID, TopicID: topic.ID}).
			Assign(server.TopicStatus{ReadAt: &now, NotifiedAt: &now}).
			FirstOrCreate(&topicStatus)
	}
	Ctx.Db.Model(&topic).UpdateColumn("total_views", gorm.Expr("total_views + ?", 1))
	//Ctx.Db.Model(&topic).UpdateColumn("commented_at", time.Now())
}

// TopicList returns list of topics
func (Ctx *ConnCtx) TopicList(frame server.Frame) {
	var topics []server.Topic
	var data map[string]string
	page := 0
	frame.BindJSON(&data)
	var query *gorm.DB
	query = Ctx.Db.Table("topics t").Unscoped()
	Ctx.Logger.Println(data)
	CategoryMissed := true
	if catSlug, ok := data["category"]; ok && catSlug != "-" && catSlug != "" {
		query = query.Joins("INNER JOIN categories tc ON (t.category_id = tc.id AND tc.`deleted_at` IS NULL AND tc.slug = ?)", catSlug)
		CategoryMissed = false
	}
	if CategoryMissed  {
		query = query.Joins("INNER JOIN categories tc ON (t.category_id = tc.id AND tc.`deleted_at` IS NULL)")
	}
	UserID := uint64(0)
	if Ctx.User != nil && Ctx.User.ID > 0 {
		UserID = Ctx.User.ID
		query = query.
			Joins("LEFT JOIN topic_acl tacl ON (t.id = tacl.topic_id)").
			Where("t.private = 0 OR t.private IS NULL OR t.user_id = ? OR tacl.user_id = ?", UserID, UserID)
	} else {
		query = query.Where("t.private = 0 OR t.private IS NULL")
	}

	if searchTerm, ok := data["term"]; ok {
		if len(searchTerm) > 1 {
			query = query.
				Where(
					"t.title LIKE ? OR t.body LIKE ?",
					fmt.Sprint("%", searchTerm, "%"),
					fmt.Sprint("%", searchTerm, "%"),
				)
		}
	}

	if section, ok := data["section"]; ok {
		switch section {
		case "newTopics":
			query = query.
				Joins("LEFT JOIN topic_statuses ts ON (t.id = ts.topic_id AND ts.user_id = ? )", UserID)
		case "recentlyViewed":
			query = query.
				Joins("INNER JOIN topic_statuses ts ON (t.id = ts.topic_id AND ts.user_id = ? )", UserID)
		case "participated":
			// ORDER BY unread_comments DESC
			query = query.Where("cmts.user_id = ?", UserID).
				Joins("INNER JOIN topic_statuses ts ON (t.id = ts.topic_id AND ts.user_id = ? )", UserID)
		case "bookmarks":
			// ts.vote = 1
			query = query.
				Joins("INNER JOIN topic_statuses ts ON (t.id = ts.topic_id AND ts.user_id = ? )", UserID).
				Where("ts.vote = ?", 1)
		}
	}

	query = query.
		Joins("LEFT JOIN comments cmts ON " +
		"(t.id = cmts.topic_id AND	cmts.`deleted_at` IS NULL AND cmts.`spam` = 0 AND (cmts.`created_at` > ts.read_at OR ts.`read_at` is null))	").
		Where("t.`deleted_at` IS NULL")

	if rp, ok := data["page"]; ok {
		page, _ := strconv.Atoi(rp)
		if page <= 0 {
			page = 1
		}
	}

	query.
		Preload("User").
		Preload("Category").
		Select(
			"t.id, t.title, t.slug, t.category_id, t.user_id, cmts.id, "+
				"t.created_at, t.updated_at, t.rank, t.pinned, t.total_views, t.total_comments, " +
			"ts.vote, ts.favorite," +
			"tc.slug category_slug, tc.name category_name," +
			"COUNT(cmts.id) as unread_comments",
		).
		Group("t.id").
		Order("t.updated_at DESC, t.created_at DESC").
		Offset((page - 1) * 100).
		Limit(100).
		Find(&topics)
	for _, v := range topics {
		v.SafeTopic()
	}
	f := server.NewFrame(server.TopicListFrame, topics, frame.RequestID)
	Ctx.Pool.Write(Ctx.Ws, f)
}

// TopicDelete destroys the topic
func (Ctx *ConnCtx) TopicDelete(frame server.Frame) {
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

// TopicSave saves the topic
func (Ctx *ConnCtx) TopicSave(frame server.Frame) {
	if !Ctx.IsAuthorized() {
		Ctx.Logger.WithField("msg", "Unauthorized topic save call").Info()
		return
	}
	var topic server.Topic
	err := frame.BindJSON(&topic)
	Ctx.Logger.Info("Decoded topic", topic)
	Ctx.Logger.Info("User", Ctx.User)
	if err != nil {
		Ctx.Logger.WithError(err).Error("Can't unmarshall topic")
		return
	}
	topic.Private = len(topic.ACL) > 0
	if topic.ID > 0 {
		// find topic in DB and update it
		var oldTopic server.Topic
		Ctx.Db.Where("id = ?", topic.ID).First(&oldTopic)
		if oldTopic.ID > 0 {
			oldTopic.Title = topic.Title
			oldTopic.ACL = topic.ACL
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
}
