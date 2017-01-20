package agw

import (
	"encoding/json"
	"net/url"

	simplejson "github.com/bitly/go-simplejson"
)

type LambdaProxyEvent struct {
	raw, body *simplejson.Json

	resource,
	httpMethod,
	path string
	queryStringParameters url.Values
	stageVariables,
	pathParameters,
	headers map[string]string
}

func NewLambdaProxyEvent(raw []byte) (*LambdaProxyEvent, error) {
	var lpe LambdaProxyEvent
	sj, err := simplejson.NewJson(raw)
	if err != nil {
		return nil, err
	}
	lpe.raw = sj
	return &lpe, nil
}

func (lpe *LambdaProxyEvent) StageVariables() map[string]string {
	if lpe.stageVariables == nil {
		lpe.stageVariables = toStringMap(lpe.raw.Get("stageVariables").MustMap())
	}
	return lpe.stageVariables
}

func (lpe *LambdaProxyEvent) Resource() string {
	if lpe.resource == "" {
		lpe.resource = lpe.raw.Get("resource").MustString()
	}
	return lpe.resource
}

func (lpe *LambdaProxyEvent) HTTPMethod() string {
	if lpe.httpMethod == "" {
		lpe.httpMethod = lpe.raw.Get("httpMethod").MustString()
	}
	return lpe.httpMethod
}

func (lpe *LambdaProxyEvent) QueryStringParameters() url.Values {
	if lpe.queryStringParameters == nil {
		qm := toStringMap(lpe.raw.Get("queryStringParameters").MustMap())
		lpe.queryStringParameters = make(url.Values)
		for k, v := range qm {
			lpe.queryStringParameters.Add(k, v)
		}
	}

	return lpe.queryStringParameters
}

func (lpe *LambdaProxyEvent) PathParameters() map[string]string {
	if lpe.pathParameters == nil {
		lpe.pathParameters = toStringMap(lpe.raw.Get("pathParameters").MustMap())
	}
	return lpe.pathParameters
}

func (lpe *LambdaProxyEvent) Path() string {
	if lpe.path == "" {
		lpe.path = lpe.raw.Get("path").MustString()
	}
	return lpe.path
}

func (lpe *LambdaProxyEvent) Headers() map[string]string {
	if lpe.headers == nil {
		lpe.headers = toStringMap(lpe.raw.Get("headers").MustMap())
	}
	return lpe.headers
}

func (lpe *LambdaProxyEvent) Body() (*simplejson.Json, error) {
	if lpe.body == nil {
		body := lpe.raw.Get("body").MustString()
		sj, err := simplejson.NewJson([]byte(body))
		if err != nil {
			return nil, err
		}
		lpe.body = sj
	}
	return lpe.body, nil
}

func (lpe *LambdaProxyEvent) JSONBody() (map[string]interface{}, error) {
	body, err := lpe.Body()
	if err != nil {
		return nil, err
	}
	return parseJSONType(body.MustMap()), nil
}

func (lpe *LambdaProxyEvent) Raw() *simplejson.Json {
	return lpe.raw
}

func parseJSONType(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		switch t := v.(type) {
		case json.Number:
			e, err := t.Float64()
			if err != nil {
				result[k] = 0
			} else {
				result[k] = e
			}
		default:
			result[k] = v
		}
	}
	return result
}

func toStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v.(string)
	}
	return result
}
