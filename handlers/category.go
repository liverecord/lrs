package handlers

import (
	"encoding/json"
	"fmt"

	. "github.com/liverecord/server"
)

func (Ctx *AppContext) CategoryList(frame Frame) {
	var categories []Category
	Ctx.Db.Find(&categories)
	cats, err := json.Marshal(categories)
	if err == nil {
		Ctx.Ws.WriteJSON(Frame{Type: CategoryListFrame, Data: string(cats)})
	} else {
		Ctx.Logger.WithError(err)
	}
}

func (Ctx *AppContext) CategorySave(frame Frame) {
	if Ctx.IsAuthorized() {
		var category Category
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
