package results

import (
	"encoding/json"

	"bitbucket.org/openbankingteam/conformance-suite/internal/pkg/test"

	"errors"
	"testing"
)

func TestNewTestCaseResult123(t *testing.T) {
	assert := test.NewAssert(t)

	err := errors.New("some error")
	result := NewTestCaseResult("123", true, NoMetrics, err)

	assert.Equal("123", result.Id)
	assert.True(result.Pass)
	assert.Equal(NoMetrics, result.Metrics)
	assert.Equal(err.Error(), result.Fail)
}

func TestNewTestCaseResult321(t *testing.T) {
	assert := test.NewAssert(t)

	err := errors.New("some error")

	result := NewTestCaseResult("321", true, NoMetrics, err)
	assert.Equal("321", result.Id)
	assert.True(result.Pass)
	assert.Equal(NoMetrics, result.Metrics)
	assert.Equal(err.Error(), result.Fail)
}

func TestNewTestCaseFailResult(t *testing.T) {
	assert := test.NewAssert(t)
	err := errors.New("some error")

	result := NewTestCaseFail("id", NoMetrics, err)

	assert.Equal("id", result.Id)
	assert.False(result.Pass)
	assert.Equal(NoMetrics, result.Metrics)
	assert.Equal(err.Error(), result.Fail)
}

func TestTestCaseResultJsonMarshal(t *testing.T) {
	require := test.NewRequire(t)

	result := NewTestCaseResult("123", true, NoMetrics, nil)

	expected := `
{
	"id": "123",
	"pass": true,
	"metrics": {
		"response_time": 0,
		"response_size": 0
	}
}
	`
	actual, err := json.Marshal(result)
	require.NoError(err)
	require.NotEmpty(actual)

	require.JSONEq(expected, string(actual))
}
