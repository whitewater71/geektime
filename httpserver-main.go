package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/golang/glog"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	for k, v := range r.Header {
		for _, value := range v {
			w.Header().Set(k, value)
			fmt.Printf("%s=%s\n", k, value)
		}
	}
	w.Header().Set("Version", os.Getenv("VERSION"))
	fmt.Printf("VERSION=%s\n", os.Getenv("VERSION"))

	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		fmt.Printf("Remote IP=%s\n", forwarded)
	} else {
		fmt.Printf("Remote IP=%s\n", r.RemoteAddr)
	}
}

func wrapHandlerWithLogging(wrappedHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("--> %s %s", req.Method, req.URL.Path)
		lrw := NewLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lrw, req)
		statusCode := lrw.statusCode
		log.Printf("<-- %d %s", statusCode, http.StatusText(statusCode))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	io.WriteString(w, "ok\n")
}

func main() {
	flag.Set("v", "4")
	flag.Parse()
	glog.V(2).Info("Starting http server...")
	rootHandler := wrapHandlerWithLogging(http.HandlerFunc(handleRoot))
	http.Handle("/", rootHandler)
	http.HandleFunc("/healthz", healthz)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
