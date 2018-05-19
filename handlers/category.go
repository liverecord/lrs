package handlers

// CategoryList sends list of categories
func (Ctx *AppContext) CategoryList(frame lrs.Frame) {
	var categories []lrs.Category
	Ctx.Db.Find(&categories)
	Ctx.Pool.Write(Ctx.Ws, lrs.NewFrame(lrs.CategoryListFrame, categories, frame.RequestID))
}

// CategorySave saves the category
func (Ctx *AppContext) CategorySave(frame lrs.Frame) {
	if Ctx.IsAuthorized() {
		var category lrs.Category
		err := frame.BindJSON(&category)
		Ctx.Logger.Info("Decoded category", category)
		Ctx.Logger.Info("User", Ctx.User)
		if err != nil {
			Ctx.Logger.WithError(err).Error("Can't unmarshall category")
			return
		}
		if category.ID > 0 {
			Ctx.Logger.WithField("msg", "Category updates not supported yet").Info()
		} else {
			category.ID = 0
			err = Ctx.Db.Set("gorm:association_autoupdate", false).Save(&category).Error
			Ctx.Pool.Write(Ctx.Ws, lrs.NewFrame(lrs.CategorySaveFrame, category, frame.RequestID))
		}
		if err != nil {
			Ctx.Logger.WithError(err).Error("Unable to save category")
		}
	} else {
		Ctx.Logger.WithField("msg", "Unauthorized category save call").Info()
	}
}
