package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/astaxie/beego/logs"
)

func VersionGet() string {
	return "v1.1.0"
}

func ListenCheck(addr string, port int) bool {
	list, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		logs.Error(err.Error())
		return false
	}
	defer list.Close()
	return true
}

func SaveToFile(name string, body []byte) error {
	return os.WriteFile(name, body, 0664)
}

func InterfaceAddsGet(iface *net.Interface) ([]net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, 0)
	for _, v := range addrs {
		ipone, _, err := net.ParseCIDR(v.String())
		if err != nil {
			continue
		}
		if len(ipone) > 0 {
			ips = append(ips, ipone)
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("interface not any address.")
	}
	return ips, nil
}

func CapSignal(proc func()) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	go func() {
		sig := <-signalChan
		proc()
		logs.Error("recv signcal %s, ready to exit", sig.String())
		os.Exit(-1)
	}()
}

func ByteView(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < (1024 * 1024) {
		return fmt.Sprintf("%.1fKB", float64(size)/float64(1024))
	} else if size < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.1fMB", float64(size)/float64(1024*1024))
	} else if size < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.1fGB", float64(size)/float64(1024*1024*1024))
	} else {
		return fmt.Sprintf("%.1fTB", float64(size)/float64(1024*1024*1024*1024))
	}
}

type logconfig struct {
	Filename string `json:"filename"`
	Level    int    `json:"level"`
	MaxLines int    `json:"maxlines"`
	MaxSize  int    `json:"maxsize"`
	Daily    bool   `json:"daily"`
	MaxDays  int    `json:"maxdays"`
	Color    bool   `json:"color"`
}

var logCfg = logconfig{Filename: os.Args[0], Level: 7, Daily: true, MaxDays: 30, Color: true}

func LogInit() error {
	logCfg.Filename = fmt.Sprintf("%s%c%s", LogDirGet(), os.PathSeparator, "runlog.log")
	value, err := json.Marshal(&logCfg)
	if err != nil {
		return err
	}

	if DebugFlag() {
		err = logs.SetLogger(logs.AdapterConsole)
	} else {
		err = logs.SetLogger(logs.AdapterFile, string(value))
	}

	if err != nil {
		return err
	}

	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	return nil
}

func WriteFull(w io.Writer, body []byte) error {
	begin := 0
	for {
		cnt, err := w.Write(body[begin:])
		if cnt > 0 {
			begin += cnt
		}
		if begin >= len(body) {
			return err
		}
		if err != nil {
			return err
		}
	}
}
