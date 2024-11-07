package main

import (
	"fmt"
	"net"

	"github.com/astaxie/beego/logs"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type BackendConfig struct {
	Address  string `json:"Address"`
	Port     int    `json:"Port"`
	Protocol string `json:"Protocol"`
	Tls      string `json:"Tls"`
	Timeout  int    `json:"Timeout"`
}

type LinkConfig struct {
	Address  string        `json:"Address"`
	Port     int           `json:"Port"`
	Protocol string        `json:"Protocol"`
	Tls      string        `json:"Tls"`
	Backend  BackendConfig `json:"Backend"`
}

func IfaceOptions() []string {
	output := []string{"0.0.0.0", "::"}

	ifaces, err := net.Interfaces()
	if err != nil {
		logs.Error(err.Error())
		return output
	}
	for _, v := range ifaces {
		if v.Flags&net.FlagUp == 0 {
			continue
		}
		address, err := InterfaceAddsGet(&v)
		if err != nil {
			continue
		}
		for _, addr := range address {
			output = append(output, addr.String())
		}
	}
	return output
}

type BackendItem struct {
	Index   int
	Address string
	Tls     bool
	Timeout int
}

func AddToolBar() {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton

	var consoleIface *walk.ComboBox
	var consolePort *walk.NumberEdit
	var consoleTls *walk.ComboBox
	var consoleProtocol *walk.ComboBox

	var backendAddr *walk.LineEdit
	var backendPort *walk.NumberEdit
	var backendTls *walk.ComboBox
	var backendProtocol *walk.ComboBox
	var backendTimeout *walk.NumberEdit

	var addLink LinkConfig
	var backend BackendConfig

	addLink.Address = "0.0.0.0"
	addLink.Port = 8080
	addLink.Tls = "NULL"
	addLink.Protocol = "tcp"

	backend.Port = 8080
	backend.Tls = "NULL"
	backend.Protocol = "tcp"
	backend.Timeout = 0

	cnt, err := Dialog{
		AssignTo:      &dlg,
		Title:         "Add Link",
		Icon:          ICON_TOOL_ADD,
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		Size:          Size{150, 250},
		MinSize:       Size{150, 250},
		Layout:        VBox{Margins: Margins{Top: 10, Bottom: 10, Left: 10, Right: 10}},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "Listen Address:",
					},
					ComboBox{
						AssignTo:     &consoleIface,
						CurrentIndex: 0,
						Model:        IfaceOptions(),
						OnCurrentIndexChanged: func() {
							addLink.Address = consoleIface.Text()
						},
					},
					Label{
						Text: "Listen Port:",
					},
					NumberEdit{
						AssignTo:    &consolePort,
						Value:       float64(addLink.Port),
						ToolTipText: "1~65535",
						MaxValue:    65535,
						MinValue:    1,
						OnValueChanged: func() {
							addLink.Port = int(consolePort.Value())
						},
					},
					Label{
						Text: "Listen Tls:",
					},
					ComboBox{
						AssignTo:     &consoleTls,
						CurrentIndex: 0,
						Model:        []string{"NULL", "TLS1.2", "TLS1.3"},
						OnCurrentIndexChanged: func() {
							addLink.Tls = consoleTls.Text()
						},
					},
					Label{
						Text: "Listen Protocol:",
					},
					ComboBox{
						AssignTo:     &consoleProtocol,
						CurrentIndex: 0,
						Model:        []string{"tcp", "tcp4", "tcp6"},
						OnCurrentIndexChanged: func() {
							addLink.Protocol = consoleProtocol.Text()
						},
					},
					Label{
						Text: "Backend Address:",
					},
					LineEdit{
						AssignTo:  &backendAddr,
						CueBanner: "192.168.1.2",
						Text:      "",
						OnTextChanged: func() {
							backend.Address = backendAddr.Text()
						},
					},
					Label{
						Text: "Backend Port:",
					},
					NumberEdit{
						AssignTo:    &backendPort,
						Value:       float64(8080),
						ToolTipText: "1~65535",
						MaxValue:    65535,
						MinValue:    1,
						OnValueChanged: func() {
							backend.Port = int(backendPort.Value())
						},
					},
					Label{
						Text: "Backend Tls:",
					},
					ComboBox{
						AssignTo:     &backendTls,
						CurrentIndex: 0,
						Model:        []string{"NULL", "TLS1.2", "TLS1.3"},
						OnCurrentIndexChanged: func() {
							backend.Tls = backendTls.Text()
						},
					},

					Label{
						Text: "Backend Protocol:",
					},
					ComboBox{
						AssignTo:     &backendProtocol,
						CurrentIndex: 0,
						Model:        []string{"tcp", "tcp4", "tcp6"},
						OnCurrentIndexChanged: func() {
							backend.Protocol = backendProtocol.Text()
						},
					},
					Label{
						Text: "Backend Timeout:",
					},
					NumberEdit{
						AssignTo:    &backendTimeout,
						Value:       float64(0),
						ToolTipText: "0~60",
						MaxValue:    60,
						MinValue:    0,
						Suffix:      " Second",
						OnValueChanged: func() {
							backend.Timeout = int(backendTimeout.Value())
						},
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &acceptPB,
						Text:     "Add",
						OnClicked: func() {
							acceptPB.SetEnabled(false)
							cancelPB.SetEnabled(false)

							go func() {
								defer func() {
									acceptPB.SetEnabled(true)
									cancelPB.SetEnabled(true)
								}()

								if !ListenCheck(addLink.Address, addLink.Port) {
									ErrorBoxAction(dlg,
										fmt.Sprintf("Address %s:%d binding failed!",
											addLink.Address, addLink.Port))
									return
								}

								if backend.Address == "" {
									ErrorBoxAction(dlg, "Backend address is empty!")
									return
								}

								addLink.Backend = backend
								err := LinkAdd(addLink)
								if err != nil {
									ErrorBoxAction(dlg, err.Error())
									return
								}

								dlg.Accept()
							}()
						},
					},
					PushButton{
						AssignTo: &cancelPB,
						Text:     "Cancel",
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
		logs.Info("add link dialog return %d", cnt)
	}
}
