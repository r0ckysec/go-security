package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/r0ckysec/go-security/secio"
	"github.com/thinkeridea/go-extend/exstrings"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	dialTimout = 5 * time.Second
	keepAlive  = 15 * time.Second
	UserAgents = []string{
		"Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; AcooBrowser; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; Acoo Browser; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; .NET CLR 3.0.04506)",
		"Mozilla/4.0 (compatible; MSIE 7.0; AOL 9.5; AOLBuild 4337.35; Windows NT 5.1; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (Windows; U; MSIE 9.0; Windows NT 9.0; en-US)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.2; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 3.0.04506.30)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN) AppleWebKit/523.15 (KHTML, like Gecko, Safari/419.3) Arora/0.3 (Change: 287 c9dfb30)",
		"Mozilla/5.0 (X11; U; Linux; en-US) AppleWebKit/527+ (KHTML, like Gecko, Safari/419.3) Arora/0.6",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.2pre) Gecko/20070215 K-Ninja/2.1.1",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN; rv:1.9) Gecko/20080705 Firefox/3.0 Kapiko/3.0",
		"Mozilla/5.0 (X11; Linux i686; U;) Gecko/20070322 Kazehakase/0.4.5",
		"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.8) Gecko Fedora/1.9.0.8-1.fc10 Kazehakase/0.5.6",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.20 (KHTML, like Gecko) Chrome/19.0.1036.7 Safari/535.20",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; fr) Presto/2.9.168 Version/11.52",
	}
	dialer = net.Dialer{
		Timeout:   dialTimout,
		KeepAlive: keepAlive,
	}
	transport = http.Transport{
		DialContext:         dialer.DialContext,
		MaxConnsPerHost:     2000,
		MaxIdleConnsPerHost: 2000,
		MaxIdleConns:        2000,
		IdleConnTimeout:     keepAlive,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: dialTimout,
		DisableKeepAlives:   false,
	}
	requestPool sync.Pool
)

func AcquireRequest() *HttpRequest {
	v := requestPool.Get()
	if v == nil {
		return NewRequest()
	}
	return v.(*HttpRequest)
}

func ReleaseRequest(req *HttpRequest) {
	if req.cancelFunc != nil {
		req.cancelFunc()
	}
	req.Reset()
	requestPool.Put(req)
}

type HttpRequest struct {
	host       string
	headers    map[string]string
	proxy      string
	timeout    int
	transport  *http.Transport
	cancelFunc context.CancelFunc
}

func NewRequest() *HttpRequest {
	return &HttpRequest{
		timeout:   10,
		headers:   make(map[string]string, 0),
		transport: &transport,
	}
}

func (req *HttpRequest) SetHost(host string) {
	req.host = host
}

func (req *HttpRequest) SetHeaders(key string, value string) {
	req.headers[key] = value
}
func (req *HttpRequest) SetProxy(proxy string) {
	req.proxy = proxy
}
func (req *HttpRequest) SetTimeout(timeout int) {
	req.timeout = timeout
}

func (req *HttpRequest) Request(method string, Url string, data string) ([]byte, http.Header, error) {
	request, err := http.NewRequest(method, Url, strings.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	//ua := req.headers["User-Agent"]
	if len(req.headers) > 0 {
		for k, v := range req.headers {
			request.Header.Add(k, v)
		}
		if req.headers["User-Agent"] == "" {
			request.Header.Add("User-Agent", getUserAgent())
		}
		if req.headers["Content-Type"] == "" && len(data) > 0 {
			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
	} else {
		request.Header.Add("User-Agent", getUserAgent())
		if len(data) > 0 {
			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	request.Close = true

	client := http.Client{}
	//修改HTTP超时时间
	if req.timeout != 0 {
		client.Timeout = time.Duration(req.timeout) * time.Second
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(req.timeout)*time.Second)
	req.cancelFunc = cancelFunc
	request.WithContext(ctx)
	//修改HOST值
	if req.host != "" {
		request.Host = req.host
	}
	//修改代理选项
	if req.proxy != "" {
		uri, _ := url.Parse(req.proxy)
		req.transport.Proxy = http.ProxyURL(uri)
	}
	client.Transport = req.transport
	resp, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	if resp != nil {
		defer func() {
			//if resp.Body != nil && resp.Body != http.NoBody {
			_, _ = io.Copy(ioutil.Discard, resp.Body)
			_ = resp.Body.Close()
			//}
		}()
	}
	defer func() {
		if cancelFunc != nil {
			cancelFunc()
		}
	}()
	body, err := secio.ReadAll(resp.Body)
	return body, resp.Header, err
}

func (req *HttpRequest) Get(Url string) ([]byte, error) {
	body, _, err := req.Request("GET", Url, "")
	return body, err
}

func (req *HttpRequest) Post(Url string, data string) ([]byte, error) {
	body, _, err := req.Request("POST", Url, data)
	return body, err
}

func (req *HttpRequest) GetH(Url string) ([]byte, http.Header, error) {
	return req.Request("GET", Url, "")
}

func (req *HttpRequest) PostH(Url string, data string) ([]byte, http.Header, error) {
	return req.Request("POST", Url, data)
}

func (req *HttpRequest) Reset() {
	*req = HttpRequest{
		timeout:   10,
		headers:   map[string]string{},
		transport: &transport,
	}
}

func (req *HttpRequest) GetHeaders(header http.Header) string {
	result := &bytes.Buffer{}
	defer result.Reset()
	for i := range header {
		hs := header.Values(i)
		for _, h := range hs {
			result.Write(exstrings.Bytes(fmt.Sprintf("%s: %s\n", i, h)))
		}
	}
	return result.String()
}

func getUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(UserAgents))
	return UserAgents[i]
}
