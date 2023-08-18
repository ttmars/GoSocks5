package pkg

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"time"
)

type Socks5 struct {
	Port        int        // 监听端口
	OpenAuth    bool       // 是否开启认证
	User        string     // 用户名
	Pass        string     // 密码
	LogLevel    slog.Level // 日志等级
	SaveLogFile bool       // 是否保存到日志文件
	DialTimeout int        // 拨号超时,秒
}

func NewSocks5(port int, openAuth bool, user string, pass string, logLevel slog.Level, saveLogFile bool, dialTimeout int) *Socks5 {
	return &Socks5{
		Port:        port,
		OpenAuth:    openAuth,
		User:        user,
		Pass:        pass,
		LogLevel:    logLevel,
		SaveLogFile: saveLogFile,
		DialTimeout: dialTimeout,
	}
}

var DefaultSocks5 = &Socks5{
	Port:        23333,
	OpenAuth:    false,
	User:        "hello",
	Pass:        "world",
	LogLevel:    10,
	SaveLogFile: false,
	DialTimeout: 10,
}

func (s *Socks5) Start() {
	level := InitSlog(s.SaveLogFile)
	level.Set(s.LogLevel)
	server, err := net.Listen("tcp", fmt.Sprintf(":%v", s.Port))
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("server start", "parameters", *s)
	for {
		client, err := server.Accept()
		if err != nil {
			slog.Error("accept fail", "error", err.Error())
			continue
		}
		go s.process(client)
	}
}

func (s *Socks5) process(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic", "error", err)
		}
	}()
	var buf = make([]byte, 512) // ???
	var reply []byte
	var err error
	var n int

	// first dialogue
	n, err = conn.Read(buf)
	if err != nil || buf[0] != 5 {
		slog.Error("read fail", "error", err, "s.Buf[0]", buf[0])
		conn.Close()
		return
	}
	if s.OpenAuth {
		reply = []byte{0x05, 0x02}
	} else {
		reply = []byte{0x05, 0x00}
	}
	slog.Debug("first dialogue", "read", buf[:n], "reply", reply)
	n, err = conn.Write(reply)
	if err != nil || n != 2 {
		slog.Error("write fail", "error", err, "n", n)
		conn.Close()
		return
	}

	// second dialogue
	var user, pass string
	if s.OpenAuth {
		n, err = conn.Read(buf)
		if err != nil || buf[0] != 0x01 {
			slog.Error("read fail", "error", err, "buf[0]", buf[0])
			conn.Close()
			return
		}
		reply = []byte{0x01, 0x00}
		slog.Debug("second dialogue", "read", buf[:n], "reply", reply)
		userLen := buf[1]
		user = string(buf[2 : 2+userLen])
		passLen := buf[2+userLen]
		pass = string(buf[3+userLen : 3+userLen+passLen])
		if user != s.User || pass != s.Pass {
			slog.Error("auth fail", "user", user, "pass", pass)
			conn.Close()
			return
		}
		n, err = conn.Write(reply)
		if err != nil || n != 2 {
			slog.Error("write fail", "error", err, "n", n)
			conn.Close()
			return
		}
	}

	// thirdly dialogue
	var address string
	var port uint16
	n, err = conn.Read(buf)
	if err != nil {
		slog.Error("read fail", "error", err, "n", n)
		conn.Close()
		return
	}
	reply = []byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0} // ???
	slog.Debug("thirdly dialogue", "read", buf[:n], "reply", reply)
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
		slog.Error("read fail", "error", err, "buf[3]", buf[3])
		conn.Close()
		return
	}
	host := fmt.Sprintf("%v:%v", address, port)
	dstConn, err := net.DialTimeout("tcp", host, time.Second*time.Duration(s.DialTimeout))
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

	// forward
	//dstConn.SetReadDeadline(time.Now().Add(time.Second * 3)) 									// ???
	Socks5Forward(conn, dstConn)
	slog.Info("record", "host", host, "user", user, "pass", pass)
}
