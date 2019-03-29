package manifest

import (
	"errors"
	"fmt"
	"strings"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"
	"github.com/sirupsen/logrus"
)

// TestCasePermission -
type TestCasePermission struct {
	ID     string   `json:"id,omitempty"`
	Perms  []string `json:"perms,omitempty"`
	Permsx []string `json:"permsx,omitempty"`
}

// RequiredTokens -
type RequiredTokens struct {
	Name         string   `json:"name,omitempty"`
	Token        string   `json:"token,omitempty"`
	IDs          []string `json:"ids,omitempty"`
	Perms        []string `json:"perms,omitempty"`
	Permsx       []string `json:"permsx,omitempty"`
	AccessToken  string
	ConsentURL   string
	ConsentID    string
	ConsentParam string
}

// TokenStore eats tokens
type TokenStore struct {
	currentID int
	store     []RequiredTokens
}

// GetSpecType -
func GetSpecType(s string) (string, error) {
	spec := strings.TrimSpace(s)
	switch spec {
	case "Account and Transaction API Specification":
		return "accounts", nil
	case "Payment Initiation API":
		return "payments", nil
	case "Confirmation of Funds API Specification":
		return "funds", nil
	case "Event Notification API Specification - ASPSP Endpoints":
		return "notifications", nil
	}
	return "unknown", errors.New("Unknown specification:  `" + spec + "`")
}

// GetRequiredTokensFromTests - Given a set of testcases with the permissions defined
// in the context using 'permissions' and 'permissions-excluded'
// provides a RequiredTokens structure which can be used to capture token requirements
func GetRequiredTokensFromTests(tcs []model.TestCase, spec string) (rt []RequiredTokens, err error) {
	switch spec {
	case "accounts":
		tcp, err := getTestCasePermissions(tcs)
		if err != nil {
			return nil, err
		}
		rt, err = getRequiredTokens(tcp)
	case "payments":
		rt, err = getPaymentPermissions(tcs)
	}
	return rt, err
}

func getPaymentPermissions(tcs []model.TestCase) ([]RequiredTokens, error) {
	rt := make([]RequiredTokens, 0)
	ts := TokenStore{}
	ts.store = rt

	for _, tc := range tcs {
		ctx := tc.Context
		consentRequired, found := ctx.GetString("requestConsent")
		if found != nil {
			continue
		}
		if consentRequired == "true" {
			// get consentid
			consentID := getConsentIDFromMatches(tc)

			rx := RequiredTokens{Name: ts.GetNextTokenName("payment"), ConsentParam: consentID}
			rx.IDs = append(rx.IDs, tc.ID)
			rt = append(rt, rx)
		}
	}

	return rt, nil
}

func updateTokensFromConsent(rts []RequiredTokens, tcs []model.TestCase) ([]RequiredTokens, error) {
	for rtidx, rt := range rts {
		for _, test := range tcs {
			ctx := test.Context
			value, _ := ctx.GetString("consentId")
			if len(value) > 1 {
				if rt.ConsentParam == value[1:] {
					rt.IDs = append(rt.IDs, test.ID)
					rts[rtidx] = rt
				}
			}
		}
	}
	return rts, nil
}

func getConsentIDFromMatches(tc model.TestCase) string {
	matches := tc.Expect.ContextPut.Matches
	for _, m := range matches {
		if m.Value == "json.Data.ConsentId" {
			return m.ContextName
		}
	}
	return ""
}

// GetTestCasePermissions -
func getTestCasePermissions(tcs []model.TestCase) ([]TestCasePermission, error) {
	tcps := []TestCasePermission{}
	for _, tc := range tcs {
		ctx := tc.Context
		perms, found := ctx.GetStringSlice("permissions")
		if found != nil {
			continue
		}
		permsx, _ := ctx.GetStringSlice("permissions-excluded")
		tcp := TestCasePermission{ID: tc.ID, Perms: perms, Permsx: permsx}
		tcps = append(tcps, tcp)
	}
	return tcps, nil
}

// GetRequiredTokens - gathers all tokens
func getRequiredTokens(tcps []TestCasePermission) ([]RequiredTokens, error) {
	te := TokenStore{}
	for _, tcp := range tcps {
		te.createOrUpdate(tcp)
	}
	return te.store, nil
}

// MapTokensToTestCases - applies consented tokens to testcases
func MapTokensToTestCases(rt []RequiredTokens, tcs []model.TestCase) map[string]string {
	tokenMap := make(map[string]string, 0)
	for k, test := range tcs {
		tokenName, isEmptyToken, err := getRequiredTokenForTestcase(rt, test.ID)
		if err != nil {
			logrus.Warnf("no token for testcase %s", test.ID)
			continue
		}
		if !isEmptyToken {
			test.InjectBearerToken("$" + tokenName)
		}
		tcs[k] = test
	}
	for _, v := range rt {
		tokenMap[v.Name] = v.Token
	}

	return tokenMap
}

func getRequiredTokenForTestcase(rt []RequiredTokens, testcaseID string) (tokenName string, isEmptyToken bool, err error) {
	for _, v := range rt {
		if len(v.Perms) == 0 {
			return "", true, nil
		}
		for _, id := range v.IDs {
			if testcaseID == id {
				return v.Name, false, nil
			}
		}
	}
	return "", false, errors.New("token not found for " + testcaseID)
}

func dumpTG(tg []RequiredTokens) {
	for _, v := range tg {
		fmt.Printf("grouplineitem: %v - %v -  %v\n", v.IDs, v.Perms, v.Permsx)
	}
}

// GetNextTokenName -
func (te *TokenStore) GetNextTokenName(s string) string {
	te.currentID++
	return fmt.Sprintf("%sToken%4.4d", s, te.currentID)
}

// create or update TokenGethereer
func (te *TokenStore) createOrUpdate(tcp TestCasePermission) {

	if len(te.store) == 0 { // First time - no permissions - just add
		tpg := RequiredTokens{Name: te.GetNextTokenName("account"), IDs: []string{tcp.ID}, Perms: tcp.Perms, Permsx: tcp.Permsx}
		te.store = append(te.store, tpg)
		return
	}

	if len(tcp.Perms) == 0 && len(tcp.Permsx) == 0 {
		for idx, tgItem := range te.store {
			if len(tgItem.Perms) == 0 && len(tgItem.Permsx) == 0 {
				te.store[idx].IDs = append(te.store[idx].IDs, tcp.ID)
				return
			}
		}
		tpg := RequiredTokens{Name: te.GetNextTokenName("account"), IDs: []string{tcp.ID}, Perms: tcp.Perms, Permsx: tcp.Permsx}
		te.store = append(te.store, tpg)
	}

	for idx, tgItem := range te.store { // loop through each Gathered Item
		tcPermxConflict := false
		tcPermConflict := false

		// Check groupPermissions against testcaseExclusions
		for _, tgperm := range tgItem.Perms { // loop through all
			for _, tcpermx := range tcp.Permsx {
				if tgperm == tcpermx {
					tcPermxConflict = true
					break
				}
			}
			if tcPermxConflict {
				break
			}
		}
		if tcPermxConflict { //move onto next group item
			continue
		}

		// Check groupExclusions against testcasePermissions
		for _, tgpermx := range tgItem.Permsx {
			for _, tcperm := range tcp.Perms {
				if tgpermx == tcperm {
					tcPermConflict = true
					break
				}
			}
			if tcPermConflict {
				break
			}
		}
		if tcPermConflict {
			continue
		}
		newItem := addPermToGathererItem(tcp, tgItem)
		te.store[idx] = newItem
		return
	}
	tpg := RequiredTokens{Name: te.GetNextTokenName("account"), IDs: []string{tcp.ID}, Perms: tcp.Perms, Permsx: tcp.Permsx}
	te.store = append(te.store, tpg)

	return
}

func addPermToGathererItem(tp TestCasePermission, tg RequiredTokens) RequiredTokens {
	tg.IDs = append(tg.IDs, tp.ID)
	permsToAdd := []string{}
	permsxToAdd := []string{}
	for _, tgPerm := range tg.Perms {
		for _, tpPerm := range tp.Perms {
			if tpPerm == tgPerm {
				continue
			} else {
				if tpPerm != "" {
					permsToAdd = append(permsToAdd, tpPerm)
				}
			}
		}
	}
	for _, tgPermx := range tg.Permsx {
		for _, tpPermx := range tp.Permsx {
			if tpPermx == tgPermx {
				continue
			} else {
				if tpPermx != "" {
					permsxToAdd = append(permsxToAdd, tpPermx)
				}
			}
		}
	}
	tg.Perms = append(tg.Perms, permsToAdd...)
	tg.Perms = uniqueSlice(tg.Perms)
	tg.Permsx = append(tg.Permsx, permsxToAdd...)
	tg.Permsx = uniqueSlice(tg.Permsx)

	return tg
}

func uniqueSlice(inslice []string) []string {
	compressor := map[string]bool{}
	for _, v := range inslice {
		compressor[v] = true
	}
	tmpslice := []string{}
	for k := range compressor {
		tmpslice = append(tmpslice, k)
	}
	return tmpslice
}
