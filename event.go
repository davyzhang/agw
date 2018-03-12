package agw

import (
	"net/url"

	"github.com/json-iterator/go"
)

type contentParser interface {
}

type apiGateParser struct {
	content []byte
}

func newAPIGateParser(c []byte) *apiGateParser {
	return &apiGateParser{
		c,
	}
}

func (agp *apiGateParser) body() []byte {
	return []byte(jsoniter.Get(agp.content, "body").ToString())
}
func (agp *apiGateParser) path() string {
	return jsoniter.Get(agp.content, "path").ToString()
}
func (agp *apiGateParser) queryStringParameters() url.Values {
	qp := map[string]interface{}{}
	jsoniter.Get(agp.content, "queryStringParameters").ToVal(&qp)
	re := make(url.Values, len(qp))
	for k, v := range qp {
		re.Add(k, v.(string))
	}
	return re
}
func (agp *apiGateParser) method() string {
	return jsoniter.Get(agp.content, "httpMethod").ToString()
}

func (agp *apiGateParser) url() string {
	qs := agp.queryStringParameters()
	if len(qs) <= 0 {
		return agp.path()
	}
	return agp.path() + "?" + qs.Encode()
}
