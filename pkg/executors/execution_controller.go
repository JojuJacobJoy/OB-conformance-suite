package executors

import (
	"encoding/json"
	"fmt"
	"sync"

	"gopkg.in/resty.v1"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/authentication"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/discovery"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/executors/results"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/generation"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/model"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/tracer"
)

// RunDefinition captures all the information required to run the test cases
type RunDefinition struct {
	DiscoModel    *discovery.Model
	TestCaseRun   generation.TestCasesRun
	SigningCert   authentication.Certificate
	TransportCert authentication.Certificate
}

type TestCaseRunner struct {
	executor         TestCaseExecutor
	definition       RunDefinition
	daemonController DaemonController
	logger           *logrus.Entry
	runningLock      *sync.Mutex
	running          bool
}

// NewTestCaseRunner -
func NewTestCaseRunner(logger *logrus.Entry, definition RunDefinition, daemonController DaemonController) *TestCaseRunner {
	return &TestCaseRunner{
		executor:         NewExecutor(),
		definition:       definition,
		daemonController: daemonController,
		logger:           logger.WithField("module", "TestCaseRunner"),
		runningLock:      &sync.Mutex{},
		running:          false,
	}
}

// NewConsentAcquisitionRunner -
func NewConsentAcquisitionRunner(logger *logrus.Entry, definition RunDefinition, daemonController DaemonController) *TestCaseRunner {
	return &TestCaseRunner{
		executor:         NewExecutor(),
		definition:       definition,
		daemonController: daemonController,
		logger:           logger.WithField("module", "ConsentAcquisitionRunner"),
		runningLock:      &sync.Mutex{},
		running:          false,
	}
}

// NewExchangeComponentRunner -
func NewExchangeComponentRunner(definition RunDefinition, daemonController DaemonController) *TestCaseRunner {
	return &TestCaseRunner{
		executor:         NewExecutor(),
		definition:       definition,
		daemonController: daemonController,
		logger:           logrus.StandardLogger().WithField("module", "ExchangeComponent"),
		runningLock:      &sync.Mutex{},
		running:          false,
	}
}

// RunTestCases runs the testCases
func (r *TestCaseRunner) RunTestCases(ctx *model.Context) error {
	r.runningLock.Lock()
	defer func() {
		r.runningLock.Unlock()
	}()
	if r.running {
		return errors.New("test cases runner already running")
	}
	r.running = true

	go r.runTestCasesAsync(ctx)

	return nil
}

// RunConsentAcquisition -
func (r *TestCaseRunner) RunConsentAcquisition(item TokenConsentIDItem, ctx *model.Context, consentType string, consentIDChannel chan<- TokenConsentIDItem) error {
	r.runningLock.Lock()
	defer func() {
		r.runningLock.Unlock()
	}()
	if r.running {
		return errors.New("consent acquisition test cases runner already running")
	}
	r.running = true
	logrus.Tracef("runConsentAquisition with %s, %s, %s\n", item.TokenName, item.ConsentURL, item.Permissions)
	go r.runConsentAcquisitionAsync(item, ctx, consentType, consentIDChannel)

	return nil
}

func (r *TestCaseRunner) runTestCasesAsync(ctx *model.Context) {
	err := r.executor.SetCertificates(r.definition.SigningCert, r.definition.TransportCert)
	if err != nil {
		r.logger.WithError(err).Error("running test cases async")
	}

	ruleCtx := r.makeRuleCtx(ctx)

	ctxLogger := r.logger.WithField("id", uuid.New())
	for _, spec := range r.definition.TestCaseRun.TestCases {
		r.executeSpecTests(spec, ruleCtx, ctxLogger)
	}
	r.daemonController.SetCompleted()

	r.setNotRunning()
}

func (r *TestCaseRunner) runConsentAcquisitionAsync(item TokenConsentIDItem, ctx *model.Context, consentType string, consentIDChannel chan<- TokenConsentIDItem) {
	err := r.executor.SetCertificates(r.definition.SigningCert, r.definition.TransportCert)
	if err != nil {
		r.logger.WithError(err).Error("running consent acquisition async")
	}

	ruleCtx := r.makeRuleCtx(ctx)
	ruleCtx.PutString("consent_id", item.TokenName)
	ruleCtx.PutString("token_name", item.TokenName)
	ruleCtx.PutString("permission_list", item.Permissions)

	ctxLogger := r.logger.WithField("id", uuid.New())
	var comp model.Component
	if consentType == "psu" {
		comp, err = model.LoadComponent("PSUConsentProviderComponent.json")
		if err != nil {
			r.AppMsg("Load PSU Component Failed: " + err.Error())
			r.setNotRunning()
			return
		}
	} else {
		comp, err = model.LoadComponent("headlessTokenProviderProviderComponent.json")
		if err != nil {
			r.AppMsg("Load HeadlessConsent Component Failed: " + err.Error())
			r.setNotRunning()
			return
		}
	}

	err = comp.ValidateParameters(ruleCtx) // correct parameters for component exist in context
	if err != nil {
		msg := fmt.Sprintf("component execution error: component (%s) cannot ValidateParameters: %s", comp.Name, err.Error())
		r.AppMsg(msg)
		r.setNotRunning()
		return
	}

	for k, v := range comp.GetTests() {
		v.ProcessReplacementFields(ruleCtx, true)
		comp.Tests[k] = v
	}

	r.executeComponentTests(&comp, ruleCtx, ctxLogger, item, consentIDChannel)
	clientGrantToken, err := ruleCtx.GetString("client_access_token")
	if err == nil {
		logrus.StandardLogger().WithFields(logrus.Fields{
			"clientGrantToken": clientGrantToken,
		}).Debugf("Setting client_access_token")
		ctx.PutString("client_access_token", clientGrantToken)
	}

	r.setNotRunning()
}

func (r *TestCaseRunner) executeComponentTests(comp *model.Component, ruleCtx *model.Context, logger *logrus.Entry, item TokenConsentIDItem, consentIDChannel chan<- TokenConsentIDItem) {
	ctxLogger := logger.WithFields(logrus.Fields{
		"component": comp.Name,
		"module":    "TestCaseRunner",
		"function":  "executeComponentTests",
	})

	for _, testcase := range comp.Tests {
		if r.daemonController.ShouldStop() {
			ctxLogger.Debug("stop component test run received, aborting runner")
			return
		}

		testResult := r.executeTest(testcase, ruleCtx, logger)
		r.daemonController.AddResult(testResult)

		if testResult.Pass {
			ctxLogger.WithFields(logrus.Fields{
				"item": fmt.Sprintf("%#v", item),
			}).Debug("hanging around for token (TokenConsentIDItem)")
			consentURL, err := ruleCtx.GetString("consent_url")
			if err == model.ErrNotFound {
				continue
			}

			item.ConsentURL = consentURL
			ruleCtx.DumpContext()
			consentID, err := ruleCtx.GetString(item.TokenName)
			if err == model.ErrNotFound {
				ctxLogger.WithFields(logrus.Fields{
					"item.TokenName": fmt.Sprintf("%+v", item.TokenName),
					"err":            err,
				}).Warn("Did not find consentID in context for item.TokenName")
			}
			item.ConsentID = consentID

			ctxLogger.WithFields(logrus.Fields{
				"item": fmt.Sprintf("%#v", item),
			}).Debug("Sending item (TokenConsentIDItem) to consentIDChannel")
			consentIDChannel <- item
		} else if len(testResult.Fail) > 0 {
			item.Error = testResult.Fail[0]
			consentIDChannel <- item
		}
	}
}

func (r *TestCaseRunner) setNotRunning() {
	logger := logrus.StandardLogger().WithFields(logrus.Fields{
		"function": "setNotRunning",
		"module":   "TestCaseRunner",
	})

	logger.Debug("acquiring runningLock")
	r.runningLock.Lock()
	logger.Debug("acquired runningLock")
	defer func() {
		logger.Debug("releasing runningLock")
		r.runningLock.Unlock()

	}()
	r.running = false
}

func (r *TestCaseRunner) makeRuleCtx(ctx *model.Context) *model.Context {
	ruleCtx := &model.Context{}
	ruleCtx.Put("SigningCert", r.definition.SigningCert)
	ruleCtx.PutContext(ctx)
	return ruleCtx
}

func (r *TestCaseRunner) executeSpecTests(spec generation.SpecificationTestCases, ruleCtx *model.Context, ctxLogger *logrus.Entry) {
	ctxLogger = ctxLogger.WithField("spec", spec.Specification.Name)
	for _, testcase := range spec.TestCases {
		if r.daemonController.ShouldStop() {
			ctxLogger.Info("stop test run received, aborting runner")
			return
		}
		ruleCtx.DumpContext("ruleCtx before: " + testcase.ID)
		testResult := r.executeTest(testcase, ruleCtx, ctxLogger)
		r.daemonController.AddResult(testResult)
	}
}

func (r *TestCaseRunner) executeTest(tc model.TestCase, ruleCtx *model.Context, logger *logrus.Entry) results.TestCase {
	ctxLogger := logWithTestCase(logger, tc)
	req, err := tc.Prepare(ruleCtx)
	if err != nil {
		ctxLogger.WithError(err).Error("preparing executing test")
		return results.NewTestCaseFail(tc.ID, results.NoMetrics, []error{err})
	}
	resp, metrics, err := r.executor.ExecuteTestCase(req, &tc, ruleCtx)
	ctxLogger = logWithMetrics(ctxLogger, metrics)
	if err != nil {
		ctxLogger.WithError(err).WithFields(logrus.Fields{"result": "FAIL", "ID": tc.ID}).Error("test result")
		return results.NewTestCaseFail(tc.ID, metrics, []error{err})
	}

	result, errs := tc.Validate(resp, ruleCtx)
	if errs != nil {
		detailedErrors := detailedErrors(errs, resp)
		ctxLogger.WithField("errs", detailedErrors).WithFields(logrus.Fields{"result": passText[result], "ID": tc.ID}).Error("test result validate")
		return results.NewTestCaseFail(tc.ID, metrics, detailedErrors)
	}

	if !result {
		ctxLogger.WithError(err).WithFields(logrus.Fields{"result": passText[result], "ID": tc.ID}).Error("test result blank")
	} else {
		ctxLogger.WithError(err).WithFields(logrus.Fields{"result": passText[result], "ID": tc.ID}).Info("test result")
	}

	return results.NewTestCaseResult(tc.ID, result, metrics, []error{})
}

func detailedErrors(errs []error, resp *resty.Response) []error {
	var detailedErrors []error
	for _, err := range errs {
		detailedError := errors.WithMessagef(err, "Response: (%.250s)", resp.String())
		detailedErrors = append(detailedErrors, detailedError)
	}
	return detailedErrors
}

var passText = map[bool]string{
	true:  "PASS",
	false: "FAIL",
}

func logWithTestCase(logger *logrus.Entry, tc model.TestCase) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"TestCase.Name":              tc.Name,
		"TestCase.Input.Method":      tc.Input.Method,
		"TestCase.Input.Endpoint":    tc.Input.Endpoint,
		"TestCase.Expect.StatusCode": tc.Expect.StatusCode,
	})
}

func logWithMetrics(logger *logrus.Entry, metrics results.Metrics) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"responsetime": fmt.Sprintf("%v", metrics.ResponseTime),
		"responsesize": metrics.ResponseSize,
	})
}

// AppMsg - application level trace
func (r *TestCaseRunner) AppMsg(msg string) string {
	tracer.AppMsg("TestCaseRunner", msg, r.String())
	return msg
}

// AppErr - application level trace error msg
func (r *TestCaseRunner) AppErr(msg string) error {
	tracer.AppErr("TestCaseRunner", msg, r.String())
	return errors.New(msg)
}

// String - object represetation
func (r *TestCaseRunner) String() string {
	bites, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		// String() doesn't return error but still want to log as error to tracer ...
		return r.AppErr(fmt.Sprintf("error converting TestCaseRunner  %s", err.Error())).Error()
	}
	return string(bites)
}
