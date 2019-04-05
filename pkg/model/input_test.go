package model

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/authentication"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRequestEmptyEndpointOrMethod(t *testing.T) {
	i := &Input{}
	req, err := i.CreateRequest(emptyTestCase, emptyContext)
	assert.NotNil(t, err)
	assert.Nil(t, req)

	i = &Input{Endpoint: "http://google.com"}
	req, err = i.CreateRequest(emptyTestCase, emptyContext)
	assert.NotNil(t, err)
	assert.Nil(t, req)

	i = &Input{Method: "GET"}
	req, err = i.CreateRequest(emptyTestCase, emptyContext)
	assert.NotNil(t, err)
	assert.Nil(t, req)
}

// func TestInputGetValuesMissingContextVariable(t *testing.T) {
// 	match := Match{Description: "simple match test", ContextName: "GetValueToFind"}
// 	accessor := ContextAccessor{Matches: []Match{match}}
// 	i := &Input{Method: "GET", Endpoint: "http://google.com", ContextGet: accessor}
// 	req, err := i.CreateRequest(emptyTestCase, emptyContext)
// 	assert.NotNil(t, err)
// 	assert.Nil(t, req)
// }

func TestCreateRequestionNilContext(t *testing.T) {
	i := &Input{Method: "GET", Endpoint: "http://google.com"}
	req, err := i.CreateRequest(emptyTestCase, nil)
	assert.NotNil(t, err)
	assert.Nil(t, req)
}

func TestCreateRequestionNilTestcase(t *testing.T) {
	i := &Input{Method: "GET", Endpoint: "http://google.com"}
	req, err := i.CreateRequest(nil, emptyContext)
	assert.NotNil(t, err)
	assert.Nil(t, req)
}

func TestCreateRequestNilHeaderContext(t *testing.T) {
	headers := map[string]string{
		"Myheader": "myValue",
	}
	i := &Input{Method: "GET", Endpoint: "http://google.com", Headers: headers}
	req, err := i.CreateRequest(emptyTestCase, emptyContext)
	assert.Nil(t, err)
	assert.NotNil(t, req)
	for k, v := range req.Header {
		assert.Equal(t, "Myheader", k)
		assert.Equal(t, "myValue", v[0])
	}
}

func TestCreateRequestHeaderContext(t *testing.T) {
	headers := map[string]string{
		"Myheader": "$replacement",
	}
	ctx := Context{
		"replacement": "myNewValue",
	}
	i := &Input{Method: "GET", Endpoint: "http://google.com", Headers: headers}
	req, err := i.CreateRequest(emptyTestCase, &ctx)
	assert.Nil(t, err)
	assert.NotNil(t, req)
	for k, v := range req.Header {
		assert.Equal(t, "Myheader", k)
		assert.Equal(t, "myNewValue", v[0])
	}
}

func TestSetBearerAuthTokenFromContext(t *testing.T) {
	headers := map[string]string{
		"authorization": "Bearer $access_token",
	}
	ctx := Context{
		"access_token": "myShineyNewAccessTokenHotOffThePress",
	}
	i := &Input{Method: "GET", Endpoint: "http://google.com", Headers: headers}
	req, err := i.CreateRequest(emptyTestCase, &ctx)
	assert.Nil(t, err)
	assert.NotNil(t, req)
	for k, v := range req.Header {
		assert.Equal(t, "Authorization", k)
		assert.Equal(t, "Bearer myShineyNewAccessTokenHotOffThePress", v[0])
	}
}

func TestCreateHeaderContextMissingForReplacement(t *testing.T) {
	ctx := Context{
		"nomatch": "myNewValue",
	}
	result, err := replaceContextField("$replacement", &ctx)
	assert.NotNil(t, err)
	assert.Equal(t, "$replacement", result)

}

// func TestCheckAuthorizationTokenProcessed(t *testing.T) {
// 	m := Match{Description: "TokenProcessing", Authorisation: "Bearer"}
// 	tc := TestCase{Expect: Expect{Matches: []Match{m}, StatusCode: 200}}
// 	resp := test.CreateHTTPResponse(200, "OK", "TheRainInSpain", "Authorization", "Bearer 1010110101010101")
// 	result, err := tc.Validate(resp, emptyContext)
// 	assert.Equal(t, "1010110101010101", tc.Expect.Matches[0].Result)
// 	assert.Nil(t, err)
// 	assert.True(t, result)

// 	ctx := &Context{
// 		"access_token": "1010101010101010",
// 	}
// 	match := Match{Description: "test", ContextName: "access_token", Authorisation: "bearer"}
// 	accessor := ContextAccessor{Matches: []Match{match}}
// 	i := &Input{Method: "GET", Endpoint: "http://google.com", ContextGet: accessor}
// 	req, err := i.CreateRequest(emptyTestCase, ctx)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, req)
// }

func TestFormData(t *testing.T) {
	i := Input{Endpoint: "/accounts", Method: "POST", FormData: map[string]string{
		"grant_type": "client_credentials",
		"scope":      "accounts openid"}}
	ctx := Context{"baseurl": "http://mybaseurl"}
	tc := TestCase{Input: i, Context: ctx}
	req, err := tc.Prepare(emptyContext)
	assert.Nil(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, 2, len(req.FormData))
}

func TestFormDataMissingContextVariable(t *testing.T) {
	i := Input{Endpoint: "/accounts", Method: "POST", FormData: map[string]string{
		"grant_type": "$client_credentials",
		"scope":      "accounts openid"}}
	ctx := Context{"baseurl": "http://mybaseurl"}
	tc := TestCase{Input: i, Context: ctx}
	req, err := tc.Prepare(emptyContext)
	assert.NotNil(t, err)
	assert.Nil(t, req)
}

func TestInputBody(t *testing.T) {
	i := Input{Endpoint: "/accounts", Method: "POST", RequestBody: "The Rain in Spain Falls Mainly on the Plain"}
	ctx := Context{"baseurl": "http://mybaseurl"}
	tc := TestCase{Input: i, Context: ctx}
	req, err := tc.Prepare(emptyContext)
	assert.Nil(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "The Rain in Spain Falls Mainly on the Plain", req.Body.(string))
}

func TestInputClaims(t *testing.T) {
	i := Input{Endpoint: "/accounts", Method: "POST",
		Generation: map[string]string{
			"strategy": "consenturl",
		},
		Claims: map[string]string{
			"iss":          "8672384e-9a33-439f-8924-67bb14340d71",
			"scope":        "openid accounts",
			"redirect_url": "https://test.example.co.uk/redir",
			"responseType": "code",
		}}
	ctx := Context{"baseurl": "http://mybaseurl"}
	tc := TestCase{Input: i, Context: ctx}
	req, err := tc.Prepare(emptyContext)
	assert.Nil(t, err)
	assert.NotNil(t, req)

	m, err := url.ParseQuery(req.URL)
	require.NoError(t, err)
	assert.Equal(t, m["request"][0], "eyJhbGciOiJub25lIn0.eyJhdWQiOiIiLCJjbGFpbXMiOnsiaWRfdG9rZW4iOnsib3BlbmJhbmtpbmdfaW50ZW50X2lkIjp7ImVzc2VudGlhbCI6dHJ1ZSwidmFsdWUiOiIifX19LCJpc3MiOiI4NjcyMzg0ZS05YTMzLTQzOWYtODkyNC02N2JiMTQzNDBkNzEiLCJyZWRpcmVjdF91cmkiOiJodHRwczovL3Rlc3QuZXhhbXBsZS5jby51ay9yZWRpciIsInNjb3BlIjoib3BlbmlkIGFjY291bnRzIn0.")
}

func TestInputClaimsWithContextReplacementParameters(t *testing.T) {
	i := Input{Endpoint: "/accounts", Method: "POST",
		Generation: map[string]string{
			"strategy": "consenturl",
		},
		Claims: map[string]string{
			"aud":          "$baseurl",
			"iss":          "8672384e-9a33-439f-8924-67bb14340d71",
			"scope":        "openid accounts",
			"redirect_url": "https://test.example.co.uk/redir",
			"consentId":    "$consent_id",
			"responseType": "code",
		}}
	ctx := Context{"baseurl": "http://mybaseurl", "consent_id": "myconsentid"}
	tc := TestCase{Input: i, Context: ctx}
	req, err := tc.Prepare(emptyContext)
	assert.Nil(t, err)
	assert.NotNil(t, req)

	m, err := url.ParseQuery(req.URL)
	require.NoError(t, err)
	assert.Equal(t, m["request"][0], "eyJhbGciOiJub25lIn0.eyJhdWQiOiJodHRwOi8vbXliYXNldXJsIiwiY2xhaW1zIjp7ImlkX3Rva2VuIjp7Im9wZW5iYW5raW5nX2ludGVudF9pZCI6eyJlc3NlbnRpYWwiOnRydWUsInZhbHVlIjoibXljb25zZW50aWQifX19LCJpc3MiOiI4NjcyMzg0ZS05YTMzLTQzOWYtODkyNC02N2JiMTQzNDBkNzEiLCJyZWRpcmVjdF91cmkiOiJodHRwczovL3Rlc3QuZXhhbXBsZS5jby51ay9yZWRpciIsInNjb3BlIjoib3BlbmlkIGFjY291bnRzIn0.")

}

func TestInputClaimsConsentId(t *testing.T) {
	ctx := Context{"consent_id": "aac-fee2b8eb-ce1b-48f1-af7f-dc8f576d53dc", "xchange_code": "10e9d80b-10d4-4abd-9fe0-15789cc512b5", "baseurl": "https://modelobankauth2018.o3bank.co.uk:4101", "access_token": "18d5a754-0b76-4a8f-9c68-dc5caaf812e2"}
	i := Input{Endpoint: "/accounts", Method: "POST",
		Generation: map[string]string{
			"strategy": "consenturl",
		},
		Claims: map[string]string{
			"aud":          "$baseurl",
			"iss":          "8672384e-9a33-439f-8924-67bb14340d71",
			"scope":        "openid accounts",
			"redirect_url": "https://test.example.co.uk/redir",
			"consentId":    "$consent_id",
			"responseType": "code",
		}}
	tc := TestCase{Input: i, Context: ctx}
	res, err := i.CreateRequest(&tc, &ctx)
	assert.NoError(t, err, "create request should succeed")
	assert.NotNil(t, res)
}

func TestClaimsJWTBearer(t *testing.T) {
	cert, err := authentication.NewCertificate(selfsignedDummypub, selfsignedDummykey)
	require.NoError(t, err)
	ctx := Context{
		"consent_id":   "aac-fee2b8eb-ce1b-48f1-af7f-dc8f576d53dc",
		"xchange_code": "10e9d80b-10d4-4abd-9fe0-15789cc512b5",
		"baseurl":      "https://matls-sso.openbankingtest.org.uk",
		"access_token": "18d5a754-0b76-4a8f-9c68-dc5caaf812e2",
		"client_id":    "12312",
		"scope":        "AuthoritiesReadAccess ASPSPReadAccess TPPReadAll",
		"SigningCert":  cert,
	}
	i := Input{Endpoint: "/as/token.oauth2", Method: "POST",
		Generation: map[string]string{
			"strategy": "jwt-bearer",
		},
		Claims: map[string]string{
			"aud":   "$baseurl",
			"iss":   "$client_id",
			"scope": "AuthoritiesReadAccess ASPSPReadAccess TPPReadAll",
		},
		FormData: map[string]string{
			"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
			"grant_type":            "client_credentials",
			"client_id":             "$client_id",
			"scope":                 "$scope",
		},
	}
	tc := TestCase{Input: i, Context: ctx}
	res, err := i.CreateRequest(&tc, &ctx)
	require.NoError(t, err, "create request should succeed")
	assert.NotNil(t, res)
	jwtbearer, exists := ctx.Get("jwtbearer")
	assert.True(t, exists)
	assert.True(t, len(jwtbearer.(string)) > 20)
}

func TestJWTSignRS256(t *testing.T) {
	cert, err := authentication.NewCertificate(selfsignedDummypub, selfsignedDummykey)
	require.NoError(t, err)
	require.NotNil(t, cert)

	alg := jwt.GetSigningMethod("RS256")
	if alg == nil {
		fmt.Printf("Couldn't find signing method: %v\n", alg)
	}
	claims := jwt.MapClaims{}
	claims["iat"] = time.Now().Unix()
	token := jwt.NewWithClaims(alg, claims) // create new token
	token.Header["kid"] = "mykeyid"
	prikey := cert.PrivateKey()
	tokenString, err := token.SignedString(prikey) // sign the token - get as encoded string
	if err != nil {
		fmt.Println("error signing jwt: ", err)
	}
	assert.True(t, len(tokenString) > 30)

}

func TestBodyLiteral(t *testing.T) {
	ctx := Context{
		"replacebody": "this is my body",
	}

	i := Input{Method: "POST", Endpoint: "https://google.com", RequestBody: "This is my literal body"}
	tc := TestCase{Input: i}
	req, err := tc.Prepare(&ctx)
	assert.Nil(t, err)
	assert.Equal(t, "This is my literal body", req.Body)
}

func TestBodyReplacement(t *testing.T) {
	ctx := Context{
		"replacebody": "this is my body",
	}

	i := Input{Method: "POST", Endpoint: "https://google.com", RequestBody: "$replacebody"}
	tc := TestCase{Input: i}
	req, err := tc.Prepare(&ctx)
	assert.Nil(t, err)
	assert.Equal(t, "this is my body", req.Body)
}

func TestBodyTwoReplacements(t *testing.T) {
	ctx := Context{
		"replacebody": "this is my body",
		"replace2":    "and this is my heart",
	}

	i := Input{Method: "POST", Endpoint: "https://google.com", RequestBody: "$replacebody $replace2"}
	tc := TestCase{Input: i}
	req, err := tc.Prepare(&ctx)
	assert.Nil(t, err)
	assert.Equal(t, "this is my body and this is my heart", req.Body)
}

func TestPaymentBodyReplace(t *testing.T) {
	ctx := Context{
		"initiation":                "{\"InstructionIdentification\":\"SIDP01\",\"EndToEndIdentification\":\"FRESCO.21302.GFX.20\",\"InstructedAmount\":{\"Amount\":\"15.00\",\"Currency\":\"GBP\"},\"CreditorAccount\":{\"SchemeName\":\"SortCodeAccountNumber\",\"Identification\":\"20000319470104\",\"Name\":\"Messers Simplex & Co\"}}",
		"consent_id":                "sdp-1-b5bbdb18-eeb1-4c11-919d-9a237c8f1c7d",
		"domestic_payment_template": "{\"Data\": {\"ConsentId\": \"$consent_id\",\"Initiation\":$initiation },\"Risk\":{}}",
	}

	i := Input{Method: "POST", Endpoint: "https://google.com", RequestBody: "$domestic_payment_template"}
	tc := TestCase{Input: i}
	req, err := tc.Prepare(&ctx)
	assert.Nil(t, err)
	assert.Equal(t, "{\"Data\": {\"ConsentId\": \"sdp-1-b5bbdb18-eeb1-4c11-919d-9a237c8f1c7d\",\"Initiation\":{\"InstructionIdentification\":\"SIDP01\",\"EndToEndIdentification\":\"FRESCO.21302.GFX.20\",\"InstructedAmount\":{\"Amount\":\"15.00\",\"Currency\":\"GBP\"},\"CreditorAccount\":{\"SchemeName\":\"SortCodeAccountNumber\",\"Identification\":\"20000319470104\",\"Name\":\"Messers Simplex & Co\"}} },\"Risk\":{}}", req.Body)
}

func TestPaymentBodyReplaceTestCase100300(t *testing.T) {
	ctx := Context{
		"x-fapi-financial-id": "myfapiid",
		"thisSchemeName":      "myscheme",
		"thisIdentification":  "myid",
	}
	_ = ctx
	var tc TestCase
	err := json.Unmarshal(paymentTestCaseData100300, &tc)
	assert.Nil(t, err)
	fmt.Printf("%#v\n", tc)
	ctx.PutContext(&tc.Context)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.TraceLevel)
	ctx.DumpContext("Testcase")
	req, err := tc.Prepare(&ctx)

	assert.Nil(t, err)
	_ = req
	fmt.Printf("%#v\n", tc)

}

var paymentTestCaseData100300 = []byte(`
{
    "@id": "OB-301-DOP-100300",
    "name": "Domestic Payment consents succeeds with minimal data set with additional schema checks.",
    "purpose": "Check that the resource succeeds posting a domestic payment consents with a minimal data set and checks additional schema.",
    "input": {
        "method": "POST",
        "endpoint": "/domestic-payment-consents",
        "headers": {
            "Content-Type": "application/json; charset=utf-8",
            "x-fapi-financial-id": "$x-fapi-financial-id",
            "x-fapi-interaction-id": "b4405450-febe-11e8-80a5-0fcebb1574e1",
            "x-fcs-testcase-id": "OB-301-DOP-100300"
        },
        "bodyData": "{\n    \"Data\": {\n        \"Initiation\": {\n            \"CreditorAccount\": {\n                \"Identification\": \"$thisIdentification\",\n                \"Name\": \"CF Tool\",\n                \"SchemeName\": \"$thisSchemeName\"\n            },\n            \"EndToEndIdentification\": \"$thisInstructionIdentification\",\n            \"InstructedAmount\": {\n                \"Amount\": \"1.00\",\n                \"Currency\": \"$thisCurrency\"\n            },\n            \"InstructionIdentification\": \"$thisInstructionIdentification\"\n        }\n    },\n    \"Risk\": {}\n}"
    },
    "context": {
        "baseurl": "http://mybaseurl",
        "requestConsent": "true",
        "thisCurrency": "GBP",
        "thisInstructionIdentification": "OB-301-DOP-100300",
        "tokenScope": "payments",
        "x-fapi-financial-id": "$x-fapi-financial-id"
    },
    "expect": {
        "status-code": 201,
        "schema-validation": true,
        "matches": [
            {
                "header-present": "x-fapi-interaction-id"
            },
            {
                "json": "Data.Status",
                "value": "AwaitingAuthorisation"
            },
            {
                "json": "Data.ConsentId"
            }
        ],
        "contextPut": {
            "matches": [
                {
                    "name": "OB-301-DOP-100300-ConsentId",
                    "json": "Data.ConsentId"
                }
            ]
        }
    }
}`)

var selfsignedDummykey = `-----BEGIN RSA PRIVATE KEY----- 
MIIEpAIBAAKCAQEA8Gl2x9KsmqwdmZd+BdZYtDWHNRXtPd/kwiR6luU+4w76T+9m
lmePXqALi7aSyvYQDLeffR8+2dSGcdwvkf6bDWZNeMRXl7Z1jsk+xFN91mSYNk1n
R6N1EsDTK2KXlZZyaTmpu/5p8SxwDO34uE5AaeESeM3RVqqOgRcXskmp/atwUMC+
qLfuXPoNSjKguWdcuZwJxbjJbqbvF5/zXISEoKly5iGK+11eDRcX2Rp8yRpOhO84
LtSpC21QTkMSK8VA4O3e1tOW+DXaJb3QtzwocTb+wXTw74agwvcAQP9yDilgR1t6
fdqeGrW4Y25bDulXlsD+S6PzhVo+EVPq1T3RJQIDAQABAoIBADZfQ9Pxm8PnhVJF
ZuUfEzS+nnOtH9jMmEooQel6s3xa2NXXSRZfGZfHDpVsl0p72CloJhQASxCs9jMu
HzwfnyWqq37SuRTA2VmPvjhcwasJWTt+ygrztvikz52SUMIuInYV6oNwCLnY2Qaz
k3rrh7nqg2j684tsS4p6lItoCaArA5xcQwxn6librK/NzHzLaXN0zLufn4WYuPMc
3NTuZWY6EYqbyHbuiwrsZGin8JKw+bqfG6OEtt5GVJbzmacjQrVTEnEcJNO8Pe3H
bC/ZczFBb9Vsznfp0EICKf/OZVen7zSZ58+l3zg/+h0A8Z1D4jbWkXWDS3dYiAQU
g2C9x8kCgYEA/sllVEZXyCduyUvP7nVYPasBHKbIuS8G/cIKfwy+Wd1ZhKg4JgIy
5nhERYfOJeDwSoQUYxJSZoCgnByc4jx3kSX4oyTdKdT+yj1Sma38GONRm3Erxany
aZvw40cj5vCn8iGl6hsSpqWWiHWizEyO+XvctfMaFq5vOQxgjTF4Yw8CgYEA8Y6L
VlZxByVO7kQwZXdJ1zEtu1YzZyw95kiHmnxHdOqhstDV9mwtcDLD12CYR2LVweT/
ndTTU4U3q1z/Zuo1t1HYvTHU4ss/4GBCskO0JIKFPDr2KdfUiDGn6eMWNmoQ35zU
Z2zfi3BMtX2dbobX+dBDyh+ZJf21zKsyujp/eIsCgYEA6poU9IGE6KbuiumEv6RL
KRVhg7lLD8Dupg/azFu2llaLy+t9L/pMVgydiIxg1F4HxAVUJFlFiF6eBMEP7/0P
d5ZIGCikgJVAOoY2nY0nmN8PUJrnXC19KaNOLmhd9ZLYgcpb1HEzPkEwl9wBmC5S
ZASaGOuMtR/PB++Oo9PObx8CgYEAjwcdH+kdEeMYYmKD2YCRe1bGQlefJib/G9y0
VlfiI6tORUf8eOXC3d1hMqUiZZpzAVTrufOrkZeex9vP6oshdUOENzpLWGKKlvvI
Yi9OehPCelBbM5l1YZMtXoK0w1F4Xj9JUVgY4UKEWS5gynITbfrQON0O3HzmaaKw
7a33jlMCgYBXA5OvGOav/uMQ45SiYf8KhyNx2gQqQlvSeF9z8yDqmHr2epoAVoEu
nuz0bzQ7ARNqBkWLQ4bqzwy0aKXlcvbIMBaVXyfQTiwzwWAZsAWr6WHPlvWDP6mP
1vAUje5xEMtsIwj6UnkJ3OpPVVeJ56aKQIxg6QU2ROrWYDccx4gg0g==
-----END RSA PRIVATE KEY-----`
var selfsignedDummypub = `-----BEGIN CERTIFICATE-----
MIIDBzCCAe+gAwIBAgIJAOze8GNkMIMMMA0GCSqGSIb3DQEBBQUAMBoxGDAWBgNV
BAMMD3d3dy5leGFtcGxlLmNvbTAeFw0xOTAxMjExMzUyMjFaFw0yOTAxMTgxMzUy
MjFaMBoxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBAPBpdsfSrJqsHZmXfgXWWLQ1hzUV7T3f5MIkepblPuMO
+k/vZpZnj16gC4u2ksr2EAy3n30fPtnUhnHcL5H+mw1mTXjEV5e2dY7JPsRTfdZk
mDZNZ0ejdRLA0ytil5WWcmk5qbv+afEscAzt+LhOQGnhEnjN0VaqjoEXF7JJqf2r
cFDAvqi37lz6DUoyoLlnXLmcCcW4yW6m7xef81yEhKCpcuYhivtdXg0XF9kafMka
ToTvOC7UqQttUE5DEivFQODt3tbTlvg12iW90Lc8KHE2/sF08O+GoML3AED/cg4p
YEdben3anhq1uGNuWw7pV5bA/kuj84VaPhFT6tU90SUCAwEAAaNQME4wHQYDVR0O
BBYEFG3WDJMv5wa4QvWwxcJpNU/RTBp/MB8GA1UdIwQYMBaAFG3WDJMv5wa4QvWw
xcJpNU/RTBp/MAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEFBQADggEBALDYGYv6
0KoSAbQamSOZT6h2LBJj/AbGV+W9ffUDW6OuJ1C7sa7sDki2HQgz7vfS0BtrKY/q
tszfZqWKlx8AFbLhusMcv3gc6Dv4Onod90EaIKuRU+sElo1BEak5asY4oHru5GIK
QxGi8GkcwKSwnxSrkKQz8xXcL+P3daOmaAUQDo6JPqxYE4DNsQ3HRtkCj9kTUk8+
ppJAzXoBrutQz7e2daEXHUNc+1+KcD+se5cmvK2cJg6vk1vpgY1kjXdLQr1CySxJ
XgfLm2jJfzMF/L5RX5Vdnon6x4ufi7e/3fOThjlhLRXMOkhlb0E+wSYP0NvLA12E
rjs761ndZ9Qrb0s=
-----END CERTIFICATE-----
`
