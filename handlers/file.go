package handlers

import (
	"fmt"

	. "github.com/liverecord/server"
)

func (Ctx *AppContext) CategoryList(frame Frame) {
	var categories []Category
	Ctx.Db.Find(&categories)
	Ctx.Pool.Write(Ctx.Ws, NewFrame(CategoryListFrame, categories, frame.RequestID))
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

				Ctx.Pool.Write(Ctx.Ws, NewFrame(CategorySaveFrame, category, frame.RequestID))

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
