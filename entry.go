package agw

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

//LambdaContext is used to store system environment variables
var LambdaContext *lambdacontext.LambdaContext

//RawMessage is used to store the complete content from apigateway
var RawMessage json.RawMessage

//GatewayHandler is a suitable type for handle apigateway messages
//use RawMessage to delay parsing
type GatewayHandler func(context.Context, json.RawMessage) (map[string]interface{}, error)

//Handler return handler function for apigateway
func Handler(h http.Handler) GatewayHandler {
	return func(ctx context.Context, content json.RawMessage) (map[string]interface{}, error) {
		RawMessage = content
		var ok bool
		LambdaContext, ok = lambdacontext.FromContext(ctx)
		if !ok {
			return nil, errors.New("no valid lambda context found")
		}
		return Process(content, h), nil
		// return `{"statusCode":200,"headers":{},"body":"{}","isBase64Encoded":false}`, nil
	}
}
