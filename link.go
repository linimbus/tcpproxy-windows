package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego/logs"
)

type LinkChannel struct {
	key    string
	remote net.Conn
	local  net.Conn
}

type LinkInstance struct {
	sync.RWMutex
	sync.WaitGroup

	close    bool
	address  string
	config   LinkConfig
	server   *tls.Config
	client   *tls.Config
	flow     int64
	listen   net.Listener
	channels map[string]*LinkChannel
}

func NewLinkInstance(config LinkConfig) (*LinkInstance, error) {
	address := fmt.Sprintf("%s:%d", config.Address, config.Port)

	listen, err := net.Listen(config.Protocol, address)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}

	link := new(LinkInstance)
	link.address = address
	link.listen = listen
	link.channels = make(map[string]*LinkChannel, 128)
	link.config = config

	if config.Tls != "NULL" {
		link.server, err = TlsConfigServer(config.Address, config.Tls)
		if err != nil {
			logs.Error(err.Error())
			return nil, err
		}
	}

	if config.Backend.Tls != "NULL" {
		link.client, err = TlsConfigClient(config.Backend.Address, config.Address, config.Tls)
		if err != nil {
			logs.Error(err.Error())
			return nil, err
		}
	}

	link.Add(1)
	go link.start()

	return link, nil
}

func (l *LinkInstance) proxy(wg *sync.WaitGroup, local net.Conn) {
	var err error
	var remote net.Conn

	defer func() {
		wg.Done()
		local.Close()
		if remote != nil {
			remote.Close()
		}
	}()

	key := local.RemoteAddr().String()
	backend := l.config.Backend

	address := fmt.Sprintf("%s:%d", backend.Address, backend.Port)

	if backend.Timeout == 0 {
		remote, err = net.Dial(backend.Protocol, address)
	} else {
		timeout := time.Second * time.Duration(backend.Timeout)
		remote, err = net.DialTimeout(backend.Protocol, address, timeout)
	}

	if err != nil {
		logs.Error(err.Error())
		return
	}

	if l.client != nil {
		remote = tls.Client(remote, l.client)
	}

	if l.server != nil {
		local = tls.Server(local, l.server)
	}

	channel := new(LinkChannel)
	channel.remote = remote
	channel.local = local
	channel.key = key

	l.Lock()
	l.channels[key] = channel
	l.Unlock()

	wg2 := new(sync.WaitGroup)
	wg2.Add(2)
	go connect(wg2, local, remote, &l.flow)
	go connect(wg2, remote, local, &l.flow)
	wg2.Wait()

	l.Lock()
	delete(l.channels, key)
	l.Unlock()
}

func (l *LinkInstance) start() {
	defer l.Done()
	logs.Info("link instance %s start", l.address)

	wg := new(sync.WaitGroup)
	for {
		if l.close {
			break
		}
		conn, err := l.listen.Accept()
		if err != nil {
			logs.Error(err.Error())
			continue
		}
		wg.Add(1)
		go l.proxy(wg, conn)
	}
	wg.Wait()

	logs.Info("link instance %s shutdown", l.address)
}

func (l *LinkInstance) Close() {
	l.Lock()
	l.close = true
	l.listen.Close()
	for _, v := range l.channels {
		v.remote.Close()
	}
	l.Unlock()

	l.Wait()
	logs.Info("link instance %s close", l.address)
}

func (l *LinkInstance) Channels() int {
	l.RLock()
	defer l.RUnlock()
	return len(l.channels)
}

func (l *LinkInstance) Flows() int64 {
	l.RLock()
	defer l.RUnlock()

	return l.flow
}

func connect(wg *sync.WaitGroup, local net.Conn, remote net.Conn, flow *int64) {
	defer func() {
		wg.Done()
	}()

	var buf [8192]byte
	for {
		cnt, err1 := local.Read(buf[:])
		if cnt > 0 {
			err2 := WriteFull(remote, buf[:cnt])
			if err2 != nil {
				return
			}
			atomic.AddInt64(flow, int64(cnt))
		}
		if err1 != nil {
			return
		}
	}
}
