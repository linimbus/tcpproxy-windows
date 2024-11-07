package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func ShowToolBar(cfg *LinkConfig) {
	var dlg *walk.Dialog
	var acceptPB *walk.PushButton

	cnt, err := Dialog{
		AssignTo:      &dlg,
		Title:         "Link Detail",
		Icon:          walk.IconInformation(),
		DefaultButton: &acceptPB,
		Size:          Size{250, 150},
		MinSize:       Size{250, 150},
		Layout: VBox{
			Alignment: AlignHNearVCenter,
			Margins:   Margins{Top: 10, Bottom: 10, Left: 10, Right: 10},
		},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "Listen Address:",
					},
					Label{
						Text: cfg.Address,
					},
					Label{
						Text: "Listen Port:",
					},
					Label{
						Text: fmt.Sprintf("%d", cfg.Port),
					},
					Label{
						Text: "Listen Tls:",
					},
					Label{
						Text: cfg.Tls,
					},
					Label{
						Text: "Backend Address:",
					},
					Label{
						Text: cfg.Backend.Address,
					},
					Label{
						Text: "Backend Port:",
					},
					Label{
						Text: fmt.Sprintf("%d", cfg.Backend.Port),
					},
					Label{
						Text: "Backend Tls:",
					},
					Label{
						Text: cfg.Backend.Tls,
					},
				},
			},
			Composite{
				Layout: VBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(MainWindowsCtrl())
	if err != nil {
		logs.Error(err.Error())
	} else {
		logs.Info("show link dialog return %d", cnt)
	}
}
