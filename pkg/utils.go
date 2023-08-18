package pkg

import (
	"io"
	"log/slog"
	"net"
	"os"
	"time"
)

func InitSlog(saveLogFile bool) *slog.LevelVar {
	// 动态设置日志输出等级
	var programLevel = new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug)

	// 设置日志输出
	var w io.Writer
	if saveLogFile {
		f, err := os.OpenFile(time.Now().Format("2006-01-02")+".log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		w = io.MultiWriter(os.Stdout, f)
	} else {
		w = os.Stdout
	}

	// 开启行号记录
	h := slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       programLevel,
		ReplaceAttr: nil,
	})

	// 设置为系统默认日志
	slog.SetDefault(slog.New(h))
	return programLevel
}

func Socks5Forward(client, target net.Conn) {
	defer client.Close()
	defer target.Close()
	forward := func(src, dest net.Conn, ch chan bool) {
		io.Copy(src, dest)
		ch <- true
	}
	ch := make(chan bool)
	go forward(client, target, ch)
	go forward(target, client, ch)
	<-ch
	<-ch
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
