package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func main() {
	//pu, err := url.Parse("socks5://127.0.0.1:1080")
	pu, err := url.Parse("socks5://hello:world@127.0.0.1:7757")

	if err != nil {
		log.Fatal(err)
	}
	cli := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}
	r, err := cli.Get("http://blog.youthsweet.com/weibo")
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
