package agw

import (
	"log"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-zoo/bone"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func testhandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("test", "test header")
	w.(*LPResponse).WriteBody(map[string]string{"test": "test body"})
}

func TestGorillaMuxFlow(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/test/hello", testhandler)
	content, err := Process([]byte(ev), r)
	if err != nil {
		t.Errorf("flow test error %+v", err)
	}
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
	r.Get("/test/hello", http.HandlerFunc(testhandler))
	content, err := Process([]byte(ev), r)
	if err != nil {
		t.Errorf("flow test error %+v", err)
	}
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
	r.Get("/test/:var", http.HandlerFunc(testhandler2))
	content, err := Process([]byte(ev), r)
	if err != nil {
		t.Errorf("flow test error %+v", err)
	}
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

func TestWithAlice(t *testing.T) {
	mux := bone.New()
	cors := alice.New(EnableCORS)
	mux.Get("/test/:var", cors.ThenFunc(testhandler))
	content, err := Process([]byte(ev), mux)
	if err != nil {
		t.Errorf("flow test error %+v", err)
	}
	log.Printf("content %+v", content)
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

func Test_composeURL(t *testing.T) {
	lpe, err := NewLambdaProxyEvent([]byte(ev))
	if err != nil {
		t.Errorf("composeUrl error %+v", err)
		return
	}
	type args struct {
		lpe *LambdaProxyEvent
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"t1", args{lpe}, "/test/hello?name=me"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := composeURL(tt.args.lpe); got != tt.want {
				t.Errorf("composeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
