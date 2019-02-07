// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import authentication "bitbucket.org/openbankingteam/conformance-suite/pkg/authentication"
import discovery "bitbucket.org/openbankingteam/conformance-suite/pkg/discovery"
import executors "bitbucket.org/openbankingteam/conformance-suite/pkg/executors"
import mock "github.com/stretchr/testify/mock"
import server "bitbucket.org/openbankingteam/conformance-suite/pkg/server"

// Journey is an autogenerated mock type for the Journey type
type Journey struct {
	mock.Mock
}

// DiscoveryModel provides a mock function with given fields:
func (_m *Journey) DiscoveryModel() (*discovery.Model, error) {
	ret := _m.Called()

	var r0 *discovery.Model
	if rf, ok := ret.Get(0).(func() *discovery.Model); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*discovery.Model)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Results provides a mock function with given fields:
func (_m *Journey) Results() executors.DaemonController {
	ret := _m.Called()

	var r0 executors.DaemonController
	if rf, ok := ret.Get(0).(func() executors.DaemonController); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(executors.DaemonController)
		}
	}

	return r0
}

// RunTests provides a mock function with given fields:
func (_m *Journey) RunTests() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetCertificates provides a mock function with given fields: signing, transport
func (_m *Journey) SetCertificates(signing authentication.Certificate, transport authentication.Certificate) {
	_m.Called(signing, transport)
}

// SetDiscoveryModel provides a mock function with given fields: discoveryModel
func (_m *Journey) SetDiscoveryModel(discoveryModel *discovery.Model) (discovery.ValidationFailures, error) {
	ret := _m.Called(discoveryModel)

	var r0 discovery.ValidationFailures
	if rf, ok := ret.Get(0).(func(*discovery.Model) discovery.ValidationFailures); ok {
		r0 = rf(discoveryModel)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(discovery.ValidationFailures)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*discovery.Model) error); ok {
		r1 = rf(discoveryModel)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StopTestRun provides a mock function with given fields:
func (_m *Journey) StopTestRun() {
	_m.Called()
}

// TestCases provides a mock function with given fields:
func (_m *Journey) TestCases() (server.TestCasesRun, error) {
	ret := _m.Called()

	var r0 server.TestCasesRun
	if rf, ok := ret.Get(0).(func() server.TestCasesRun); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(server.TestCasesRun)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
