package results

import (
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"
	"github.com/go-resty/resty/v2"
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
		ResponseTime  float64       `json:"response_time,omitempty"`
		ResponseSize  int           `json:"response_size,omitempty"`
		DNSLookup     time.Duration `json:"dns_lookup"`
		ConnTime      time.Duration `json:"conn_time"`
		TLSHandshake  time.Duration `json:"tls_handshake"`
		ServerTime    time.Duration `json:"server_time"`
		TotalTime     time.Duration `json:"total_time"`
		IsConnReused  bool          `json:"is_conn_reused"`
		IsConnWasIdle bool          `json:"is_conn_was_idle"`
		ConnIdleTime  time.Duration `json:"conn_idle_time"`
	}{
		ResponseTime: float64(m.ResponseTime) / float64(time.Millisecond),
		ResponseSize: int(m.ResponseSize),
	})
}

func NoMetrics() Metrics {
	return Metrics{}
}

func NewMetricsFromRestyResponse(testCase *model.TestCase, response *resty.Response) Metrics {
	ti := response.Request.TraceInfo()
	fmt.Println("DNSLookup    :", ti.DNSLookup)
	fmt.Println("ConnTime     :", ti.ConnTime)
	fmt.Println("TLSHandshake :", ti.TLSHandshake)
	fmt.Println("ServerTime   :", ti.ServerTime)
	fmt.Println("ResponseTime :", ti.ResponseTime)
	fmt.Println("TotalTime    :", ti.TotalTime)
	fmt.Println("IsConnReused :", ti.IsConnReused)
	fmt.Println("IsConnWasIdle:", ti.IsConnWasIdle)
	fmt.Println("ConnIdleTime :", ti.ConnIdleTime)

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
