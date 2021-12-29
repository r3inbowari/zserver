package zserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/aliyun/fc-runtime-go-sdk/fc"
	"github.com/aliyun/fc-runtime-go-sdk/fccontext"
	"github.com/gorilla/mux"
	"github.com/r3inbowari/common"
	"github.com/sirupsen/logrus"
	"github.com/wuwenbao/gcors"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	r *mux.Router
	s *http.Server
	Options
}

type Options struct {
	Log    *logrus.Logger
	Addr   string
	Mode   common.Mode
	CaCert string
	CaKey  string
}

func DefaultServer(opts Options) *Server {
	server := NewServer(opts)
	server.r.Use(server.LoggingMiddlewareBuilder())
	server.Map("/hello", Hello)
	return server
}

func NewServer(opts Options) *Server {
	if opts.Log == nil {
		opts.Log = logrus.New()
	}
	if opts.Addr == "" {
		opts.Addr = ":9090"
	}
	r := mux.NewRouter()
	cors := gcors.New(
		r,
		gcors.WithOrigin("*"),
		gcors.WithMethods("POST, GET, PUT, DELETE, OPTIONS"),
		gcors.WithHeaders("Authorization"),
	)

	s := &http.Server{
		Addr:     opts.Addr,
		Handler:  cors,
		ErrorLog: log.New(opts.Log.Writer(), "[BSC] ", 0), // cc https://github.com/sirupsen/logrus/issues/1063
	}
	return &Server{r: r, s: s, Options: opts}
}

// AFCServerBuilder ali function compute builder
func (s *Server) AFCServerBuilder() interface{} {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		lc, _ := fccontext.FromContext(ctx)
		fmt.Printf("context: %#v\n", lc)
		fmt.Printf("request: %#v\n", r.Header)
		fmt.Printf("routing: %#v\n", r.URL.String())
		fmt.Printf("rmethod: %#v\n", r.Method)
		s.r.ServeHTTP(w, r)
		return nil
	}
}

func (s *Server) Start() {
	var err error
	// select ali function compute
	if s.Mode == common.ALI {
		fc.StartHttp(s.AFCServerBuilder())
		return
	}
	// listen
	s.Log.Info("[BCS] listened on " + s.Addr)
	if s.CaCert != "" && s.CaKey != "" {
		_, err := tls.LoadX509KeyPair(s.CaCert, s.CaKey)
		if err != nil {
			s.Log.Warn("[BSC] please check your cert path whether is right, downgrading to http now")
			err = s.s.ListenAndServe()
		}
		s.Log.Info("[BSC] tls enabled")
		err = s.s.ListenAndServeTLS(s.CaCert, s.CaKey)
	} else {
		err = s.s.ListenAndServe()
	}
	// finally
	if strings.HasSuffix(err.Error(), "normally permitted.") || strings.Index(err.Error(), "bind") != -1 {
		s.Log.WithField("addr", s.s.Addr).Error("[BCS] socket's port is occupied.")
		common.Exit(common.SocketOccupy)
	}
	// goroutine block here not need exit
	s.Log.WithFields(logrus.Fields{"err": err.Error()}).Info("[BCS] service has been terminated")
	time.Sleep(time.Second * 10)
}

func (s *Server) Shutdown(ctx context.Context) {
	if s.s != nil {
		err := s.s.Shutdown(ctx)
		if err != nil {
			s.Log.Error("[BSC] shutdown failed")
			common.Exit(common.SocketShutdownFailed)
		}
		s.Log.Info("[BSC] release completed")
	}
}

func (s *Server) Map(path string, f func(http.ResponseWriter,
	*http.Request), method ...string) *Server {
	if len(method) == 1 {
		s.Log.Info("[BSC] add route path [" + method[0] + "] -> " + path)
		s.r.HandleFunc(path, f).Methods(method[0])
	} else {
		s.Log.Info("[BSC] add route path [ALL] -> " + path)
		s.r.HandleFunc(path, f)
	}
	return s
}
