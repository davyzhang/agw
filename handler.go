package agw

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"time"
)

type contextKey string

const (
	//ContextKeyBody for json body parse
	ContextKeyBody = contextKey("body")
)

/*EnableCORS will add CORS headers to request.
AWS apigateway won't add them even after choosen to enable CORS in console,
*/
func EnableCORS(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Methods", "*")
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func ParseBodyBytes(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("read body err %+v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer func() {
			if err := r.Body.Close(); err != nil {
				log.Printf("close http request error %+v", err)
				return
			}
		}()
		ctx := context.WithValue(r.Context(), ContextKeyBody, bs)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("[%.05fs][%s:%q]", time.Since(st).Seconds(), r.Method, r.URL.String())
	})
}
