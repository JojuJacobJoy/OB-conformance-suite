package netclient

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

var (
	tlsconfig   *tls.Config
	debug       bool
	client      *resty.Client
	redirpolicy interface{}
	logfile     *os.File
	outlog      *log.Logger
)

type Logger struct {
}

func NewRequest() *resty.Request {
	client := getClient()
	return client.R().EnableTrace()
}

func SetTLSClientConfig(config *tls.Config) {
	tlsconfig = config
	client = nil
}

func getClient() *resty.Client {
	if client == nil {
		client = resty.New()
		if tlsconfig != nil {
			client.SetTLSClientConfig(tlsconfig)
		}
		if debug {
			client.SetDebug(true)
			client.EnableTrace()
		}
		if logfile != nil {
			client.SetLogger(Logger{})
		}
		if redirpolicy != nil {
			client.SetRedirectPolicy(redirpolicy)
		}
	} else {
		logrus.Debug("netclient:GetExistingClient")
	}
	return client
}

func SetDebug(debugflag bool) {
	debug = debugflag
}

func SetLoggerFile(lfile *os.File) {
	debug = true
	outlog = log.New(lfile, "", 0)
	logfile = lfile
	outlog.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	outlog.SetOutput(lfile)
}

func SetRedirectPolicy(policy interface{}) {
	redirpolicy = policy
}

func NewLogger() resty.Logger {
	return Logger{}
}

func (l Logger) Errorf(format string, v ...interface{}) {
	outlog.Printf(format, v...)
}

func (l Logger) Warnf(format string, v ...interface{}) {
	outlog.Printf(format, v...)
}

func (l Logger) Debugf(format string, v ...interface{}) {
	outlog.Printf(format, v...)
}
