package results

import (
	"encoding/json"
	"time"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type Metrics struct {
	TestCase      *model.TestCase
	ResponseTime  time.Duration // ResponseTime is a duration since first response byte from server to request completion.
	ResponseSize  int64         // Size in bytes of the HTTP Response body
	DNSLookup     time.Duration
	ConnTime      time.Duration // ConnTime is a duration that took to obtain a successful connection.
	TLSHandshake  time.Duration // TLSHandshake is a duration that TLS handshake took place.
	ServerTime    time.Duration // ServerTime is a duration that server took to respond first byte.
	TotalTime     time.Duration // TotalTime is a duration that total request took end-to-end.
	IsConnReused  bool          // IsConnReused is whether this connection has been previously used for another HTTP request.
	IsConnWasIdle bool          // IsConnWasIdle is whether this connection was obtained from an idle pool.
	ConnIdleTime  time.Duration // ConnIdleTime is a duration how long the connection was previously idle, if IsConnWasIdle is true.
}

// MarshalJSON is a custom marshaler which formats a Metrics struct
// with a response time represented as unit of milliseconds
// response time decimal precision is up the nanosecond eg: 1.234ms
func (m Metrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ResponseTime  float64 `json:"response_time"`
		ResponseSize  int     `json:"response_size"`
		DNSLookup     float64 `json:"dns_lookup"`
		ConnTime      float64 `json:"conn_time"`
		TLSHandshake  float64 `json:"tls_handshake"`
		ServerTime    float64 `json:"server_time"`
		TotalTime     float64 `json:"total_time"`
		IsConnReused  bool    `json:"is_conn_reused"`
		IsConnWasIdle bool    `json:"is_conn_was_idle"`
		ConnIdleTime  float64 `json:"conn_idle_time"`
	}{
		ResponseTime:  float64(m.ResponseTime) / float64(time.Millisecond),
		ResponseSize:  int(m.ResponseSize),
		DNSLookup:     float64(m.DNSLookup) / float64(time.Millisecond),
		ConnTime:      float64(m.ConnTime) / float64(time.Millisecond),
		TLSHandshake:  float64(m.TLSHandshake) / float64(time.Millisecond),
		ServerTime:    float64(m.ServerTime) / float64(time.Millisecond),
		TotalTime:     float64(m.TotalTime) / float64(time.Millisecond),
		IsConnReused:  m.IsConnReused,
		IsConnWasIdle: m.IsConnWasIdle,
		ConnIdleTime:  float64(m.ConnIdleTime) / float64(time.Millisecond),
	})
}

func NoMetrics() Metrics {
	return Metrics{}
}

func NewMetricsFromRestyResponse(testCase *model.TestCase, response *resty.Response) Metrics {

	return NewMetricsWithTrace(testCase, response)

}

func NewMetrics(testCase *model.TestCase, responseTime time.Duration, responseSize int) Metrics {
	return Metrics{
		TestCase:     testCase,
		ResponseTime: responseTime,
		ResponseSize: int64(responseSize),
	}
}

func NewMetricsWithTrace(testCase *model.TestCase, response *resty.Response) Metrics {

	ti := response.Request.TraceInfo()

	logrus.WithFields(logrus.Fields{
		"CaseID":        testCase.ID,
		"ResponseTime":  response.Time(),
		"ResponseSize":  response.Size(),
		"DNSLookup":     ti.DNSLookup,
		"ConnTime":      ti.ConnTime,
		"TLSHandshake":  ti.TLSHandshake,
		"ServerTime":    ti.ServerTime,
		"TotalTime":     ti.TotalTime,
		"IsConnReused":  ti.IsConnReused,
		"IsConnWasIdle": ti.IsConnWasIdle,
		"ConnIdleTime":  ti.ConnIdleTime,
	}).Trace("ResponseInfo")

	return Metrics{
		TestCase:      testCase,
		ResponseTime:  response.Time(),
		ResponseSize:  response.Size(),
		DNSLookup:     ti.DNSLookup,
		ConnTime:      ti.ConnTime,
		TLSHandshake:  ti.TLSHandshake,
		ServerTime:    ti.ServerTime,
		TotalTime:     ti.TotalTime,
		IsConnReused:  ti.IsConnReused,
		IsConnWasIdle: ti.IsConnWasIdle,
		ConnIdleTime:  ti.ConnIdleTime,
	}
}
