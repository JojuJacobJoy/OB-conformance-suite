package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/discovery"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"
)

func TestGenerateTestCases(t *testing.T) {
	apiSpec := discovery.ModelAPISpecification{
		SchemaVersion: accountSwaggerLocation31,
	}
	tests, err := GenerateTestCases(apiSpec, "http://mybaseurl", &model.Context{}, readDiscovery())
	assert.Nil(t, err)

	perms, err := getAccountPermissions(tests)
	assert.Nil(t, err)
	m := map[string]string{}
	for _, v := range perms {
		t.Logf("perms: %s %-50.50s %s\n", v.ID, v.Path, v.Permissions)
		m[v.Path] = v.ID
	}
	requiredTokens, err := GetRequiredTokensFromTests(tests, "accounts")
	for _, v := range requiredTokens {
		fmt.Println(v)
	}
}

func TestPaymentPermissions(t *testing.T) {
	apiSpec := discovery.ModelAPISpecification{
		SchemaVersion: accountSwaggerLocation30,
	}
	tests, err := GenerateTestCases(apiSpec, "http://mybaseurl", &model.Context{}, readDiscovery())
	fmt.Printf("we have %d tests\n", len(tests))
	for _, v := range tests {
		dumpJSON(v)
	}

	requiredTokens, err := GetPaymentPermissions(tests)
	assert.Nil(t, err)

	for _, v := range requiredTokens {
		fmt.Printf("%#v\n", v)
	}

	fmt.Println("where are my tests?")
}

func TestDataReferencesAndDump(t *testing.T) {
	data, err := loadAssert()
	assert.Nil(t, err)

	for k, v := range data.References {
		body := jsonString(v.Body)
		l := len(body)
		if l > 0 {
			v.BodyData = body
			v.Body = ""
			data.References[k] = v
		}
	}
}

func loadAssert() (References, error) {
	refs, err := loadReferences("../../manifests/data.json")

	if err != nil {
		fmt.Println("what the hell is going on " + err.Error())
		refs, err = loadReferences("manifesxts/data.json")
		if err != nil {
			fmt.Println("what the hell is going on " + err.Error())
			return References{}, err
		}
	}

	for k, v := range refs.References { // read in data references with body payloads
		body := jsonString(v.Body)
		l := len(body)
		if l > 0 {
			v.BodyData = body
			v.Body = ""
			refs.References[k] = v
		}
	}
	dumpJSON(refs)
	return refs, err
}

func TestPermissionFiteringAccounts(t *testing.T) {

	ctx := model.Context{
		"accountId":           "123123123",
		"client_access_token": "abc-defg-hijk-lmno-pqrs",
	}

	endpoints := readDiscovery()
	apiSpec := discovery.ModelAPISpecification{
		SchemaVersion: accountSwaggerLocation31,
	}
	tests, err := GenerateTestCases(apiSpec, "http://mybaseurl", &ctx, endpoints)
	assert.Nil(t, err)
	fmt.Printf("%d tests loaded", len(tests))

	scripts, _, err := loadGenerationResources("accounts")
	if err != nil {
		fmt.Println("Error on loadGenerationResources")
		return
	}

	filteredScripts, err := filterTestsBasedOnDiscoveryEndpointsPlayground(scripts, endpoints)
	if err != nil {

	}
	for _, v := range filteredScripts.Scripts {
		dumpJSON(v)
	}
}

func readDiscovery() []discovery.ModelEndpoint {
	discoveryJSON, err := ioutil.ReadFile("../discovery/templates/ob-v3.1-ozone.json")
	if err != nil {
		fmt.Println("discovery read failed")
		return nil
	}

	disco := &discovery.Model{}

	err = json.Unmarshal(discoveryJSON, &disco)

	return disco.DiscoveryModel.DiscoveryItems[0].Endpoints

}

func filterTestsBasedOnDiscoveryEndpointsPlayground(scripts Scripts, endpoints []discovery.ModelEndpoint) (Scripts, error) {

	lookupMap := make(map[string]bool)
	_ = lookupMap
	filteredScripts := []Script{}
	fmt.Println("***Discovery Endpoint URLs")

	for _, ep := range endpoints {
		for _, regpath := range accountsRegex {
			matched, err := regexp.MatchString(regpath.Regex, ep.Path)
			if err != nil {
				continue
			}
			if matched {
				lookupMap[regpath.Regex] = true
				fmt.Printf("endpoint %40.40s matched by regex %42.42s: %s\n", ep.Path, regpath.Regex, regpath.Name)
			}
		}
	}
	fmt.Println("***Script URLs")
	for _, scr := range scripts.Scripts {
		for _, regpath := range accountsRegex {
			stripped := strings.Replace(scr.URI, "$", "", -1) // only works with a single character
			matched, err := regexp.MatchString(regpath.Regex, stripped)
			if err != nil {
				fmt.Printf("matching err " + err.Error())
				continue
			}
			if matched {
				fmt.Printf("%40.40s matched by regex %42.42s: %s\n", scr.URI, regpath.Regex, regpath.Name)
			} else {
				//fmt.Printf("No match %s\n", scr.URI)
			}
		}
	}

	fmt.Println("***dmp")
	for k := range lookupMap {
		fmt.Printf("lookup map %s\n", k)

	}

	fmt.Println("***Lookup Map")
	for k := range lookupMap {
		for i, scr := range scripts.Scripts {
			stripped := strings.Replace(scr.URI, "$", "", -1) // only works with a single character
			matched, err := regexp.MatchString(k, stripped)
			if err != nil {
				continue
			}
			if matched {
				fmt.Printf("endpoint %40.40s matched by regex %42.42s\n", scr.URI, k)
				filteredScripts = append(filteredScripts, scripts.Scripts[i])
			}
		}
	}
	myscripts := Scripts{Scripts: filteredScripts}

	return myscripts, nil
}

func TestPaymentTestCaseCreation(t *testing.T) {
	ctx := &model.Context{
		"consent_id":                          "aac-fee2b8eb-ce1b-48f1-af7f-dc8f576d53dc",
		"xchange_code":                        "10e9d80b-10d4-4abd-9fe0-15789cc512b5",
		"baseurl":                             "https://matls-sso.openbankingtest.org.uk",
		"access_token":                        "18d5a754-0b76-4a8f-9c68-dc5caaf812e2",
		"client_id":                           "12312",
		"scope":                               "AuthoritiesReadAccess ASPSPReadAccess TPPReadAll",
		"authorisation_endpoint":              "https://example.com/authorisation",
		"OB-301-DOP-100300-ConsentId":         "100100-ConsentId",
		"OB-301-DOP-100600-DomesticPaymentId": "100600-DomesticPaymentId-PaymentId",
		"OB-301-DOP-100100-ConsentId":         "100100-ConsentId",
		"OB-301-DOP-100800-ConsentId":         "100800-Consentid",
		"creditorIdentification":              "1231231231",
		"thisCurrency":                        "GBP",
		"creditorScheme":                      "default",
	}

	apiSpec := discovery.ModelAPISpecification{
		SchemaVersion: accountSwaggerLocation31,
	}

	tests, err := GenerateTestCases(apiSpec, "http://mybaseurl", ctx, readDiscovery())
	assert.Nil(t, err)
	fmt.Printf("we have %d tests\n", len(tests))
	for _, v := range tests {
		//if v.ID == "OB-301-DOP-101000" {
		dumpJSON(v)
		//}
	}

}
