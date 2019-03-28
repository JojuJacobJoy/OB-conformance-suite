package manifest

import (
	"fmt"
	"testing"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"
	"github.com/stretchr/testify/assert"
)

func loadData() (References, error) {
	data, err := loadReferences("../../manifests/data.json")
	//data, err := loadReferences("../../manifests/assertions.json")
	if err != nil {
		return References{}, err
	}
	return data, nil
}

func TestDataReferencesAndDump(t *testing.T) {
	data, err := loadData()
	assert.Nil(t, err)
	dumpJSON(data)
	for k, v := range data.References {
		body := jsonString(v.Body)
		l := len(body)
		if l > 0 {
			v.BodyData = body
			v.Body = ""
			data.References[k] = v
		}
	}
	dumpJSON(data)
}

func TestGenerateTestCases(t *testing.T) {
	tests, err := GenerateTestCases("TestSpec", "http://mybaseurl", &model.Context{})
	assert.Nil(t, err)

	perms, err := GetPermissions(tests)
	assert.Nil(t, err)
	m := make(map[string]string, 0)
	for _, v := range perms {
		m[v.Path] = v.ID
	}
	for k := range m {
		fmt.Println(k)
	}
	fmt.Println("DumpTests...............................:----")
	for _, v := range tests {
		dumpJSON(v)
	}

}

func TestCheckPaymentPermissions(t *testing.T) {
	tests, err := GenerateTestCases("TestSpec", "http://mybaseurl", &model.Context{})
	assert.Nil(t, err)

	perms, err := GetPaymentPermissions(tests)
	assert.Nil(t, err)
	m := make(map[string]string, 0)
	for _, v := range perms {
		m[v.Path] = v.ID
		fmt.Printf("%s %s\n", v.ID, v.Path)
	}

	fmt.Println("DumpTests:----")
	for _, v := range tests {
		dumpJSON(v)
	}

}
