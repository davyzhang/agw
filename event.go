package agw

import (
	"net/url"

	"github.com/json-iterator/go"
)

type EventParser interface {
	BodyString() string
	Body() []byte
	Path() string
	Method() string
	Url() string
	StageVariables() map[string]string
	Headers() map[string]string
}

type APIGateParser struct {
	content []byte
}

func NewAPIGateParser(c []byte) *APIGateParser {
	return &APIGateParser{
		c,
	}
}

func (agp *APIGateParser) BodyString() string {
	return jsoniter.Get(agp.content, "body").ToString()
}

func (agp *APIGateParser) Body() []byte {
	return []byte(agp.BodyString())
}
func (agp *APIGateParser) Path() string {
	return jsoniter.Get(agp.content, "path").ToString()
}
func (agp *APIGateParser) QueryStringParameters() url.Values {
	qp := map[string]interface{}{}
	q := jsoniter.Get(agp.content, "queryStringParameters")
	if q.ValueType() != jsoniter.NilValue {
		q.ToVal(&qp)
	}
	re := make(url.Values, len(qp))
	for k, v := range qp {
		re.Add(k, v.(string))
	}
	return re
}
func (agp *APIGateParser) Method() string {
	return jsoniter.Get(agp.content, "httpMethod").ToString()
}

func (agp *APIGateParser) Url() string {
	qs := agp.QueryStringParameters()
	if len(qs) <= 0 {
		return agp.Path()
	}
	return agp.Path() + "?" + qs.Encode()
}

func (agp *APIGateParser) StageVariables() map[string]string {
	re := map[string]string{}
	sv := jsoniter.Get(agp.content, "stageVariables")
	if sv.ValueType() != jsoniter.NilValue {
		sv.ToVal(&re)
	}
	return re
}

func (agp *APIGateParser) Headers() map[string]string {
	re := map[string]string{}
	sv := jsoniter.Get(agp.content, "headers")
	if sv.ValueType() != jsoniter.NilValue {
		sv.ToVal(&re)
	}
	return re
}
