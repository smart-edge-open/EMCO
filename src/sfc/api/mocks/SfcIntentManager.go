// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	model "github.com/open-ness/EMCO/src/sfc/pkg/model"
	mock "github.com/stretchr/testify/mock"
)

// SfcIntentManager is an autogenerated mock type for the SfcIntentManager type
type SfcIntentManager struct {
	mock.Mock
}

// CreateSfcIntent provides a mock function with given fields: sfc, pr, ca, caver, dig, netctrlint, exists
func (_m *SfcIntentManager) CreateSfcIntent(sfc model.SfcIntent, pr string, ca string, caver string, dig string, netctrlint string, exists bool) (model.SfcIntent, error) {
	ret := _m.Called(sfc, pr, ca, caver, dig, netctrlint, exists)

	var r0 model.SfcIntent
	if rf, ok := ret.Get(0).(func(model.SfcIntent, string, string, string, string, string, bool) model.SfcIntent); ok {
		r0 = rf(sfc, pr, ca, caver, dig, netctrlint, exists)
	} else {
		r0 = ret.Get(0).(model.SfcIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.SfcIntent, string, string, string, string, string, bool) error); ok {
		r1 = rf(sfc, pr, ca, caver, dig, netctrlint, exists)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteSfcIntent provides a mock function with given fields: name, pr, ca, caver, dig, netctrlint
func (_m *SfcIntentManager) DeleteSfcIntent(name string, pr string, ca string, caver string, dig string, netctrlint string) error {
	ret := _m.Called(name, pr, ca, caver, dig, netctrlint)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string) error); ok {
		r0 = rf(name, pr, ca, caver, dig, netctrlint)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllSfcIntents provides a mock function with given fields: pr, ca, caver, dig, netctrlint
func (_m *SfcIntentManager) GetAllSfcIntents(pr string, ca string, caver string, dig string, netctrlint string) ([]model.SfcIntent, error) {
	ret := _m.Called(pr, ca, caver, dig, netctrlint)

	var r0 []model.SfcIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string) []model.SfcIntent); ok {
		r0 = rf(pr, ca, caver, dig, netctrlint)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.SfcIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string) error); ok {
		r1 = rf(pr, ca, caver, dig, netctrlint)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSfcIntent provides a mock function with given fields: name, pr, ca, caver, dig, netctrlint
func (_m *SfcIntentManager) GetSfcIntent(name string, pr string, ca string, caver string, dig string, netctrlint string) (model.SfcIntent, error) {
	ret := _m.Called(name, pr, ca, caver, dig, netctrlint)

	var r0 model.SfcIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string) model.SfcIntent); ok {
		r0 = rf(name, pr, ca, caver, dig, netctrlint)
	} else {
		r0 = ret.Get(0).(model.SfcIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, string) error); ok {
		r1 = rf(name, pr, ca, caver, dig, netctrlint)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
