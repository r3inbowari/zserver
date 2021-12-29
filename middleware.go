package zserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s *Server) LoggingMiddlewareBuilder() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UnixNano()
			unescape, err := url.QueryUnescape(r.RequestURI)
			if err != nil {
				s.Log.Errorf("[BSC] error route %s -> %s", r.RemoteAddr, unescape)
				return
			}
			next.ServeHTTP(w, r)
			end := time.Now().UnixNano()
			s.Log.WithFields(logrus.Fields{"time": fmt.Sprintf("%dms", (end-start)/100000)}).Infof("[BSC] route %s -> %s", r.RemoteAddr, unescape)
		})
	}
}

func (s *Server) AuthMiddlewareBuilder(checkFunc func(token string, r *http.Request) error) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI == "/login" || r.RequestURI == "/version" {

			} else {
				auth := r.Header.Get("Authorization")
				sa := strings.Split(auth, " ")
				if len(sa) != 2 {
					ResponseCommon(w, "unauthorized", "what", 1, http.StatusOK, 6401)
					return
				}
				err := checkFunc(sa[1], r)
				if err != nil {
					ResponseCommon(w, err.Error(), "about", 1, http.StatusOK, 6401)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
