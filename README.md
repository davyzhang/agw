# AGW
 AGW transform AWS lambda event message to standard http.Request which can make it easy to work with existing http routers and chaining libraries  

In short, the usage is
```go
//your standard http handler
func testhandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("test", "test header")
	w.(*agw.LPResponse).WriteBody(map[string]string{"test": "test body"})
}

//get aws request event message []byte from commdline from nodejs or whatever wrapper you are using
func AWSLambdaProxyHandler(lambdaEventMessage []byte ){
	//use any exsiting router supporting the standard http.Handler 
	//like 	"github.com/gorilla/mux"
    mux := mux.NewRouter()
    mux.HandleFunc("/test/hello", testhandler)

    ret,err := agw.Process(lambdaEventMessage,mux)
    //write back ret and handle err...
}

```

### The Full Picture
To use it in the real project we need some more setups
 1. AWS APIGateway **must** be configured with lambda **proxy** mode, typically like {/proxy+} here's the [doc](http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-create-api-as-simple-proxy-for-lambda.html) from aws
 2. Use any wrapper to get the lambda proxy request string, which is a pure json data, you can get it from nodejs shim or python shim. Personally I recommend [Apex](https://github.com/apex/apex) and [Go-Apex](https://github.com/apex/go-apex "Go-Apex")
 3. Using any http router like lightning fast [Bone](https://github.com/go-zoo/bone) or popular and feature rich [Gorilla Mux](https://github.com/gorilla/mux) and even chaining libraries like [Alice](https://github.com/justinas/alice) to write your middlewares


### Practical Example
```go

func main() {
	mux := bone.New()
	cors := alice.New(agw.EnableCORS)
	h := cors.Append(yourMiddleware)
	mux.Get("/test", h.ThenFunc(yourHandler)

	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		//the simplest way
		return agw.Process(event,mux)

		//if you want to handle the event before processing messages
		//lpe, err := agw.NewLambdaProxyEvent([]byte(event))
		//if err != nil {
		//	return nil, err
		//}
		//alias := lpe.StageVariables()["lbAlias"]
		//do something with the alias
		//return agw.ProcessEvent(lpe, mux)
	})
}

```

###Notes

 - The ResponseWriter.Write([]byte) (int, error) is not going to work as normal http response due to the way how lambda and aws apigateway works
 - You have to type assert  ResponseWriter as (*agw.LPResponse) and use WriteBody(out) to set the return body 
```go
 func MyHandler(w http.ResponseWriter, r *http.Request) {
	//your logic ....
	w.(*agw.LPResponse).WriteBody(out)
}
```
- Since the AWS event message is evolving during the time, AGW uses [simplejson](https://github.com/bitly/go-simplejson) as the major json parser to extract only the useful key and values.
- Read request json body with simplejson or ParseJSONBody middleware 
```go
func handler(w http.ResponseWriter, r *http.Request) {
	sj, err := simplejson.NewFromReader(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
//from middleware
jBody := alice.New(agw.ParseJSONBody)
mux.Post("/test", jBody.ThenFunc(handler))

func handler(w http.ResponseWriter, r *http.Request) {
	b := r.Context().Value(agw.ContextKeyBody).(*simplejson.Json)
	val1 := b.Get("yourkey").MustString()
	//...
}
```
- 
- If a returned body is a pure string or number, it will be returned as a plaintext instead of a json object with quotes
- Context is working as expected.

###TODO

 - More tests
 - Comments

###License
BSD licensed. See the LICENSE file for details.