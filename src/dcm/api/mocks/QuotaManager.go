// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import module "github.com/open-ness/EMCO/src/dcm/pkg/module"

// QuotaManager is an autogenerated mock type for the QuotaManager type
type QuotaManager struct {
	mock.Mock
}

// CreateQuota provides a mock function with given fields: project, logicalCloud, c
func (_m *QuotaManager) CreateQuota(project string, logicalCloud string, c module.Quota) (module.Quota, error) {
	ret := _m.Called(project, logicalCloud, c)

	var r0 module.Quota
	if rf, ok := ret.Get(0).(func(string, string, module.Quota) module.Quota); ok {
		r0 = rf(project, logicalCloud, c)
	} else {
		r0 = ret.Get(0).(module.Quota)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, module.Quota) error); ok {
		r1 = rf(project, logicalCloud, c)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteQuota provides a mock function with given fields: project, logicalCloud, name
func (_m *QuotaManager) DeleteQuota(project string, logicalCloud string, name string) error {
	ret := _m.Called(project, logicalCloud, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(project, logicalCloud, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllQuotas provides a mock function with given fields: project, logicalCloud
func (_m *QuotaManager) GetAllQuotas(project string, logicalCloud string) ([]module.Quota, error) {
	ret := _m.Called(project, logicalCloud)

	var r0 []module.Quota
	if rf, ok := ret.Get(0).(func(string, string) []module.Quota); ok {
		r0 = rf(project, logicalCloud)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]module.Quota)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(project, logicalCloud)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetQuota provides a mock function with given fields: project, logicalCloud, name
func (_m *QuotaManager) GetQuota(project string, logicalCloud string, name string) (module.Quota, error) {
	ret := _m.Called(project, logicalCloud, name)

	var r0 module.Quota
	if rf, ok := ret.Get(0).(func(string, string, string) module.Quota); ok {
		r0 = rf(project, logicalCloud, name)
	} else {
		r0 = ret.Get(0).(module.Quota)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(project, logicalCloud, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateQuota provides a mock function with given fields: project, logicalCloud, name, c
func (_m *QuotaManager) UpdateQuota(project string, logicalCloud string, name string, c module.Quota) (module.Quota, error) {
	ret := _m.Called(project, logicalCloud, name, c)

	var r0 module.Quota
	if rf, ok := ret.Get(0).(func(string, string, string, module.Quota) module.Quota); ok {
		r0 = rf(project, logicalCloud, name, c)
	} else {
		r0 = ret.Get(0).(module.Quota)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, module.Quota) error); ok {
		r1 = rf(project, logicalCloud, name, c)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
