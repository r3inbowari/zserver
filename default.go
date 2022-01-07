package zserver

import (
	"crypto/tls"
	"github.com/r3inbowari/common"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

func SimpleServer(s *Server) {
	var err error
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
