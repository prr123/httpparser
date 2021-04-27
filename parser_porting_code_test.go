package httparser

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type message struct {
	name  string
	hType ReqOrRsp
	raw   string
	//httpMethod      method
	statusCode     int
	responseStatus string
	requestPath    string
	requestUrl     string
	body           string
	bodySize       string
	host           string
	userinfo       string
	port           uint16
	//enum { NONE=0, FIELD, VALUE } last_header_element; //TODO最近移值
	headers         [][2]string
	shouldKeepAlive bool

	numChunks         int
	numChunksComplete int
	chunkLengths      []int

	upgrade string // upgraded body

	httpMajor     uint16
	httpMinor     uint16
	contentLength uint64

	messageBeginCbCalled    bool
	headersCompleteCbCalled bool
	messageCompleteCbCalled bool
	statusCbCalled          bool
	messageCompleteOnEof    bool
	bodyIsFinal             int
	allowChunkedLength      bool
}

func (m *message) eq(t *testing.T, m2 *message) bool {
	b := assert.Equal(t, m.headers, m2.headers)
	if !b {
		return false
	}

	b = assert.Equal(t, m.httpMajor, m2.httpMajor, "major")
	if !b {
		return false
	}
	b = assert.Equal(t, m.httpMinor, m2.httpMinor, "minor")
	if !b {
		return false
	}
	b = assert.Equal(t, m.hType, m2.hType, "htype")
	if !b {
		return false
	}

	b = assert.Equal(t, m.requestUrl, m2.requestUrl, "request url")
	if !b {
		return false
	}

	b = assert.Equal(t, m.body, m2.body, "body")
	if !b {
		return false
	}

	return true
}

var requests = []message{
	{
		name:  "curl get",
		hType: REQUEST,
		raw: "GET /test HTTP/1.1\r\n" +
			"User-Agent: curl/7.18.0 (i486-pc-linux-gnu) libcurl/7.18.0 OpenSSL/0.9.8g zlib/1.2.3.3 libidn/1.1\r\n" +
			"Host: 0.0.0.0=5000\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_GET,
		requestUrl:    "/test",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"User-Agent", "curl/7.18.0 (i486-pc-linux-gnu) libcurl/7.18.0 OpenSSL/0.9.8g zlib/1.2.3.3 libidn/1.1"},
			{"Host", "0.0.0.0=5000"},
			{"Accept", "*/*"},
		},
	},
	{
		name:  "firefox get",
		hType: REQUEST,
		raw: "GET /favicon.ico HTTP/1.1\r\n" +
			"Host: 0.0.0.0=5000\r\n" +
			"User-Agent: Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9) Gecko/2008061015 Firefox/3.0\r\n" +
			"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\n" +
			"Accept-Language: en-us,en;q=0.5\r\n" +
			"Accept-Encoding: gzip,deflate\r\n" +
			"Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7\r\n" +
			"Keep-Alive: 300\r\n" +
			"Connection: keep-alive\r\n" +
			"\r\n",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_GET,
		requestUrl:    "/favicon.ico",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"Host", "0.0.0.0=5000"},
			{"User-Agent", "Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9) Gecko/2008061015 Firefox/3.0"},
			{"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
			{"Accept-Language", "en-us,en;q=0.5"},
			{"Accept-Encoding", "gzip,deflate"},
			{"Accept-Charset", "ISO-8859-1,utf-8;q=0.7,*;q=0.7"},
			{"Keep-Alive", "300"},
			{"Connection", "keep-alive"},
		},
	},
	{
		name:  "dumbluck",
		hType: REQUEST,
		raw: "GET /dumbluck HTTP/1.1\r\n" +
			"aaaaaaaaaaaaa:++++++++++\r\n" +
			"\r\n",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_GET,
		requestUrl:    "/dumbluck",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"aaaaaaaaaaaaa", "++++++++++"},
		},
	},
	{
		name:  "fragment in url",
		hType: REQUEST,
		raw: "GET /forums/1/topics/2375?page=1#posts-17408 HTTP/1.1\r\n" +
			"\r\n",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_GET,
		requestUrl:    "/forums/1/topics/2375?page=1#posts-17408",
		contentLength: math.MaxUint64,
	},
	{
		name:  "get no headers no body",
		hType: REQUEST,
		raw: "GET /get_no_headers_no_body/world HTTP/1.1\r\n" +
			"\r\n",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_GET,
		requestUrl:    "/get_no_headers_no_body/world",
		contentLength: math.MaxUint64,
	},
	{
		name:  "get one header no body",
		hType: REQUEST,
		raw: "GET /get_one_header_no_body HTTP/1.1\r\n" +
			"Accept: */*\r\n" +
			"\r\n",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_GET,
		requestUrl:    "/get_one_header_no_body",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"Accept", "*/*"},
		},
	},
	{
		name:  "get funky content length body hello",
		hType: REQUEST,
		raw: "GET /get_funky_content_length_body_hello HTTP/1.0\r\n" +
			"conTENT-Length: 5\r\n" +
			"\r\n" +
			"HELLO",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            0,
		//method: HTTP_GET,
		requestUrl:    "/get_funky_content_length_body_hello",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"conTENT-Length", "5"},
		},
		body: "HELLO",
	},
	{
		name:  "post identity body world",
		hType: REQUEST,
		raw: "POST /post_identity_body_world?q=search#hey HTTP/1.1\r\n" +
			"Accept: */*\r\n" +
			"Content-Length: 5\r\n" +
			"\r\n" +
			"World",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_POST,
		requestUrl:    "/post_identity_body_world?q=search#hey",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"Accept", "*/*"},
			{"Content-Length", "5"},
		},
		body: "World",
	},
	{
		name:  "post - chunked body: all your base are belong to us",
		hType: REQUEST,
		raw: "POST /post_chunked_all_your_base HTTP/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"\r\n" +
			"1e\r\nall your base are belong to us\r\n" +
			"0\r\n" +
			"\r\n",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_POST,
		requestUrl:    "/post_chunked_all_your_base",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"Transfer-Encoding", "chunked"},
		},
		body: "all your base are belong to us",
	},
	{
		name:  "two chunks ; triple zero ending",
		hType: REQUEST,
		raw: "POST /two_chunks_mult_zero_end HTTP/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"\r\n" +
			"5\r\nhello\r\n" +
			"6\r\n world\r\n" +
			"000\r\n" +
			"\r\n",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_POST,
		requestUrl:    "/two_chunks_mult_zero_end",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"Transfer-Encoding", "chunked"},
		},
		body: "hello world",
	},
	{
		name:  "chunked with trailing headers. blech.",
		hType: REQUEST,
		raw: "POST /chunked_w_trailing_headers HTTP/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"\r\n" +
			"5\r\nhello\r\n" +
			"6\r\n world\r\n" +
			"0\r\n" +
			"Vary: *\r\n" +
			"Content-Type: text/plain\r\n" +
			"\r\n",

		shouldKeepAlive:      true,
		messageCompleteOnEof: false,
		httpMajor:            1,
		httpMinor:            1,
		//method: HTTP_POST,
		requestUrl:    "/chunked_w_trailing_headers",
		contentLength: math.MaxUint64,
		headers: [][2]string{
			{"Transfer-Encoding", "chunked"},
			{"Vary", "*"},
			{"Content-Type", "text/plain"},
		},
		body: "hello world",
	},
}

var settingTest Setting = Setting{
	MessageBegin: func(p *Parser) {
		m := p.GetUserData().(*message)
		m.messageBeginCbCalled = true
		m.hType = p.hType
	},
	URL: func(p *Parser, url []byte) {
		m := p.GetUserData().(*message)
		m.requestUrl += string(url)
	},
	Status: func(p *Parser, status []byte) {
		m := p.GetUserData().(*message)
		m.responseStatus = string(status)
	},
	HeaderField: func(p *Parser, headerField []byte) {
		m := p.GetUserData().(*message)
		m.headers = append(m.headers, [2]string{string(headerField), ""})
	},
	HeaderValue: func(p *Parser, headerValue []byte) {
		m := p.GetUserData().(*message)
		m.headers[len(m.headers)-1][1] = string(headerValue)
	},
	HeadersComplete: func(p *Parser) {
		m := p.GetUserData().(*message)
		m.headersCompleteCbCalled = true
	},
	Body: func(p *Parser, body []byte) {
		m := p.GetUserData().(*message)
		m.body += string(body)
	},
	MessageComplete: func(p *Parser) {
		m := p.GetUserData().(*message)
		m.messageCompleteCbCalled = true
		m.httpMajor = uint16(p.Major)
		m.httpMinor = uint16(p.Minor)

	},
}

func parse(p *Parser, data string) (int, error) {
	return p.Execute(&settingTest, []byte(data))
}

func test_Message(t *testing.T, m *message) {
	for msg1len := 0; msg1len < len(m.raw); msg1len++ {
		p := New(m.hType)
		got := &message{}
		p.SetUserData(got)

		msg1Message := m.raw[:msg1len]
		msg2Message := m.raw[msg1len:]

		var (
			n1   int
			err1 error
		)
		if msg1len > 0 {
			n1, err1 = parse(p, msg1Message)
			assert.NoError(t, err1)
			msg1Message = msg1Message[n1:]
		}

		_, err := parse(p, msg1Message+msg2Message)
		assert.NoError(t, err)
		if b := m.eq(t, got); !b {
			t.Logf("msg1.len:%d, msg2.len:%d, test case name:%s\n", len(msg1Message), len(msg2Message), m.name)
			t.Logf("msg1len:%d, msg1(%s)", msg1len, msg1Message)
			t.Logf("msg2(%s)", msg2Message)
			break
		}

	}
}

func Test_Message(t *testing.T) {
	for _, req := range requests {
		test_Message(t, &req)
		_ = req
	}
}
