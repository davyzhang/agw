package agw

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"io"

	jsoniter "github.com/json-iterator/go"
)

// newRequest is a helper function to create a new request with a method and url.
// The request returned is a 'server' request as opposed to a 'client' one through
// simulated write onto the wire and read off of the wire.
// The differences between requests are detailed in the net/http package.
func newRequest(method, url string, headers map[string]string, bd io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, bd)
	if err != nil {
		panic(err)
	}
	// extract the escaped original host+path from url
	// http://localhost/path/here?v=1#frag -> //localhost/path/here
	opaque := ""
	if i := len(req.URL.Scheme); i > 0 {
		opaque = url[i+1:]
	}

	if i := strings.LastIndex(opaque, "?"); i > -1 {
		opaque = opaque[:i]
	}
	if i := strings.LastIndex(opaque, "#"); i > -1 {
		opaque = opaque[:i]
	}
	// Escaped host+path workaround as detailed in https://golang.org/pkg/net/url/#URL
	// for < 1.5 client side workaround
	req.URL.Opaque = opaque

	// Simulate writing to wire
	var buff bytes.Buffer
	req.Write(&buff)
	ioreader := bufio.NewReader(&buff)

	// Parse request off of 'wire'
	req, err = http.ReadRequest(ioreader)
	if err != nil {
		panic(err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	return req
}

//LPResponse mimic the behaviour of  http.ResponseWriter
type LPResponse struct {
	header http.Header
	body   interface{}
	status int
	base64 bool
}

func NewLPResponse() *LPResponse {
	var lpr LPResponse
	lpr.header = make(http.Header)
	lpr.status = http.StatusOK
	return &lpr
}

func (lpr *LPResponse) Header() http.Header { return lpr.header }
func (lpr *LPResponse) WriteHeader(s int)   { lpr.status = s }
func (lpr *LPResponse) Write(b []byte) (int, error) {
	return len(b), errors.New("Should not use Write() method on LPResponse, try WriteBody instead")
}
func (lpr *LPResponse) WriteBody(i interface{}, isBase64 bool) {
	lpr.body = i
	lpr.base64 = isBase64
}

type LPServer struct {
}

func (lps *LPServer) Process(req *http.Request, handler http.Handler) map[string]interface{} {
	resp := NewLPResponse()
	handler.ServeHTTP(resp, req)
	return resp.composeResp()
}

/*
{
    "isBase64Encoded": true|false,
    "statusCode": httpStatusCode,
    "headers": { "headerName": "headerValue", ... },
    "body": "..."
}
*/
func (lpr *LPResponse) composeResp() map[string]interface{} {
	// s int, h http.Header, body interface{}) map[string]interface{} {
	mh := make(map[string]string)
	for k, v := range lpr.header {
		mh[k] = strings.Join(v, ";")
	}

	var bd string
	switch t := lpr.body.(type) {
	case string:
		bd = t
	case []byte:
		bd = string(t)
	default:
		var err error
		bd, err = jsoniter.MarshalToString(lpr.body)
		if err != nil {
			panic(err)
		}
	}
	ret := map[string]interface{}{
		"statusCode":      lpr.status,
		"headers":         mh,
		"body":            bd,
		"isBase64Encoded": lpr.base64,
	}
	return ret
}

func Process(agp EventParser, h http.Handler) interface{} {
	buf := bytes.NewBuffer(agp.Body())
	req := newRequest(agp.Method(), agp.Url(), agp.Headers(), buf)
	return new(LPServer).Process(req, h)
}

func WriteResponse(w http.ResponseWriter, i interface{}, isBase64 bool) (int, error) {
	t := reflect.TypeOf(w).String()
	if t == "*http.response" { //dirty hack to get internal type
		switch i.(type) {
		case []byte:
			return w.Write(i.([]byte))
		case string:
			return w.Write([]byte(i.(string)))
		default:
			bs, err := jsoniter.Marshal(i)
			if err != nil {
				return 0, err
			}
			return w.Write(bs)
		}
	} else if t == "*agw.LPResponse" {
		w.(*LPResponse).WriteBody(i, isBase64)
	}
	return 0, nil
}
