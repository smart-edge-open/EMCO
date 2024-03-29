// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	model "github.com/open-ness/EMCO/src/sfc/pkg/model"
	mock "github.com/stretchr/testify/mock"
)

// SfcClientSelectorIntentManager is an autogenerated mock type for the SfcClientSelectorIntentManager type
type SfcClientSelectorIntentManager struct {
	mock.Mock
}

// CreateSfcClientSelectorIntent provides a mock function with given fields: sfc, pr, ca, caver, dig, netctrlint, sfcIntent, exists
func (_m *SfcClientSelectorIntentManager) CreateSfcClientSelectorIntent(sfc model.SfcClientSelectorIntent, pr string, ca string, caver string, dig string, netctrlint string, sfcIntent string, exists bool) (model.SfcClientSelectorIntent, error) {
	ret := _m.Called(sfc, pr, ca, caver, dig, netctrlint, sfcIntent, exists)

	var r0 model.SfcClientSelectorIntent
	if rf, ok := ret.Get(0).(func(model.SfcClientSelectorIntent, string, string, string, string, string, string, bool) model.SfcClientSelectorIntent); ok {
		r0 = rf(sfc, pr, ca, caver, dig, netctrlint, sfcIntent, exists)
	} else {
		r0 = ret.Get(0).(model.SfcClientSelectorIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.SfcClientSelectorIntent, string, string, string, string, string, string, bool) error); ok {
		r1 = rf(sfc, pr, ca, caver, dig, netctrlint, sfcIntent, exists)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteSfcClientSelectorIntent provides a mock function with given fields: name, pr, ca, caver, dig, netctrlint, sfcIntent
func (_m *SfcClientSelectorIntentManager) DeleteSfcClientSelectorIntent(name string, pr string, ca string, caver string, dig string, netctrlint string, sfcIntent string) error {
	ret := _m.Called(name, pr, ca, caver, dig, netctrlint, sfcIntent)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string, string) error); ok {
		r0 = rf(name, pr, ca, caver, dig, netctrlint, sfcIntent)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllSfcClientSelectorIntents provides a mock function with given fields: pr, ca, caver, dig, netctrlint, sfcIntent
func (_m *SfcClientSelectorIntentManager) GetAllSfcClientSelectorIntents(pr string, ca string, caver string, dig string, netctrlint string, sfcIntent string) ([]model.SfcClientSelectorIntent, error) {
	ret := _m.Called(pr, ca, caver, dig, netctrlint, sfcIntent)

	var r0 []model.SfcClientSelectorIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string) []model.SfcClientSelectorIntent); ok {
		r0 = rf(pr, ca, caver, dig, netctrlint, sfcIntent)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.SfcClientSelectorIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, string) error); ok {
		r1 = rf(pr, ca, caver, dig, netctrlint, sfcIntent)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSfcClientSelectorIntent provides a mock function with given fields: name, pr, ca, caver, dig, netctrlint, sfcIntent
func (_m *SfcClientSelectorIntentManager) GetSfcClientSelectorIntent(name string, pr string, ca string, caver string, dig string, netctrlint string, sfcIntent string) (model.SfcClientSelectorIntent, error) {
	ret := _m.Called(name, pr, ca, caver, dig, netctrlint, sfcIntent)

	var r0 model.SfcClientSelectorIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string, string) model.SfcClientSelectorIntent); ok {
		r0 = rf(name, pr, ca, caver, dig, netctrlint, sfcIntent)
	} else {
		r0 = ret.Get(0).(model.SfcClientSelectorIntent)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, string, string) error); ok {
		r1 = rf(name, pr, ca, caver, dig, netctrlint, sfcIntent)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSfcClientSelectorIntentsByEnd provides a mock function with given fields: pr, ca, caver, dig, netctrlint, sfcIntent, chainEnd
func (_m *SfcClientSelectorIntentManager) GetSfcClientSelectorIntentsByEnd(pr string, ca string, caver string, dig string, netctrlint string, sfcIntent string, chainEnd string) ([]model.SfcClientSelectorIntent, error) {
	ret := _m.Called(pr, ca, caver, dig, netctrlint, sfcIntent, chainEnd)

	var r0 []model.SfcClientSelectorIntent
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string, string) []model.SfcClientSelectorIntent); ok {
		r0 = rf(pr, ca, caver, dig, netctrlint, sfcIntent, chainEnd)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.SfcClientSelectorIntent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, string, string) error); ok {
		r1 = rf(pr, ca, caver, dig, netctrlint, sfcIntent, chainEnd)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
