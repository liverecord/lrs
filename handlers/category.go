package handlers

import (
	"encoding/json"

	"fmt"
	. "github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/model"
)

func CategoryList(ctx *AppContext, f Frame) (Frame, error) {
	var categories []model.Category
	ctx.Db.Find(&categories)
	cats, err := json.Marshal(categories)
	if err != nil {
		return Frame{}, err
	}

	return Frame{Type: CategoryListFrame, Data: string(cats)}, nil
}

func (Ctx *AppContext) CategorySave(frame Frame) {
	if Ctx.IsAuthorized() {
		var category model.Category
		err := frame.BindJSON(&category)
		Ctx.Logger.Info("Decoded category", category)
		Ctx.Logger.Info("User", Ctx.User)
		if err == nil {
			if category.ID > 0 {
				Ctx.Logger.WithField("msg", "Category updates not supported yet").Info()
			} else {
				category.ID = 0
				fmt.Println(frame.Data)
				err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&category).Error
				Ctx.Ws.WriteJSON(Frame{Type: CategorySaveFrame, Data: category.ToJSON()})
			}
			if err != nil {
				Ctx.Logger.WithError(err).Error("Unable to save category")
			}
		} else {
			Ctx.Logger.WithError(err).Error("can't unmarshall category")
		}
	} else {
		Ctx.Logger.WithField("msg", "Unauthorized category save call").Info()
	}
}
