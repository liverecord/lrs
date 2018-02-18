package handlers

import (
	"encoding/json"

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
