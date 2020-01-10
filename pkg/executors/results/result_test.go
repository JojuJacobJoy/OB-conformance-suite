package results

import (
	"encoding/json"
	"fmt"
	"os"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/netclient"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/test"

	"errors"
	"testing"
)

func TestNewTestCaseResult123(t *testing.T) {
	assert := test.NewAssert(t)

	err := errors.New("some error")
	result := NewTestCaseResult("123", true, NoMetrics(), []error{err}, "endpoint", "api-name", "api-version", "detailed description", "https://openbanking.org.uk/ref/uri")

	assert.Equal("123", result.Id)
	assert.True(result.Pass)
	assert.Equal(NoMetrics(), result.Metrics)
	assert.Equal(err.Error(), result.Fail[0])
}

func TestNewTestCaseResult321(t *testing.T) {
	assert := test.NewAssert(t)

	err := errors.New("some error")

	result := NewTestCaseResult("321", true, NoMetrics(), []error{err}, "endpoint", "api-name", "api-version", "detailed description", "https://openbanking.org.uk/ref/uri")
	assert.Equal("321", result.Id)
	assert.True(result.Pass)
	assert.Equal(NoMetrics(), result.Metrics)
	assert.Equal(err.Error(), result.Fail[0])
}

func TestNewTestCaseFailResult(t *testing.T) {
	assert := test.NewAssert(t)
	err := errors.New("some error")

	result := NewTestCaseFail("id", NoMetrics(), []error{err}, "endpoint", "api-name", "api-version", "detailed description", "https://openbanking.org.uk/ref/uri")

	assert.Equal("id", result.Id)
	assert.False(result.Pass)
	assert.Equal(NoMetrics(), result.Metrics)
	assert.Equal(err.Error(), result.Fail[0])
}

func TestTestCaseResultJsonMarshal(t *testing.T) {
	require := test.NewRequire(t)

	result := NewTestCaseResult("123", true, NoMetrics(), nil, "endpoint", "api-name", "api-version", "detailed description", "https://openbanking.org.uk/ref/uri")

	expected := `
{
	"endpoint": "endpoint",
	"id": "123",
	"pass": true,
	"metrics": {
		"response_time": 0,
		"response_size": 0
	},
	"detail": "detailed description",
	"refURI": "https://openbanking.org.uk/ref/uri"
}
	`
	actual, err := json.Marshal(result)
	require.NoError(err)
	require.NotEmpty(actual)

	require.JSONEq(expected, string(actual))
}

func TestNewTestCaseResultWithMetrics(t *testing.T) {
	assert := test.NewAssert(t)

	err := errors.New("some error")
	result := NewTestCaseResult("123", true, NoMetrics(), []error{err}, "endpoint", "api-name", "api-version", "detailed description", "https://openbanking.org.uk/ref/uri")

	assert.Equal("123", result.Id)
	assert.True(result.Pass)
	assert.Equal(NoMetrics(), result.Metrics)
	assert.Equal(err.Error(), result.Fail[0])
}

func TestLogToFile(t *testing.T) {
	f, err := os.OpenFile("testdata/test.log", os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		fmt.Println("Error opening file")
		t.Fail()
	}

	netclient.SetDebug(true)
	netclient.SetLoggerFile(f)

	resp, err := netclient.NewRequest().Get("http://httpbin.org/get")

	if err != nil {
		fmt.Printf("\nError: %v", err)
	}
	fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
	fmt.Printf("\nResponse Status: %v", resp.Status())
	fmt.Printf("\nResponse Body: %v", resp)
	fmt.Printf("\nResponse Time: %v", resp.Time())
	fmt.Printf("\nResponse Received At: %v\n", resp.ReceivedAt())

	ti := resp.Request.TraceInfo()
	fmt.Println("DNSLookup    :", ti.DNSLookup)
	fmt.Println("ConnTime     :", ti.ConnTime)
	fmt.Println("TLSHandshake :", ti.TLSHandshake)
	fmt.Println("ServerTime   :", ti.ServerTime)
	fmt.Println("ResponseTime :", ti.ResponseTime)
	fmt.Println("TotalTime    :", ti.TotalTime)
	fmt.Println("IsConnReused :", ti.IsConnReused)
	fmt.Println("IsConnWasIdle:", ti.IsConnWasIdle)
	fmt.Println("ConnIdleTime :", ti.ConnIdleTime)

}
