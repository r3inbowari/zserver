package zserver

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/r3inbowari/common"
	"github.com/sirupsen/logrus"
	"github.com/wuwenbao/gcors"
	"log"
	"net/http"
)

type Server struct {
	r *mux.Router
	s *http.Server
	Options
}

var Assembly map[string]func(s *Server)

type Options struct {
	Log          *logrus.Logger
	Addr         string
	CaCert       string
	CaKey        string
	AssemblyName string
}

func DefaultServer(opts Options) *Server {
	server := NewServer(opts)
	server.r.Use(server.LoggingMiddlewareBuilder())
	server.Map("/hello", Hello)
	if opts.AssemblyName == "" {
		server.AssemblyName = "default"
	}
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

func RegisterAssembly(name string, f func(s *Server)) {
	Assembly[name] = f
}

func init() {
	Assembly = make(map[string]func(s *Server))
	Assembly["default"] = SimpleServer
}

func (s *Server) Start() {
	f, ok := Assembly[s.AssemblyName]
	if ok {
		f(s)
	}
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
