//go:generate mockery -name Generator -inpkg
package generation

import (
	"errors"

	"bitbucket.org/openbankingteam/conformance-suite/internal/pkg/names"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/discovery"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/manifest"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/permissions"
	"github.com/sirupsen/logrus"
)

// SpecificationTestCases - test cases generated for a specification
type SpecificationTestCases struct {
	Specification discovery.ModelAPISpecification `json:"apiSpecification"`
	TestCases     []model.TestCase                `json:"testCases"`
}

type GeneratorConfig struct {
	ClientID              string
	Aud                   string
	ResponseType          string
	Scope                 string
	AuthorizationEndpoint string
	RedirectURL           string
	ResourceIDs           model.ResourceIDs
}

// Generator - generates test cases from discovery model
type Generator interface {
	GenerateSpecificationTestCases(log *logrus.Entry, config GeneratorConfig, discovery discovery.ModelDiscovery, ctx *model.Context) TestCasesRun
	GenerateManifestTests(log *logrus.Entry, config GeneratorConfig, discovery discovery.ModelDiscovery, ctx *model.Context) TestCasesRun
}

// NewGenerator - returns implementation of Generator interface
func NewGenerator() Generator {
	return generator{
		resolver: permissions.Resolver,
	}
}

// generator - implements Generator interface
type generator struct {
	resolver func(groups []permissions.Group) permissions.CodeSetResultSet
}

// Work in progress to integrate Manifest Test
func (g generator) GenerateManifestTests(log *logrus.Entry, config GeneratorConfig, discovery discovery.ModelDiscovery, ctx *model.Context) TestCasesRun {
	log = log.WithField("module", "GenerateManifestTests")
	specTestCases := []SpecificationTestCases{}
	specConsentRequirements := []model.SpecConsentRequirements{}

	for _, item := range discovery.DiscoveryItems {
		log.Tracef("GenerateManifestTests: spec=%s, uri:%s", item.APISpecification.Name, item.ResourceBaseURI)
		tcs, err := manifest.GenerateTestCases(item.APISpecification.Name, item.ResourceBaseURI, ctx) //TODO: ensure we can handle multiple specs
		if err != nil {
			log.Warnf("manifest testcase generation failed for %s", item.APISpecification.Name)
			continue
		}
		stc := SpecificationTestCases{Specification: item.APISpecification, TestCases: tcs}
		log.Debugf("%d test cases generated for %s", len(tcs), item.APISpecification.Name)
		specType, err := GetSpecType(item.APISpecification.Name)
		if err != nil {
			log.Warnf("%s\n", err.Error())
			continue
		}
		requiredTokens, _ := manifest.GetRequiredTokensFromTests(tcs, specType)
		var tokenSlice []model.SpecConsentRequirements
		tokenSlice, _ = getSpecConsentsFromRequiredTokens(requiredTokens)
		specConsentRequirements = append(specConsentRequirements, tokenSlice...)
		specTestCases = append(specTestCases, stc)
	}

	return TestCasesRun{specTestCases, specConsentRequirements}
}

// GetSpecType - returns a shorted form of the spec name
func GetSpecType(spec string) (string, error) {
	switch spec {
	case "Account and Transaction API Specification":
		return "accounts", nil
	case "Payment Initiation API":
		return "payments", nil
	}
	return "unknown", errors.New("Unknown specification " + spec)
}

func getSpecConsentsFromRequiredTokens(rt []manifest.RequiredTokens) ([]model.SpecConsentRequirements, error) {
	specConsents := make([]model.SpecConsentRequirements, 0)

	npa := []model.NamedPermission{}
	for _, v := range rt {
		np := model.NamedPermission{}
		np.Name = v.Name
		np.CodeSet = permissions.CodeSetResult{}
		np.CodeSet.TestIds = append(np.CodeSet.TestIds, permissions.StringSliceToTestID(v.IDs)...)
		np.CodeSet.CodeSet = append(np.CodeSet.CodeSet, permissions.StringSliceToCodeSet(v.Perms)...)
		npa = append(npa, np)
	}
	specConsentReq := model.SpecConsentRequirements{Identifier: "Account and Transaction API Specification", NamedPermissions: npa}
	specConsents = append(specConsents, specConsentReq)
	return specConsents, nil
}

// GenerateSpecificationTestCases - generates test cases
func (g generator) GenerateSpecificationTestCases(log *logrus.Entry, config GeneratorConfig, discovery discovery.ModelDiscovery, ctx *model.Context) TestCasesRun {
	log = log.WithField("module", "GenerateSpecificationTestCases")
	headlessTokenAcquisition := discovery.TokenAcquisition == "headless"

	specTestCases := []SpecificationTestCases{}
	customTestCases := []SpecificationTestCases{}
	customReplacements := make(map[string]string)
	originalEndpoints := make(map[string]string)
	backupEndpoints := make(map[string]string)

	for _, customTest := range discovery.CustomTests { // assume ordering is prerun i.e. customtest run before other tests
		customTestCases = append(customTestCases, GetCustomTestCases(&customTest, ctx, headlessTokenAcquisition))
		for k, v := range customTest.Replacements {
			customReplacements[k] = v
		}
		for k, testcase := range customTest.Sequence {
			if !headlessTokenAcquisition {
				ctx := model.Context{}
				ctx.PutMap(customReplacements)
				testcase.ProcessReplacementFields(&ctx, true)
			}
			customTest.Sequence[k] = testcase
		}
	}

	nameGenerator := names.NewSequentialPrefixedName("#t")
	for _, item := range discovery.DiscoveryItems {
		specTests, endpoints := generateSpecificationTestCases(log, item, nameGenerator, ctx, headlessTokenAcquisition, config)
		specTestCases = append(specTestCases, specTests)
		for k, v := range endpoints {
			originalEndpoints[k] = v
		}
	}

	tmpSpecTestCases := []SpecificationTestCases{}
	for _, specTest := range specTestCases {
		tmpSpecTestCases = append(tmpSpecTestCases, specTest)
		for x, y := range specTest.TestCases {
			backupEndpoints[y.ID] = y.Input.Endpoint
			y.Input.Endpoint = originalEndpoints[y.ID]

			specTest.TestCases[x] = y
		}
	}

	// // calculate permission set required and update the header token in the test case request
	consentRequirements := g.consentRequirements(tmpSpecTestCases) // uses pre-modified swagger urls
	log.Infof("Consent Requirements: %#v", consentRequirements)

	for _, specTest := range specTestCases {
		for x, y := range specTest.TestCases {
			y.Input.Endpoint = backupEndpoints[y.ID]
			specTest.TestCases[x] = y
		}
	}

	specTestCases = append(customTestCases, specTestCases...)
	return TestCasesRun{specTestCases, consentRequirements}

}

// consentRequirements calls resolver to get list of permission sets required to run all test cases
func (g generator) consentRequirements(specTestCases []SpecificationTestCases) []model.SpecConsentRequirements {
	nameGenerator := names.NewSequentialPrefixedName("to")
	specConsentRequirements := []model.SpecConsentRequirements{}
	for _, spec := range specTestCases {
		var groups []permissions.Group
		for _, tc := range spec.TestCases {
			g := model.NewDefaultPermissionGroup(tc)
			groups = append(groups, g)
		}
		resultSet := g.resolver(groups)
		consentRequirements := model.NewSpecConsentRequirements(nameGenerator, resultSet, spec.Specification.Name)
		specConsentRequirements = append(specConsentRequirements, consentRequirements)
	}
	return specConsentRequirements
}

// TestCasesRun represents all specs and their test and a list of tokens
// required to run those tests
type TestCasesRun struct {
	TestCases               []SpecificationTestCases        `json:"specCases"`
	SpecConsentRequirements []model.SpecConsentRequirements `json:"specTokens"`
}

func generateSpecificationTestCases(log *logrus.Entry, item discovery.ModelDiscoveryItem, nameGenerator names.Generator, ctx *model.Context, headlessTokenAcquisition bool, genConfig GeneratorConfig) (SpecificationTestCases, map[string]string) {
	testcases, originalEndpoints := GetImplementedTestCases(&item, nameGenerator, ctx, headlessTokenAcquisition, genConfig)

	for _, tc := range testcases {
		log.Debug(tc.String())
	}
	return SpecificationTestCases{Specification: item.APISpecification, TestCases: testcases}, originalEndpoints
}
