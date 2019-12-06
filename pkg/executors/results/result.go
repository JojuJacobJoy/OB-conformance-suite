package results

import (
	"encoding/json"
	"gopkg.in/resty.v1"
)

// TestCase result for a run
type TestCase struct {
	Id         string   `json:"id"`
	Pass       bool     `json:"pass"`
	Metrics    Metrics  `json:"metrics"`
	Fail       []DetailError `json:"fail,omitempty"`
	Detail     string   `json:"detail"`
	RefURI     string   `json:"refURI"`
	Endpoint   string   `json:"endpoint"`
	API        string   `json:"-"`
	APIVersion string   `json:"-"`
}

// NewTestCaseFail returns a failed test
func NewTestCaseFail(id string, metrics Metrics, errs []DetailError, endpoint, api, apiVersion, detail, refURI string) TestCase {
	return NewTestCaseResult(id, false, metrics, errs, endpoint, api, apiVersion, detail, refURI)
}

// NewTestCaseResult return a new TestCase instance
func NewTestCaseResult(id string, pass bool, metrics Metrics, errs []DetailError, endpoint, apiName, apiVersion, detail, refURI string) TestCase {
	return TestCase{
		API:        apiName,
		APIVersion: apiVersion,
		Id:         id,
		Pass:       pass,
		Metrics:    metrics,
		Fail:       errs,
		Endpoint:   endpoint,
		Detail:     detail,
		RefURI:     refURI,
	}
}

type ResultKey struct {
	APIName    string
	APIVersion string
}

type DetailError struct {
	GeneralError string `json:"generalError,omitempty"`
	EndpointResponse interface{} `json:"endpointResponse,omitempty"`
	TestCaseMessage  string `json:"testCaseMessage,omitempty"`
}

func (de DetailError) Error() string {
	j, _ := json.Marshal(de)

	return string(j)
}

func DetailedErrors(errs []error, resp *resty.Response) []DetailError {
	var detailedErrors []DetailError
	for _, err := range errs {

		detailedError := DetailError{
			TestCaseMessage:  err.Error(),
		}
		err := json.Unmarshal(resp.Body(), &detailedError.EndpointResponse)
		if err != nil {
			// Couldn't decode the response body as JSON, so just use it as is.
			// Sometimes some random HTML comes back, where there is a server level error
			detailedError.EndpointResponse = string(resp.Body())
		}
		detailedErrors = append(detailedErrors, detailedError)
	}
	return detailedErrors
}
