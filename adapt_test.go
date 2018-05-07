package agw

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-zoo/bone"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var ev1 = `
{
    "resource": "/test1/{proxy+}",
    "path": "/test/hello",
    "httpMethod": "POST",
    "headers": {"Var":"hello"},
    "queryStringParameters": {
        "k1": "v1",
        "k2": "v2"
    },
    "pathParameters": {
        "proxy": "test"
    },
    "stageVariables": {
        "lbAlias": "current"
    },
    "requestContext": {
        "path": "/test1/{proxy+}",
        "accountId": "095615327118",
        "resourceId": "ybki7l",
        "stage": "test-invoke-stage",
        "requestId": "test-invoke-request",
        "identity": {
            "cognitoIdentityPoolId": null,
            "cognitoIdentityId": null,
            "apiKey": "test-invoke-api-key",
            "cognitoAuthenticationType": null,
            "userArn": "arn:aws:iam::095615327118:root",
            "apiKeyId": "test-invoke-api-key-id",
            "userAgent": "Apache-HttpClient/4.5.x (Java/1.8.0_144)",
            "accountId": "095615327118",
            "caller": "095615327118",
            "sourceIp": "test-invoke-source-ip",
            "accessKey": "ASIAJTPDCBBJQKRD3FMQ",
            "cognitoAuthenticationProvider": null,
            "user": "095615327118"
        },
        "resourcePath": "/test1/{proxy+}",
        "httpMethod": "POST",
        "apiId": "uorto7w779"
    },
    "body": "{\"test\":\"test body\"}",
    "isBase64Encoded": false
}
`[1:]

func testhandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("test", "test header")
	w.(*LPResponse).WriteBody(map[string]string{"test": "test body"}, false)
}

func TestGorillaMuxFlow(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/test/hello", testhandler)

	content := Process(NewAPIGateParser([]byte(ev1)), r).(map[string]interface{})

	type c struct {
		want, got interface{}
	}
	ts := map[string]c{
		"statusCode": c{http.StatusOK, content["statusCode"]},
		"body":       c{`{"test":"test body"}`, content["body"]},
		"headers":    c{map[string]string{"Test": "test header"}, content["headers"]}, //uppercase for header
	}
	for k, v := range ts {
		if !reflect.DeepEqual(v.got, v.want) {
			t.Errorf("%s check wrong, want %+v, got %+v", k, v.want, v.got)
		}
	}
}

func TestBoneFlow(t *testing.T) {
	r := bone.New()
	r.Post("/test/hello", http.HandlerFunc(testhandler))
	content := Process(NewAPIGateParser([]byte(ev1)), r).(map[string]interface{})

	type c struct {
		want, got interface{}
	}
	ts := map[string]c{
		"statusCode": c{http.StatusOK, content["statusCode"]},
		"body":       c{`{"test":"test body"}`, content["body"]},
		"headers":    c{map[string]string{"Test": "test header"}, content["headers"]}, //uppercase for header key
	}
	for k, v := range ts {
		if !reflect.DeepEqual(v.got, v.want) {
			t.Errorf("%s check wrong, want %+v, got %+v", k, v.want, v.got)
		}
	}
}

func testhandler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("var", bone.GetValue(r, "var"))
}

func TestWithContext(t *testing.T) {
	r := bone.New()
	r.Post("/test/:var", http.HandlerFunc(testhandler2))
	content := Process(NewAPIGateParser([]byte(ev1)), r).(map[string]interface{})

	type c struct {
		want, got interface{}
	}
	ts := map[string]c{
		"headers": c{map[string]string{"Var": "hello"}, content["headers"]}, //uppercase for header key
	}
	for k, v := range ts {
		if !reflect.DeepEqual(v.got, v.want) {
			t.Errorf("%s check wrong, want %+v, got %+v", k, v.want, v.got)
		}
	}
}

func testhandler3(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("param", bone.GetValue(r, "var"))
	w.(*LPResponse).WriteBody(map[string]string{"test": "test body"}, false)
}
func TestWithAlice(t *testing.T) {
	mux := bone.New()
	cors := alice.New(EnableCORS)
	mux.Post("/test/:var", cors.ThenFunc(testhandler3))
	content := Process(NewAPIGateParser([]byte(ev1)), mux).(map[string]interface{})
	log.Printf("content %+v", content)
	type c struct {
		want, got interface{}
	}
	ts := map[string]c{
		"headers": c{map[string]string{"Param": "hello", "Access-Control-Allow-Origin": "*"}, content["headers"]}, //uppercase for header key
	}
	for k, v := range ts {
		if !reflect.DeepEqual(v.got, v.want) {
			t.Errorf("%s check wrong, want %+v, got %+v", k, v.want, v.got)
		}
	}
}

func TestWriteResponse(t *testing.T) {
	bd1 := map[string]string{"key": "value"}
	type args struct {
		r        http.ResponseWriter
		i        interface{}
		isBase64 bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"t1", args{NewLPResponse(), bd1, false}},
		{"t2", args{}},
		{"t3", args{}},
		{"t4", args{}},
	}

	for _, tt := range tests {
		if tt.name == "t1" {
			t.Run(tt.name, func(t *testing.T) {
				WriteResponse(tt.args.r, tt.args.i, tt.args.isBase64)
				resp := tt.args.r.(*LPResponse).composeResp()
				if resp["body"] != `{"key":"value"}` {
					t.Error("body['key'] should be value")
				}
			})
		}
		if tt.name == "t2" {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				WriteResponse(w, "test2", false)
			}))
			res, err := http.Get(svr.URL)
			if err != nil {
				t.Errorf("create http request error %+v", err)
			}
			bs, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("read http test server error %+v", err)
			}
			if string(bs) != "test2" {
				t.Errorf("expected %+v, found %+v", "test2", string(bs))
			}
		}
		if tt.name == "t3" {
			ret := "test bytes"
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				WriteResponse(w, []byte(ret), false)
			}))
			res, err := http.Get(svr.URL)
			if err != nil {
				t.Errorf("create http request error %+v", err)
			}
			bs, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("read http test server error %+v", err)
			}
			if string(bs) != ret {
				t.Errorf("expected %s, found %s", ret, string(bs))
			}
		}
		if tt.name == "t4" {
			ret := map[string]string{"key1": "value1"}
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				WriteResponse(w, ret, false)
			}))
			res, err := http.Get(svr.URL)
			if err != nil {
				t.Errorf("create http request error %+v", err)
			}
			bs, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("read http test server error %+v", err)
			}
			if string(bs) != `{"key1":"value1"}` {
				t.Errorf("expected %s, found %s", ret, string(bs))
			}
		}
	}
}
