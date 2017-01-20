package agw

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"strings"
	"util"

	"io"
)

func composeURL(lpe *LambdaProxyEvent) string {
	qs := lpe.QueryStringParameters()
	return lpe.Path() + "?" + qs.Encode()
}

func bodyReader(lpe *LambdaProxyEvent) (io.Reader, error) {
	bd, err := lpe.Body()
	if bd == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s, err := bd.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(s), nil
}

func ToHTTPRequest(lpe *LambdaProxyEvent) (*http.Request, error) {
	bd, err := bodyReader(lpe)
	if err != nil {
		return nil, err
	}
	return newRequest(lpe.HTTPMethod(), composeURL(lpe), bd), nil
}

// newRequest is a helper function to create a new request with a method and url.
// The request returned is a 'server' request as opposed to a 'client' one through
// simulated write onto the wire and read off of the wire.
// The differences between requests are detailed in the net/http package.
func newRequest(method, url string, bd io.Reader) *http.Request {
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

	return req
}

type LPResponse struct {
	header http.Header
	body   interface{}
	status int
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
	return len(b), errors.New("Should not use this method on LPResponse, try WriteBody instead")
}
func (lpr *LPResponse) WriteBody(i interface{}) { lpr.body = i }

type LPServer struct {
}

func (lps *LPServer) Process(req *http.Request, handler http.Handler) map[string]interface{} {
	resp := NewLPResponse()
	handler.ServeHTTP(resp, req)
	return composeResp(resp)
}

func composeResp(resp *LPResponse) map[string]interface{} {
	// s int, h http.Header, body interface{}) map[string]interface{} {
	mh := make(map[string]string)
	for k, v := range resp.header {
		mh[k] = strings.Join(v, ";")
	}

	var bd string
	switch t := resp.body.(type) {
	case string:
		bd = t
	default:
		bd = util.MustJSON(resp.body)
	}
	ret := map[string]interface{}{
		"statusCode": resp.status,
		"headers":    mh,
		"body":       bd,
	}
	return ret
}

var defaultServer = new(LPServer)

func Process(b []byte, h http.Handler) (map[string]interface{}, error) {
	lpe, err := NewLambdaProxyEvent(b)
	if err != nil {
		return nil, err
	}
	return ProcessEvent(lpe, h)
}

func ProcessEvent(lpe *LambdaProxyEvent, h http.Handler) (map[string]interface{}, error) {
	req, err := ToHTTPRequest(lpe)
	if err != nil {
		return nil, err
	}

	return defaultServer.Process(req, h), nil
}
