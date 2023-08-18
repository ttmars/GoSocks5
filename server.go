package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"gitee.com/ttmasx/cob"
	"io"
	"log/slog"
	"net"
	"time"
)

var buf = make([]byte, 512)
var defaultUser string
var defaultPass string
var openAuth bool

func main() {
	//log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	level := cob.InitSlog()
	level.Set(slog.LevelInfo)
	level.Set(slog.LevelDebug)

	var port int
	flag.BoolVar(&openAuth, "a", false, "开启认证模式")
	flag.StringVar(&defaultUser, "u", "hello", "用户名")
	flag.StringVar(&defaultPass, "p", "world", "密码")
	flag.IntVar(&port, "P", 7757, "端口")
	flag.Parse()

	server, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info(fmt.Sprintf("listen port %v...", port))
	for {
		client, err := server.Accept()
		if err != nil {
			slog.Error(fmt.Sprintf("Accept failed: %v", err))
			continue
		}
		go process(client)
	}
}

func process(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic occurred", "error", err)
		}
	}()
	t1 := time.Now()
	// 第一次交互
	n, err := conn.Read(buf)
	if buf[0] != 5 || err != nil {
		slog.Error("read err", "error", err.Error())
		conn.Close()
		return
	}
	var reply []byte
	if openAuth {
		reply = []byte{0x05, 0x02}
	} else {
		reply = []byte{0x05, 0x00}
	}
	slog.Debug("一次交互", "read", buf[:n], "reply", reply)
	n, err = conn.Write(reply)
	if n != 2 || err != nil {
		slog.Error("write err", "error", err.Error())
		conn.Close()
		return
	}

	// 第二次交互
	var user, pass string
	if openAuth {
		n, err = conn.Read(buf)
		if buf[0] != 0x01 || err != nil {
			slog.Error("read err", "error", err.Error())
			conn.Close()
			return
		}
		reply = []byte{0x01, 0x00}
		slog.Debug("二次交互", "read", buf[:n], "reply", reply)
		userLen := buf[1]
		user = string(buf[2 : 2+userLen])
		passLen := buf[2+userLen]
		pass = string(buf[3+userLen : 3+userLen+passLen])
		if user != defaultUser || pass != defaultPass {
			slog.Error("认证错误", "用户", user, "密码", pass)
			conn.Close()
			return
		}
		// 认证
		n, err = conn.Write(reply)
		if err != nil || n != 2 {
			slog.Error("write err", "error", err.Error())
			conn.Close()
			return
		}
	}

	// 第三次交互
	var address string
	var port uint16
	n, err = conn.Read(buf)
	if err != nil {
		slog.Error("read err", "error", err.Error())
		conn.Close()
		return
	}
	reply = []byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0} // 这里回复？
	slog.Debug("三次交互", "read", buf[:n], "reply", reply)
	switch buf[3] {
	case 0x01:
		//ipv4
		slog.Error("unsupported ipv4")
		conn.Close()
		return
	case 0x03:
		//domain
		address = string(buf[5 : 5+buf[4]])
		port = binary.BigEndian.Uint16(buf[n-2 : n])
	case 0x04:
		// ipv6
		slog.Error("unsupported ipv6")
		conn.Close()
		return
	default:
		slog.Error("read err", "error", err.Error())
		conn.Close()
		return
	}
	host := fmt.Sprintf("%v:%v", address, port)
	dstConn, err := net.Dial("tcp", host)
	if err != nil {
		slog.Error("拨号错误", "error", err.Error(), "host", host)
		conn.Close()
		return
	}
	_, err = conn.Write(reply)
	if err != nil {
		slog.Error("write err", "error", err.Error())
		conn.Close()
		dstConn.Close()
		return
	}
	t2 := time.Now()

	// 代理流量
	dstConn.SetReadDeadline(time.Now().Add(time.Second * 3)) // 有必要？
	Socks5Forward(conn, dstConn)
	t3 := time.Now()
	slog.Info("统计", "host", host, "用户名", user, "密码", pass, "认证耗时", t2.Sub(t1), "转发耗时", t3.Sub(t2))
}

func Socks5Forward(client, target net.Conn) {
	defer client.Close()
	defer target.Close()
	forward := func(src, dest net.Conn, ch chan bool) {
		io.Copy(src, dest)
		ch <- true
		slog.Debug("io.copy done", "dest.RemoteAddr", dest.RemoteAddr().String())
	}
	ch := make(chan bool)
	go forward(client, target, ch)
	go forward(target, client, ch)
	// 等待两个协程都执行结束
	for i := 0; i < 2; i++ {
		<-ch
	}
}

//func Socks5Forward(client, target net.Conn) {
//	defer client.Close()
//	defer target.Close()
//	go io.Copy(client, target)
//	io.Copy(target, client)
//}

//func Socks5Forward(client, target net.Conn) {
//	defer client.Close()
//	defer target.Close()
//	go io.Copy(target, client)
//	io.Copy(client, target) // 阻塞时间长
//}
