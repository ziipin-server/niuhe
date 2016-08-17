package adminbase

import (
	"github.com/ziipin-server/niuhe"
)

type AdminCrudView struct {
	ctrl *AdminCrudViewCtrl
}

func NewAdminCrudView(newModel interface{}, ctrl *AdminCrudViewCtrl) *AdminCrudView {
	if ctrl == nil {
		panic("AdminCrudViewCtrl cannot be nil")
	}
	ctrl.Init(newModel)
	return &AdminCrudView{
		ctrl: ctrl,
	}
}

func (view *AdminCrudView) List_GET(c *niuhe.Context) {
	page, pageSize, total, rows, err := view.ctrl.GetPage(c)
	if err != nil {
		view.ctrl.ReturnError(c, err)
	} else {
		view.ctrl.ReturnOK(c, Any{
			"rows":     rows,
			"page":     page,
			"pagesize": pageSize,
			"total":    total,
		})
	}
}

func (view *AdminCrudView) Edit_GET(c *niuhe.Context) {
	model, err := view.ctrl.GetEditModel(c)
	if err != nil {
		view.ctrl.ReturnError(c, err)
	} else if model == nil {
		view.ctrl.ReturnError(c, niuhe.NewCommError(404, "target not exists"))
	} else {
		view.ctrl.ReturnOK(c, model)
	}
}

func (view *AdminCrudView) Edit_POST(c *niuhe.Context) {
	err := view.ctrl.SaveEditModel(c)
	if err != nil {
		niuhe.LogError("Edit Post fail: %s", err.Error())
		view.ctrl.ReturnError(c, err)
	} else {
		view.ctrl.ReturnOK(c, nil)
	}
}

func (view *AdminCrudView) Add_GET(c *niuhe.Context) {
	model, err := view.ctrl.GetAddModel(c)
	if err != nil {
		view.ctrl.ReturnError(c, err)
	} else {
		view.ctrl.ReturnOK(c, model)
	}
}

func (view *AdminCrudView) Add_POST(c *niuhe.Context) {
	err := view.ctrl.SaveAddModel(c)
	if err != nil {
		niuhe.LogError("Add Post fail: %s", err.Error())
		view.ctrl.ReturnError(c, err)
	} else {
		view.ctrl.ReturnOK(c, nil)
	}
}

func (view *AdminCrudView) Del_POST(c *niuhe.Context) {
	if err := view.ctrl.Del(c); err != nil {
		view.ctrl.ReturnError(c, err)
	} else {
		view.ctrl.ReturnOK(c, nil)
	}
}
