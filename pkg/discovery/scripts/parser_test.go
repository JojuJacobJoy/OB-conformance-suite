package scripts

import (
	"io/ioutil"
	"testing"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/test"
	"github.com/sirupsen/logrus"
)

func TestParser_AccountInfoSwagger(t *testing.T) {
	require := test.NewRequire(t)

	swaggerPath := "../../schema/spec/v3.1.2/account-info-swagger.flattened.json"
	// outputFile := "pkg/discovery/scripts/generated/v3.1.2_account-info-discovery.json"
	logger := nullLogger()

	nonManadatoryFields, err := ParseSchema(swaggerPath, logger)
	require.NoError(err)
	require.NotNil(nonManadatoryFields)
}

func nullLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Out = ioutil.Discard

	return logger.WithFields(logrus.Fields{
		"test": "bitbucket.org/openbankingteam/conformance-suite/pkg/discovery/scripts",
	}).Logger
}
