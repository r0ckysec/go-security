package http

import (
	"fmt"
	"net/url"
	"testing"
)

func TestUrlParse(t *testing.T) {
	parse, err := url.Parse("127.0.0.1:65432")
	if err != nil {
		return
	}
	fmt.Println(parse)
}

func TestGet(t *testing.T) {
	req := NewRequest()
	err := req.Get("http://baidu.com")
	if err != nil {
		return
	}
	fmt.Println(req.RequestRaw, req.ResponseRaw, req.ResponseBody, req.ResponseHeaders)
}
