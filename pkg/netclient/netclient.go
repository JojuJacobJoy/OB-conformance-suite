package netclient

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

var (
	tlscert     tls.Certificate
	tlsconfig   *tls.Config
	certSet     bool
	debug       bool
	client      *resty.Client
	redirpolicy interface{}
	logfile     *os.File
	outlog      *log.Logger
)

type Logger struct {
}

func SetTLSClientConfig(config *tls.Config) {
	tlsconfig = config
}

func GetClient() *resty.Client {
	if client == nil {
		client = resty.New()
		if tlsconfig != nil {
			client.SetTLSClientConfig(tlsconfig)
		}
		if debug {
			client.SetDebug(true)
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

func NewRequest() *resty.Request {
	client := GetClient()
	return client.R()
}

func SetDebug(debugflag bool) {
	// debug = debugflag
	// outlog = log.New(os.Stderr, "", 0)
	// outlog.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func SetLoggerFile(lfile *os.File) {
	debug = true
	outlog = log.New(lfile, "", 0)
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
