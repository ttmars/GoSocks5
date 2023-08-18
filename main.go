package main

import (
	"flag"
	"gosocks5/pkg"
	"log/slog"
)

func main() {
	//run1()
	run2()
}

func run1() {
	var Port int
	var DialTimeout int
	var LogLevel int
	var OpenAuth bool
	var SaveLogFile bool
	var User string
	var Pass string
	flag.IntVar(&Port, "P", 8890, "端口")
	flag.IntVar(&DialTimeout, "t", 10, "拨号超时")
	flag.IntVar(&LogLevel, "e", -4, "日志等级,-4/0/4/8")
	flag.BoolVar(&OpenAuth, "a", false, "是否开启认证")
	flag.BoolVar(&SaveLogFile, "l", false, "是否保存日志文件")
	flag.StringVar(&User, "u", "hello", "用户名")
	flag.StringVar(&Pass, "p", "world", "密码")
	flag.Parse()
	s := pkg.NewSocks5(Port, OpenAuth, User, Pass, slog.Level(LogLevel), SaveLogFile, DialTimeout)
	s.Start()
}

func run2() {
	pkg.DefaultSocks5.Start()
}
