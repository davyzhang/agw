package agw

import (
	"net/url"
	"reflect"
	"testing"
)

var ev = `
{
        "path": "/test/hello",
        "headers": {
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
            "Accept-Encoding": "gzip, deflate, lzma, sdch, br",
            "Accept-Language": "en-US,en;q=0.8",
            "CloudFront-Forwarded-Proto": "https",
            "CloudFront-Is-Desktop-Viewer": "true",
            "CloudFront-Is-Mobile-Viewer": "false",
            "CloudFront-Is-SmartTV-Viewer": "false",
            "CloudFront-Is-Tablet-Viewer": "false",
            "CloudFront-Viewer-Country": "US",
            "Host": "wt6mne2s9k.execute-api.us-west-2.amazonaws.com",
            "Upgrade-Insecure-Requests": "1",
            "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.82 Safari/537.36 OPR/39.0.2256.48",
            "Via": "1.1 fb7cca60f0ecd82ce07790c9c5eef16c.cloudfront.net (CloudFront)",
            "X-Amz-Cf-Id": "nBsWBOrSHMgnaROZJK1wGCZ9PcRcSpq_oSXZNQwQ10OTZL4cimZo3g==",
            "X-Forwarded-For": "192.168.100.1, 192.168.1.1",
            "X-Forwarded-Port": "443",
            "X-Forwarded-Proto": "https"
        },
        "pathParameters": {"proxy": "hello"},
        "requestContext": {
            "accountId": "123456789012",
            "resourceId": "us4z18",
            "stage": "test",
            "requestId": "41b45ea3-70b5-11e6-b7bd-69b5aaebc7d9",
            "identity": {
                "cognitoIdentityPoolId": "",
                "accountId": "",
                "cognitoIdentityId": "",
                "caller": "",
                "apiKey": "",
                "sourceIp": "192.168.100.1",
                "cognitoAuthenticationType": "",
                "cognitoAuthenticationProvider": "",
                "userArn": "",
                "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.82 Safari/537.36 OPR/39.0.2256.48",
                "user": ""
            },
            "resourcePath": "/{proxy+}",
            "httpMethod": "GET",
            "apiId": "wt6mne2s9k"
        },
        "resource": "/{proxy+}",
        "httpMethod": "GET",
        "queryStringParameters": {"name": "me"},
        "stageVariables": {"stageVarName": "stageVarValue"},
		"body":"{\"test\":\"body\"}",
		"isBase64Encoded":false
    }`[1:]

func Test_Parse(t *testing.T) {
	lpe, err := NewLambdaProxyEvent([]byte(ev))
	if err != nil {
		t.Errorf("new event err %+v", err)
	}
	type check struct {
		want, got interface{}
	}
	bd, err := lpe.Body()
	if err != nil {
		t.Errorf("body parse error %+v", err)
	}
	tc := map[string]check{
		"path":  check{"/test/hello", lpe.Path()},
		"Query": check{url.Values(map[string][]string{"name": []string{"me"}}), lpe.QueryStringParameters()},
		"Body":  check{"body", bd.Get("test").MustString()},
	}
	for k, v := range tc {
		if !reflect.DeepEqual(v.got, v.want) {
			t.Errorf("%s parse error, want %+v, got %+v", k, v.want, v.got)
		}
	}
}

func Test_parseJSONType(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseJSONType(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseJSONType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toStringMap(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toStringMap(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toStringMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
