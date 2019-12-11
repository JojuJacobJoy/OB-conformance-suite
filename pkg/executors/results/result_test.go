package results

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"
	"io/ioutil"
	"net/http"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/test"

	"testing"
)

func TestNewTestCaseResult123(t *testing.T) {
	assert := test.NewAssert(t)

	de := fooBarDetailedError()
	result := NewTestCaseResult("123", true, NoMetrics(), []DetailError{de}, "endpoint", "api-name", "api-version", "detailed description", "https://openbanking.org.uk/ref/uri")

	assert.Equal("123", result.Id)
	assert.True(result.Pass)
	assert.Equal(NoMetrics(), result.Metrics)
	assert.Equal(de, result.Fail[0])
}

func TestNewTestCaseResult321(t *testing.T) {
	assert := test.NewAssert(t)

	de := fooBarDetailedError()
	result := NewTestCaseResult("321", true, NoMetrics(), []DetailError{de}, "endpoint", "api-name", "api-version", "detailed description", "https://openbanking.org.uk/ref/uri")
	assert.Equal("321", result.Id)
	assert.True(result.Pass)
	assert.Equal(NoMetrics(), result.Metrics)
	assert.Equal(de, result.Fail[0])
}

func TestNewTestCaseFailResult(t *testing.T) {
	assert := test.NewAssert(t)

	de := fooBarDetailedError()
	result := NewTestCaseFail("id", NoMetrics(), []DetailError{de}, "endpoint", "api-name", "api-version", "detailed description", "https://openbanking.org.uk/ref/uri")

	assert.Equal("id", result.Id)
	assert.False(result.Pass)
	assert.Equal(NoMetrics(), result.Metrics)
	assert.Equal(de, result.Fail[0])
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

func fooBarDetailedError() DetailError {
	return DetailError{
		GeneralError:     "some-error",
		EndpointResponse: `{"foo":"bar"}`,
		TestCaseMessage:  "same-test-case-message",
	}
}

func TestDetailedErrors(t *testing.T) {
	testTable := []struct{
		errors []error
		respBody []byte
		expectedResult []DetailError
	}{
		{
			errors: []error{ errors.New("an error")},
			respBody: []byte(""),
			expectedResult: []DetailError{
				{
					TestCaseMessage: "an error",
					EndpointResponse: "",
				},
			},
		},
		{
			errors: []error{ errors.New("an error")},
			respBody: []byte(`sdsadasdas{{[[}sada"sd}\"s=a's=das-d\''asd""0asd8as7d6a5DAS6D7//DSADSADadadamasdm/`),
			expectedResult: []DetailError{
				{
					TestCaseMessage: "an error",
					EndpointResponse: `sdsadasdas{{[[}sada"sd}\"s=a's=das-d\''asd""0asd8as7d6a5DAS6D7//DSADSADadadamasdm/`,
				},
			},
		},
		{
			errors: []error{ errors.New("an error")},
			respBody: []byte(`{ "message": "This is valid JSON", "isValid": true }`),
			expectedResult: []DetailError{
				{
					GeneralError: nil,
					TestCaseMessage: "an error",
					EndpointResponse: struct{
						Message string `json:"message"`
						IsValid bool `json:"isValid"`
					}{
						Message: "This is valid JSON",
						IsValid: true,
					},
				},
			},
		},
		{
			errors: []error{ errors.New("an error")},
			respBody: []byte(`<html><body>Some random html here <a href="https://www.openbanking.org.uk">link</a></body></html>`),
			expectedResult: []DetailError{
				{
					TestCaseMessage: "an error",
					EndpointResponse: `<html><body>Some random html here <a href="https://www.openbanking.org.uk">link</a></body></html>`,
				},
			},
		},
		{
			errors: []error{ errors.New("foobar err 1"), errors.New("error 2")},
			respBody: []byte(`<html><body>Some random html here <a href="https://www.openbanking.org.uk">link</a></body></html>`),
			expectedResult: []DetailError{
				{
					TestCaseMessage: "foobar err 1",
					EndpointResponse: `<html><body>Some random html here <a href="https://www.openbanking.org.uk">link</a></body></html>`,

				},
				{
					TestCaseMessage: "error 2",
					EndpointResponse: `<html><body>Some random html here <a href="https://www.openbanking.org.uk">link</a></body></html>`,
				},
			},
		},
	}

	for _, testItem := range testTable {
		restyResp := &resty.Response{
			RawResponse: &http.Response{
				Body: ioutil.NopCloser(bytes.NewReader(testItem.respBody)),
			},
		}
		result := DetailedErrors(testItem.errors, restyResp)
		assert.Len(t, result, len(testItem.errors))
		assert.EqualValues(t, testItem.expectedResult, result)
	}
}
