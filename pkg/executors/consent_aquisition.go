package executors

import (
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/generation"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/manifest"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	resty "gopkg.in/resty.v1"
)

var consentChannelTimeout = 30

// GetPsuConsent -
func GetPsuConsent(definition RunDefinition, ctx *model.Context, runTests *generation.TestCasesRun, permissions map[string][]manifest.RequiredTokens) (TokenConsentIDs, map[string]string, error) {
	consentRequirements := runTests.SpecConsentRequirements
	var consentIdsToReturn TokenConsentIDs
	logrus.Debugf("running with %#v\n", permissions)

	for specType := range permissions {
		logrus.Tracef("Getting PSU Consent for api type: %s\n", specType)
		tests, err := getSpecForSpecType(specType, runTests)
		if err != nil {
			return nil, nil, err
		}

		switch specType {
		case "accounts":
			consentIds, _, err := getAccountConsents(consentRequirements, definition, permissions["accounts"], ctx)
			consentIdsToReturn = append(consentIdsToReturn, consentIds...)
			if err != nil {
				logrus.Error("GetPSUConsent - accounts error: " + err.Error())
				return nil, nil, err
			}

		case "payments": //TODO: Handle multipe spec returns
			consentIds, err := getPaymentConsents(tests, definition, permissions["payments"], ctx)
			consentIdsToReturn = append(consentIdsToReturn, consentIds...)
			if err != nil {
				logrus.Error("GetPSUConsent - payments error: " + err.Error())
				return nil, nil, err
			}
		default:
			logrus.Fatalf("Support for spec type (%s) not implemented yet", specType)
		}
	}

	logrus.Warnf("No Consent Acquistion Performed\n")
	return consentIdsToReturn, nil, nil
}

func getSpecForSpecType(stype string, runTests *generation.TestCasesRun) ([]model.TestCase, error) {
	for _, spec := range runTests.TestCases {
		specType, err := manifest.GetSpecType(spec.Specification.SchemaVersion)
		if err != nil {
			logrus.Warnf("cannot get spec type from SchemaVersion %s\n", spec.Specification.SchemaVersion)
			return nil, errors.New("Cannot get spec type from SchemaVersion " + spec.Specification.SchemaVersion)
		}
		if stype == specType {
			return spec.TestCases, nil
		}
	}

	return nil, errors.New("Cannot find test cases for spec type " + stype)
}

// getAccountConsents - get required tokens
func getAccountConsents(consentRequirements []model.SpecConsentRequirements, definition RunDefinition, permissions []manifest.RequiredTokens, ctx *model.Context,
) (TokenConsentIDs, map[string]string, error) {

	consentIDChannel := make(chan TokenConsentIDItem, 100)
	logger := logrus.StandardLogger().WithField("module", "getAccountConsents")
	logger.Tracef("getAccountConsents")

	tokenParameters := make(map[string]string, 0)

	//requiredTokens, err := manifest.GetRequiredTokensFromTests(spec.TestCases, "accounts")
	requiredTokens := permissions
	logrus.Tracef("we require %d tokens for `accounts`", len(requiredTokens))
	logrus.Tracef("required tokens %#v\n", requiredTokens)
	for _, rt := range requiredTokens {
		tokenParameters[rt.Name] = buildPermissionString(rt.Perms)
	}
	logrus.Tracef("required tokens %#v\n", tokenParameters)
	//for tokenName, permissionList := range tokenParameters {
	for _, rt := range requiredTokens {
		permissionList := rt.Perms
		tokenName := rt.Name
		runner := NewConsentAcquisitionRunner(logrus.StandardLogger().WithField("module", "InitiationConsentAcquisition"), definition, NewBufferedDaemonController())
		tokenAcquisitionType := definition.DiscoModel.DiscoveryModel.TokenAcquisition
		permissionString := buildPermissionString(permissionList)
		consentInfo := TokenConsentIDItem{TokenName: tokenName, Permissions: permissionString}
		err := runner.RunConsentAcquisition(consentInfo, ctx, tokenAcquisitionType, consentIDChannel)
		if err != nil {
			logger.WithError(err).Debug("InitiationConsentAcquisition")
		}
	}

	consentItems, err := waitForConsentIDs(consentIDChannel, len(tokenParameters))
	for _, v := range consentItems {
		logger.Debugf("Setting Token: %s, ConsentId: %s", v.TokenName, v.ConsentID)
		ctx.PutString(v.TokenName, v.ConsentID)
	}
	logrus.Debugf("we have %d consentIds: %#v\n", len(consentItems), consentItems)
	return consentItems, tokenParameters, err
}

func getTokenParametersFromRequiredTokens(tokens []manifest.RequiredTokens) map[string][]string {
	tokenParameters := map[string][]string{}
	for _, reqToken := range tokens {
		tokenParameters[reqToken.Name] = reqToken.Perms
	}
	return tokenParameters
}

func waitForConsentIDs(consentIDChannel chan TokenConsentIDItem, consentIDsRequired int) (TokenConsentIDs, error) {
	logger := logrus.StandardLogger().WithFields(logrus.Fields{
		"function":           "waitForConsentIDs",
		"consentIDsRequired": consentIDsRequired,
	})

	consentItems := TokenConsentIDs{}
	consentIDsReceived := 0
	logger.Debug("Waiting for items on consentIDChannel ...")
	for {
		select {
		case item := <-consentIDChannel:
			logrus.Debugf("received consent channel item item %#v", item)
			consentIDsReceived++
			consentItems = append(consentItems, item)

			logger.WithFields(logrus.Fields{
				"consentIDsReceived": consentIDsReceived,
				"consentIDsRequired": consentIDsRequired,
				"item":               fmt.Sprintf("%#v", item),
			}).Info("Progressing ...")

			errs := errors.New("")
			if consentIDsReceived == consentIDsRequired {
				logrus.Infof("Got %d required tokens - progressing..", consentIDsReceived)
				for _, v := range consentItems {
					if len(v.Error) > 0 {
						errs = errors.WithMessage(errs, v.Error)
					} else {
						logrus.Infof("token: %s, consentid: %s", v.TokenName, v.ConsentID)
					}
				}

				if len(errs.Error()) > 0 {
					return consentItems, errs
				}
				return consentItems, nil
			}
		case <-time.After(time.Duration(consentChannelTimeout) * time.Second):
			logrus.Warnf("consent channel timeout after %d seconds", consentChannelTimeout)
			return consentItems, errors.New("ConsentChannel Timeout")
		}
	}
}

func getConsentTokensAndPermissions(consentRequirements []model.SpecConsentRequirements, logger *logrus.Entry) map[string][]string {
	tokenParameters := make(map[string][]string)
	for _, v := range consentRequirements {
		for _, namedPermission := range v.NamedPermissions {
			codeset := namedPermission.CodeSet
			for _, b := range codeset.CodeSet {
				mystring := string(b)
				set := tokenParameters[namedPermission.Name]
				set = append(set, mystring)
				tokenParameters[namedPermission.Name] = set
			}
		}
	}

	for tokenParameterKey, tokenParameterValue := range tokenParameters {
		logger.WithFields(logrus.Fields{
			"tokenParameterValue":   tokenParameterValue,
			"tokenParameterKey":     tokenParameterKey,
			"buildPermissionString": buildPermissionString(tokenParameterValue),
		}).Debugf("Getting ConsentToken")
	}
	return tokenParameters
}

func buildPermissionString(permissionSlice []string) string {
	var permissions string
	first := true
	for _, perms := range permissionSlice {
		if !first {
			permissions += ","
		} else {
			first = !first
		}
		permissions += "\"" + perms + "\""
	}
	return permissions
}

// ExchangeParameters - Captures the parameters require to exchange a code for an access token
type ExchangeParameters struct {
	Code                string
	BasicAuthentication string
	TokenEndpoint       string
	RedirectURL         string
	Scope               string
	TokenName           string
}

// ExchangeCodeForAccessToken - runs a testcase to perform this operation
func ExchangeCodeForAccessToken(tokenName, code, scope string, definition RunDefinition, ctx *model.Context) (accesstoken string, err error) {
	logger := logrus.StandardLogger().WithField("module", "ExchangeCodeForAccessToken")
	logger.WithFields(logrus.Fields{
		"code":      code,
		"tokenName": tokenName,
	}).Debug("Exchanging code for token")
	grantToken, err := exchangeCodeForToken(code, scope, ctx, logger)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"code":      code,
			"tokenName": tokenName,
			"err":       err,
		}).Error("Exchanging code for token failed")
	}
	return grantToken.AccessToken, err
}

type grantToken struct {
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	Expires     int32  `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
	IDToken     string `json:"id_token,omitempty"`
}

func exchangeCodeForToken(code, scope string, ctx *model.Context, logger *logrus.Entry) (grantToken, error) {
	logger = logger.WithFields(logrus.Fields{
		"function": "exchangeCodeForToken",
	})

	basicAuth, err := ctx.GetString("basic_authentication")
	if err != nil {
		return grantToken{}, errors.New("cannot get basic authentication for code")
	}
	tokenEndpoint, err := ctx.GetString("token_endpoint")
	if err != nil {
		return grantToken{}, errors.New("cannot get token_endpoint for code exchange")
	}
	redirectURL, err := ctx.GetString("redirect_url")
	if err != nil {
		return grantToken{}, errors.New("Cannot get redirectURL for code exchange")
	}
	if scope == "" {
		scope = "accounts" // lets default to a scope of accounts
	}

	logger.WithFields(logrus.Fields{
		"tokenEndpoint": tokenEndpoint,
	}).Debugf("Attempting POST")
	resp, err := resty.R().
		SetHeader("content-type", "application/x-www-form-urlencoded").
		SetHeader("accept", "*/*").
		SetHeader("authorization", "Basic "+basicAuth).
		SetFormData(map[string]string{
			"grant_type":   "authorization_code",
			//"scope":        scope, // accounts or payments currently
			"code":         code,
			"redirect_uri": redirectURL,
		}).
		Post(tokenEndpoint)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"tokenEndpoint": tokenEndpoint,
			"err":           err,
		}).Debug("Error accessing exchange code")
		return grantToken{}, err
	}
	if resp.StatusCode() != 200 {
		return grantToken{}, fmt.Errorf("bad status code %d from exchange token %s/token", resp.StatusCode(), tokenEndpoint)
	}

	var t grantToken
	err = json.Unmarshal(resp.Body(), &t)
	logger.WithFields(logrus.Fields{
		"t": fmt.Sprintf("%#v", t),
	}).Debugf("OK")
	return t, err
}
