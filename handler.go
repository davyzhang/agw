package agw

import "net/http"

func EnableCORS(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
