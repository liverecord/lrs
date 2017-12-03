package handlers

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"

	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/model"

)

func  (Ctx *AppContext) CategoryList(frame Frame) {
	var categories []model.Category
	Ctx.Db.Find(&categories)
	cats, err := json.Marshal(categories)
	if err == nil {
		Ctx.Ws.WriteJSON(Frame{Type: CategoryListFrame, Data: string(cats)})
	} else {
		logrus.WithError(err)
	}
}
