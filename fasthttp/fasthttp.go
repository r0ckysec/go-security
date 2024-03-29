/**
 * @Description
 * @Author r0cky
 * @Date 2021/12/2 10:59
 **/
package http

import (
	"bufio"
	"crypto/tls"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/r0ckysec/go-security/bin/misc"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

var (
	UserAgents = []string{
		//"Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
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
	dialTimout = 5 * time.Second
	keepAlive  = 15 * time.Second
	client     = NewClient()
)

func NewClient() *fasthttp.Client {
	return &fasthttp.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		},
		// 最大连接数 默认 fasthttp.DefaultMaxConnsPerHost
		MaxConnsPerHost: 2000,
		// 在这个时间间隔后，空闲的 keep-alive 连接会被关闭。 默认值为 DefaultMaxIdleConnDuration 。
		MaxIdleConnDuration: fasthttp.DefaultMaxIdleConnDuration,
		//在此持续时间后关闭保持活动的连接。 默认无限
		MaxConnDuration: time.Minute * 2,
		//读取超时，默认无限长容易阻塞
		ReadTimeout:  keepAlive * 2,
		WriteTimeout: keepAlive * 2,
		//等待空闲连接的最长持续时间，默认不会等待，立即返回 ErrNoFreeConns
		MaxConnWaitTimeout:        dialTimout,
		MaxIdemponentCallAttempts: fasthttp.DefaultMaxIdemponentCallAttempts,
		RetryIf: func(request *fasthttp.Request) bool {
			retry := request.Header.Peek("retry")
			if misc.Bytes2Str(retry) == "1" {
				return true
			}
			return false
		},
	}
}

type Request struct {
	host      string
	headers   cmap.ConcurrentMap
	proxy     string
	timeout   time.Duration
	client    *fasthttp.Client
	redirects int
}

func NewRequest() *Request {
	return &Request{
		headers:   cmap.New(),
		timeout:   dialTimout * 4,
		client:    client,
		redirects: 0,
	}
}

func (req *Request) SetHost(host string) {
	req.host = host
}

func (req *Request) SetHeaders(key string, value string) {
	req.headers.Set(key, value)
}
func (req *Request) SetProxy(proxy string) {
	req.proxy = proxy
	//修改代理选项
	if req.proxy != "" {
		u, err := url.Parse(req.proxy)
		if err == nil {
			if strings.Contains(strings.ToLower(u.Scheme), "http") {
				req.client.Dial = fasthttpproxy.FasthttpHTTPDialer(u.Host)
			} else {
				req.client.Dial = fasthttpproxy.FasthttpSocksDialer(u.Host)
			}
		} else {
			req.client.Dial = fasthttpproxy.FasthttpSocksDialer(req.proxy)
		}
		//fasthttpproxy.FasthttpSocksDialer("127.0.0.1:65432")
		//fasthttpproxy.FasthttpHTTPDialer("http://127.0.0.1:65432")
	} else {
		req.client = nil
		req.client = NewClient()
	}
}
func (req *Request) SetTimeout(timeout int) {
	req.timeout = time.Duration(timeout) * time.Second
}

func (req *Request) SetRedirects(r int) {
	req.redirects = r
}

func (req *Request) DisablePathNormalizing(b bool) {
	req.client.DisablePathNormalizing = b
}

func (req *Request) Request(method string, Url string, data string) ([]byte, *fasthttp.ResponseHeader, error) {
	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request) // 用完需要释放资源
	request.Header.SetMethod(strings.ToUpper(method))
	request.SetRequestURI(Url)
	request.URI().DisablePathNormalizing = req.client.DisablePathNormalizing
	if len(req.headers) > 0 {
		for iter := range req.headers.IterBuffered() {
			request.Header.Set(iter.Key, iter.Val.(string))
		}
		ua, ok := req.headers.Get("User-Agent")
		if !ok || ua.(string) == "" {
			request.Header.Set("User-Agent", getUserAgent())
		}
		ct, ok := req.headers.Get("Content-Type")
		if (!ok || ct.(string) == "") && len(data) > 0 {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	} else {
		request.Header.Set("User-Agent", getUserAgent())
		if len(data) > 0 {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	if len(data) > 0 {
		request.SetBodyString(data)
	}

	//修改HTTP超时时间
	//if req.timeout != 0 {
	//	timeout = time.Duration(req.timeout) * time.Second
	//}
	//修改HOST值
	if req.host != "" {
		request.SetHost(req.host)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源, 一定要释放
	request.SetConnectionClose()
	var err error
	if req.redirects > 0 {
		if err = req.client.DoRedirects(request, resp, req.redirects); err != nil {
			return nil, nil, err
		}
	} else {
		if err = req.client.DoTimeout(request, resp, req.timeout); err != nil {
			return nil, nil, err
		}
	}
	header := &fasthttp.ResponseHeader{}
	resp.Header.CopyTo(header)
	return resp.Body(), header, err
}

func (req *Request) Get(Url string) ([]byte, error) {
	body, _, err := req.Request("GET", Url, "")
	return body, err
}

func (req *Request) Post(Url string, data string) ([]byte, error) {
	body, _, err := req.Request("POST", Url, data)
	return body, err
}

func (req *Request) GetH(Url string) ([]byte, *fasthttp.ResponseHeader, error) {
	return req.Request("GET", Url, "")
}

func (req *Request) PostH(Url string, data string) ([]byte, *fasthttp.ResponseHeader, error) {
	return req.Request("GET", Url, "")
}

func getUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(UserAgents))
	return UserAgents[i]
}

func (req *Request) RequestRaw(raw string) ([]byte, *fasthttp.ResponseHeader, error) {
	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request) // 用完需要释放资源
	err := request.Read(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		return nil, nil, err
	}
	request.URI().DisablePathNormalizing = req.client.DisablePathNormalizing
	if len(req.headers) > 0 {
		for iter := range req.headers.IterBuffered() {
			request.Header.Set(iter.Key, iter.Val.(string))
		}
		ua, ok := req.headers.Get("User-Agent")
		if !ok || ua.(string) == "" {
			request.Header.Set("User-Agent", getUserAgent())
		}
	} else {
		request.Header.Set("User-Agent", getUserAgent())
	}

	//修改HTTP超时时间
	//if req.timeout != 0 {
	//	timeout = time.Duration(req.timeout) * time.Second
	//}
	//修改HOST值
	if req.host != "" {
		request.SetHost(req.host)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源, 一定要释放
	request.SetConnectionClose()
	if req.redirects > 0 {
		if err = req.client.DoRedirects(request, resp, req.redirects); err != nil {
			return nil, nil, err
		}
	} else {
		if err = req.client.DoTimeout(request, resp, req.timeout); err != nil {
			return nil, nil, err
		}
	}
	header := &fasthttp.ResponseHeader{}
	resp.Header.CopyTo(header)
	return resp.Body(), header, err
}

func (req *Request) Raw(raw string) ([]byte, error) {
	body, _, err := req.RequestRaw(raw)
	return body, err
}

func (req *Request) RawH(raw string) ([]byte, *fasthttp.ResponseHeader, error) {
	return req.RequestRaw(raw)
}

func (req *Request) HTTPRaw(method string, Url string, data string) ([]byte, *fasthttp.ResponseHeader, string, string, error) {
	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request) // 用完需要释放资源
	request.Header.SetMethod(strings.ToUpper(method))
	request.SetRequestURI(Url)
	request.URI().DisablePathNormalizing = req.client.DisablePathNormalizing
	if len(req.headers) > 0 {
		for iter := range req.headers.IterBuffered() {
			request.Header.Set(iter.Key, iter.Val.(string))
		}
		ua, ok := req.headers.Get("User-Agent")
		if !ok || ua.(string) == "" {
			request.Header.Set("User-Agent", getUserAgent())
		}
		ct, ok := req.headers.Get("Content-Type")
		if (!ok || ct.(string) == "") && len(data) > 0 {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	} else {
		request.Header.Set("User-Agent", getUserAgent())
		if len(data) > 0 {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	if len(data) > 0 {
		request.SetBodyString(data)
	}
	//修改HOST值
	if len(request.Header.Host()) > 0 {
		request.UseHostHeader = true
	} else {
		request.UseHostHeader = false
	}
	if req.host != "" {
		request.SetHost(req.host)
	}
	c, ok := req.headers.Get("Connection")
	if !ok || c.(string) == "" {
		request.SetConnectionClose()
	}
	requestRaw := request.String()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源, 一定要释放
	var err error
	if req.redirects > 0 {
		if err = req.client.DoRedirects(request, resp, req.redirects); err != nil {
			return nil, nil, requestRaw, resp.String(), err
		}
	} else {
		if err = req.client.DoTimeout(request, resp, req.timeout); err != nil {
			return nil, nil, requestRaw, resp.String(), err
		}
	}
	header := &fasthttp.ResponseHeader{}
	resp.Header.CopyTo(header)
	return resp.Body(), header, request.String(), resp.String(), err
}

func GetHeaderKeys(header *fasthttp.ResponseHeader) []string {
	var keys []string
	split := strings.Split(header.String(), "\n")
	for _, hn := range split {
		n := strings.SplitN(hn, ":", 2)
		if len(n) > 1 {
			keys = append(keys, n[0])
		}
	}
	keys = misc.RemoveDuplicatesAndEmpty(keys)
	return keys
}
