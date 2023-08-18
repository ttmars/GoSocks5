package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func main() {
	pu, err := url.Parse("socks5://127.0.0.1:7757")
	//pu, err := url.Parse("socks5://hello:world@127.0.0.1:7757")
	//pu, err := url.Parse("socks5://hello:world@39.101.203.25:8881")
	//pu, err := url.Parse("socks5://hello:world@47.245.90.149:7757")

	if err != nil {
		log.Fatal(err)
	}
	cli := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}
	//r, err := cli.Get("https://www.google.com.hk/")
	r, err := cli.Get("http://www.baidu.com/")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
