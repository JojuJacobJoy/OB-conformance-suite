package discovery

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"

	"github.com/stretchr/testify/assert"
)

// invalidTestCase
// name - name of the test case.
// config - the discovery model config.
// err - the expected err
type invalidTestCase struct {
	name        string
	config      string
	expectedErr string
}

// conditionalityCheckerMock - implements model.ConditionalityChecker interface for tests
type conditionalityCheckerMock struct {
}

// IsOptional - not used in discovery test
func (c conditionalityCheckerMock) IsOptional(method, endpoint string, specification string) (bool, error) {
	return false, nil
}

// Returns IsMandatory true for POST /account-access-consents, false for all other endpoint/methods.
func (c conditionalityCheckerMock) IsMandatory(method, endpoint string, specification string) (bool, error) {
	if method == "POST" && endpoint == "/account-access-consents" {
		return true, nil
	}
	return false, nil
}

// IsConditional - not used in discovery test
func (c conditionalityCheckerMock) IsConditional(method, endpoint string, specification string) (bool, error) {
	return false, nil
}

// Returns IsPresent true for valid GET/POST/DELETE endpoints.
func (c conditionalityCheckerMock) IsPresent(method, endpoint string, specification string) (bool, error) {
	if method == "GET" || method == "POST" || method == "DELETE" {
		return true, nil
	}
	return false, nil
}

// Returns that "POST" "/account-access-consent" is missing
func (c conditionalityCheckerMock) MissingMandatory(endpoints []model.Input, specification string) ([]model.Input, error) {
	missing := []model.Input{}
	missing = append(missing, model.Input{Method: "POST", Endpoint: "/account-access-consents"})
	return missing, nil
}

func loadDiscoveryExample(t *testing.T) *Model {
	discoveryJSON, err := ioutil.ReadFile("../../docs/discovery-example.json")
	assert.NoError(t, err)
	assert.NotNil(t, discoveryJSON)

	discovery := &Model{}
	err = json.Unmarshal(discoveryJSON, &discovery)
	assert.Nil(t, err)
	return discovery
}

func TestValidate(t *testing.T) {
	discovery := loadDiscoveryExample(t)

	result, failures, err := Validate(model.NewConditionalityChecker(), discovery)

	assert.Nil(t, err)
	assert.Equal(t, failures, make([]string, 0))
	assert.True(t, result)
}

func TestDiscovery_FromJSONString_Invalid_Cases(t *testing.T) {
	testCases := []invalidTestCase{
		{
			name:        `json_needs_to_be_valid`,
			config:      ` `,
			expectedErr: `unexpected end of JSON input`,
		},
		{
			name:   `version_and_discoveryItems_array_needs_to_specified`,
			config: `{}`,
			expectedErr: `Key: 'Model.DiscoveryModel.Version' Error:Field validation for 'Version' failed on the 'required' tag
Key: 'Model.DiscoveryModel.DiscoveryItems' Error:Field validation for 'DiscoveryItems' failed on the 'required' tag`,
		},
		{
			name: `discoveryItems_array_needs_to_be_greater_than_one`,
			config: `
{
  "discoveryModel": {
	"version": "v0.0.1",
	"discoveryItems": [
	]
  }
}
			`,
			expectedErr: `Key: 'Model.DiscoveryModel.DiscoveryItems' Error:Field validation for 'DiscoveryItems' failed on the 'gt' tag`,
		},
		{
			name: `endpoints_needs_to_be_specified`,
			config: `
{
	"discoveryModel": {
		"version": "v0.0.1",
		"discoveryItems": [
			{
				"apiSpecification": {
					"name": "Account and Transaction API Specification",
					"url": "https://openbanking.atlassian.net/wiki/spaces/DZ/pages/642090641/Account+and+Transaction+API+Specification+-+v3.0",
					"version": "v3.0",
					"schemaVersion": "https://raw.githubusercontent.com/OpenBankingUK/read-write-api-specs/v3.0.0/dist/account-info-swagger.json"
				},
				"openidConfigurationUri": "https://as.aspsp.ob.forgerock.financial/oauth2/.well-known/openid-configuration",
				"resourceBaseUri": "https://rs.aspsp.ob.forgerock.financial:443/",
				"endpoints": [
				]
			}
		]
	}
}
			`,
			expectedErr: `Key: 'Model.DiscoveryModel.DiscoveryItems[0].Endpoints' Error:Field validation for 'Endpoints' failed on the 'gt' tag`,
		},
		{
			name: `endpoints_path_and_method_need_to_be_valid`,
			config: `
{
	"discoveryModel": {
		"version": "v0.0.1",
		"discoveryItems": [
			{
				"apiSpecification": {
					"name": "Account and Transaction API Specification",
					"url": "https://openbanking.atlassian.net/wiki/spaces/DZ/pages/642090641/Account+and+Transaction+API+Specification+-+v3.0",
					"version": "v3.0",
					"schemaVersion": "https://raw.githubusercontent.com/OpenBankingUK/read-write-api-specs/v3.0.0/dist/account-info-swagger.json"
				},
				"openidConfigurationUri": "https://as.aspsp.ob.forgerock.financial/oauth2/.well-known/openid-configuration",
				"resourceBaseUri": "https://rs.aspsp.ob.forgerock.financial:443/",
				"endpoints": [
					{
						"method": "FAKE-METHOD",
						"path": "/fake-path"
					},
					{
						"method": "FAKE-METHOD2",
						"path": "/fake-path2"
					}
				]
			}
		]
	}
}
			`,
			expectedErr: `discoveryItemIndex=0, invalid endpoint Method=FAKE-METHOD, Path=/fake-path
discoveryItemIndex=0, invalid endpoint Method=FAKE-METHOD2, Path=/fake-path2`,
		},
		{
			name: `endpoints_missing_mandatory_endpoints_accounts`,
			config: `
{
	"discoveryModel": {
		"version": "v0.0.1",
		"discoveryItems": [
			{
				"apiSpecification": {
					"name": "Account and Transaction API Specification",
					"url": "https://openbanking.atlassian.net/wiki/spaces/DZ/pages/642090641/Account+and+Transaction+API+Specification+-+v3.0",
					"version": "v3.0",
					"schemaVersion": "https://raw.githubusercontent.com/OpenBankingUK/read-write-api-specs/v3.0.0/dist/account-info-swagger.json"
				},
				"openidConfigurationUri": "https://as.aspsp.ob.forgerock.financial/oauth2/.well-known/openid-configuration",
				"resourceBaseUri": "https://rs.aspsp.ob.forgerock.financial:443/",
				"endpoints": [
					{
						"method": "GET",
						"path": "/accounts/{AccountId}/balances"
					}
				]
			}
		]
	}
}
			`,
			expectedErr: `discoveryItemIndex=0, missing mandatory endpoint Method=POST, Path=/account-access-consents`,
		},
	}

	mockChecker := conditionalityCheckerMock{}

	for _, testCaseEntry := range testCases {
		// See: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		// for why we need this. Basically because we are running the tests in parallel using `t.Parallel`
		// we cannot bind to `testCaseEntry` as  there is a very good chance that when you run this code
		// you will see the last element being used all the time.
		func(testCase invalidTestCase) {
			t.Run(testCase.name, func(t *testing.T) {
				assert := assert.New(t)

				discoveryModel, err := FromJSONString(mockChecker, testCase.config)
				// fmt.Println()
				// fmt.Printf("%+v", err)
				// fmt.Println()

				assert.Nil(discoveryModel)
				assert.EqualError(err, testCase.expectedErr)
			})
		}(testCaseEntry)
	}
}

func TestDiscovery_FromJSONString_Valid(t *testing.T) {
	assert := assert.New(t)

	discoveryExample, err := ioutil.ReadFile("../../docs/discovery-example.json")
	assert.NoError(err)
	assert.NotNil(discoveryExample)
	config := string(discoveryExample)

	accountAPIDiscoveryItem := ModelDiscoveryItem{
		APISpecification: ModelAPISpecification{
			Name:          "Account and Transaction API Specification",
			URL:           "https://openbanking.atlassian.net/wiki/spaces/DZ/pages/642090641/Account+and+Transaction+API+Specification+-+v3.0",
			Version:       "v3.0",
			SchemaVersion: "https://raw.githubusercontent.com/OpenBankingUK/read-write-api-specs/v3.0.0/dist/account-info-swagger.json",
		},
		OpenidConfigurationURI: "https://as.aspsp.ob.forgerock.financial/oauth2/.well-known/openid-configuration",
		ResourceBaseURI:        "https://rs.aspsp.ob.forgerock.financial:443/",
		Endpoints: []ModelEndpoint{
			ModelEndpoint{
				Method:                "POST",
				Path:                  "/account-access-consents",
				ConditionalProperties: []ModelConditionalProperties(nil),
			},
			ModelEndpoint{
				Method:                "GET",
				Path:                  "/account-access-consents/{ConsentId}",
				ConditionalProperties: []ModelConditionalProperties(nil),
			},
			ModelEndpoint{Method: "DELETE",
				Path:                  "/account-access-consents/{ConsentId}",
				ConditionalProperties: []ModelConditionalProperties(nil),
			},
			ModelEndpoint{Method: "GET",
				Path:                  "/accounts/{AccountId}/product",
				ConditionalProperties: []ModelConditionalProperties(nil),
			},
			ModelEndpoint{Method: "GET",
				Path: "/accounts/{AccountId}/transactions",
				ConditionalProperties: []ModelConditionalProperties{
					ModelConditionalProperties{
						Schema:   "OBTransaction3Detail",
						Property: "Balance",
						Path:     "Data.Transaction[*].Balance",
					},
					ModelConditionalProperties{
						Schema:   "OBTransaction3Detail",
						Property: "MerchantDetails",
						Path:     "Data.Transaction[*].MerchantDetails",
					},
					ModelConditionalProperties{
						Schema:   "OBTransaction3Basic",
						Property: "TransactionReference",
						Path:     "Data.Transaction[*].TransactionReference",
					},
					ModelConditionalProperties{
						Schema:   "OBTransaction3Detail",
						Property: "TransactionReference",
						Path:     "Data.Transaction[*].TransactionReference",
					},
				},
			},
			ModelEndpoint{
				Method:                "GET",
				Path:                  "/accounts",
				ConditionalProperties: []ModelConditionalProperties(nil),
			},
			ModelEndpoint{
				Method:                "GET",
				Path:                  "/accounts/{AccountId}",
				ConditionalProperties: []ModelConditionalProperties(nil),
			},
			ModelEndpoint{
				Method:                "GET",
				Path:                  "/accounts/{AccountId}/balances",
				ConditionalProperties: []ModelConditionalProperties(nil),
			},
		},
	}

	modelActual, err := FromJSONString(model.NewConditionalityChecker(), config)
	assert.NoError(err)
	assert.NotNil(modelActual.DiscoveryModel)
	discoveryModel := modelActual.DiscoveryModel

	t.Run("model has a version", func(t *testing.T) {
		assert.Equal(discoveryModel.Version, "v0.0.1")
	})

	t.Run("model has correct number of discovery items", func(t *testing.T) {
		assert.Equal(len(discoveryModel.DiscoveryItems), 2)
	})

	t.Run("model has correct discovery item contents", func(t *testing.T) {
		assert.Equal(accountAPIDiscoveryItem, discoveryModel.DiscoveryItems[0])
	})
}

func TestDiscovery_Version(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(Version(), "v0.0.1")
}
