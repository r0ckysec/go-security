/**
 * @Description
 * @Author r0cky
 * @Date 2021/12/2 10:59
 **/
package http

import (
	"bufio"
	"crypto/tls"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/r0ckysec/go-security/misc"
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
	keepAlive  = 5 * time.Second
)

func newClient() *fasthttp.Client {
	return &fasthttp.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		},
		// 最大连接数 默认 fasthttp.DefaultMaxConnsPerHost
		MaxConnsPerHost: fasthttp.DefaultMaxConnsPerHost,
		// 在这个时间间隔后，空闲的 keep-alive 连接会被关闭。 默认值为 DefaultMaxIdleConnDuration 。
		MaxIdleConnDuration: fasthttp.DefaultMaxIdleConnDuration,
		//在此持续时间后关闭保持活动的连接。 默认无限
		MaxConnDuration: time.Minute * 1,
		//读取超时，默认无限长容易阻塞，不设置read超时，可能会造成连接复用失效
		ReadTimeout: keepAlive * 2,
		// 写超时时间
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
	host            string
	headers         cmap.ConcurrentMap[string, string]
	proxy           string
	timeout         time.Duration
	client          *fasthttp.Client
	redirects       int
	StatusCode      int
	RequestRaw      string
	ResponseRaw     string
	ResponseBody    []byte
	ResponseHeaders string
}

func NewRequest() *Request {
	return &Request{
		headers:   cmap.New[string](),
		timeout:   dialTimout * 6,
		client:    newClient(),
		redirects: 0,
	}
}

// SetHost 设置Host
func (req *Request) SetHost(host string) {
	req.host = host
}

// SetHeaders 设置请求头
func (req *Request) SetHeaders(key string, value string) {
	req.headers.Set(key, value)
}

// SetProxy 设置代理
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
		req.client = newClient()
	}
}

// SetTimeout 设置超时时间 默认 30秒
func (req *Request) SetTimeout(timeout int) {
	req.timeout = time.Duration(timeout) * time.Second
}

// SetRedirects 设置跳转标识
func (req *Request) SetRedirects(r int) {
	req.redirects = r
}

// DisablePathNormalizing 设置请求路径标准化
func (req *Request) DisablePathNormalizing(b bool) {
	req.client.DisablePathNormalizing = b
}

// Request 请求方法主体
func (req *Request) Request(method string, url string, data string) (err error) {
	// 从请求池中分别获取一个request、response实例
	request, response := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	// 回收到请求池
	defer func() {
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	// 设置请求方法
	request.Header.SetMethod(method)
	// 设置请求URL
	request.SetRequestURI(url)
	// 设置URL标准化
	request.URI().DisablePathNormalizing = req.client.DisablePathNormalizing
	if req.headers.Count() > 0 {
		for iter := range req.headers.IterBuffered() {
			request.Header.Set(iter.Key, iter.Val)
		}
		ua, ok := req.headers.Get("User-Agent")
		if !ok || ua == "" {
			request.Header.SetUserAgent(RandUserAgent())
		}
		ct, ok := req.headers.Get("Content-Type")
		if (!ok || ct == "") && (strings.EqualFold(method, fasthttp.MethodPost) || strings.EqualFold(method, fasthttp.MethodPut)) {
			request.Header.SetContentType("application/x-www-form-urlencoded")
		}
	} else {
		request.Header.SetUserAgent(RandUserAgent())
		if len(data) > 0 {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	if len(data) > 0 {
		request.SetBodyString(data)
	}

	//修改HTTP超时时间
	if req.timeout != 0 {
		request.SetTimeout(req.timeout)
	}
	//修改HOST值
	if req.host != "" {
		request.SetHost(req.host)
	}
	request.SetConnectionClose()

	if req.redirects > 0 {
		if err = req.client.DoRedirects(request, response, req.redirects); err != nil {
			return
		}
	} else {
		if err = req.client.Do(request, response); err != nil {
			return
		}
	}
	req.ResponseHeaders = response.Header.String()
	req.ResponseBody = response.Body()
	req.RequestRaw = request.String()
	req.ResponseRaw = response.String()
	req.StatusCode = response.StatusCode()
	return
}

//func (req *Request) Get(Url string) ([]byte, error) {
//	body, _, err := req.Request("GET", Url, "")
//	return body, err
//}

//func (req *Request) Post(Url string, data string) ([]byte, error) {
//	body, _, err := req.Request("POST", Url, data)
//	return body, err
//}

func (req *Request) Get(url string) error {
	return req.Request(fasthttp.MethodGet, url, "")
}

func (req *Request) Post(url string, data string) error {
	return req.Request(fasthttp.MethodPost, url, data)
}

func RandUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(UserAgents))
	return UserAgents[i]
}

func (req *Request) RawRequest(raw string) (err error) {
	// 从请求池中分别获取一个request、response实例
	request, response := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	// 回收到请求池
	defer func() {
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	err = request.Read(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		return
	}
	// 设置URL标准化
	request.URI().DisablePathNormalizing = req.client.DisablePathNormalizing

	//修改HTTP超时时间
	if req.timeout != 0 {
		request.SetTimeout(req.timeout)
	}
	if req.redirects > 0 {
		if err = req.client.DoRedirects(request, response, req.redirects); err != nil {
			return
		}
	} else {
		if err = req.client.Do(request, response); err != nil {
			return
		}
	}
	req.ResponseHeaders = response.Header.String()
	req.RequestRaw = request.String()
	req.ResponseRaw = response.String()
	req.ResponseBody = response.Body()
	req.StatusCode = response.StatusCode()
	return
}

//func (req *Request) Raw(raw string) ([]byte, error) {
//	body, _, err := req.RequestRaw(raw)
//	return body, err
//}

func (req *Request) Raw(raw string) error {
	return req.RawRequest(raw)
}
