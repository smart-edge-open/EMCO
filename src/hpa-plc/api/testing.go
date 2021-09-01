// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
)

func executeRequest(request *http.Request, router *mux.Router) *http.Response {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	resp := recorder.Result()
	return resp
}

type mockIntentManager struct {
	Items             []hpaModel.DeploymentHpaIntent
	ItemsSpec         []hpaModel.DeploymentHpaIntentSpec
	ConsumerItems     []hpaModel.HpaResourceConsumer
	ConsumerItemsSpec []hpaModel.HpaResourceConsumerSpec
	ResourceItems     []hpaModel.HpaResourceRequirement
	ResourceItemsSpec []hpaModel.HpaResourceRequirementSpec
	Err               error
}

func (m *mockIntentManager) AddIntent(a hpaModel.DeploymentHpaIntent, p string, ca string, v string, di string, exists bool) (hpaModel.DeploymentHpaIntent, error) {
	if m.Err != nil {
		return hpaModel.DeploymentHpaIntent{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockIntentManager) GetIntent(name string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, bool, error) {
	if m.Err != nil {
		return hpaModel.DeploymentHpaIntent{}, false, m.Err
	}

	return m.Items[0], true, nil
}

func (m *mockIntentManager) GetAllIntents(p string, ca string, v string, di string) ([]hpaModel.DeploymentHpaIntent, error) {

	if m.Err != nil {
		return []hpaModel.DeploymentHpaIntent{}, m.Err
	}

	return m.Items, nil
}

func (m *mockIntentManager) GetAllIntentsByApp(app string, p string, ca string, v string, di string) ([]hpaModel.DeploymentHpaIntent, error) {

	if m.Err != nil {
		return []hpaModel.DeploymentHpaIntent{}, m.Err
	}

	return m.Items, nil
}

func (m *mockIntentManager) GetIntentByName(i string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, error) {

	if m.Err != nil {
		return hpaModel.DeploymentHpaIntent{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockIntentManager) DeleteIntent(i string, p string, ca string, v string, di string) error {
	return m.Err
}

// consumers
func (m *mockIntentManager) AddConsumer(a hpaModel.HpaResourceConsumer, p string, ca string, v string, di string, i string, exists bool) (hpaModel.HpaResourceConsumer, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceConsumer{}, m.Err
	}

	return m.ConsumerItems[0], nil
}

func (m *mockIntentManager) GetConsumer(cn string, p string, ca string, v string, di string, i string) (hpaModel.HpaResourceConsumer, bool, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceConsumer{}, false, m.Err
	}
	return m.ConsumerItems[0], false, nil
}

func (m *mockIntentManager) GetAllConsumers(p, ca, v, di, i string) ([]hpaModel.HpaResourceConsumer, error) {
	if m.Err != nil {
		return []hpaModel.HpaResourceConsumer{}, m.Err
	}
	return m.ConsumerItems, nil
}

func (m *mockIntentManager) GetConsumerByName(cn, p, ca, v, di, i string) (hpaModel.HpaResourceConsumer, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceConsumer{}, m.Err
	}
	return m.ConsumerItems[0], nil
}

func (m *mockIntentManager) DeleteConsumer(cn, p string, ca string, v string, di string, i string) error {
	return nil
}

// resources
func (m *mockIntentManager) AddResource(a hpaModel.HpaResourceRequirement, p string, ca string, v string, di string, i string, cn string, exists bool) (hpaModel.HpaResourceRequirement, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceRequirement{}, m.Err
	}

	return m.ResourceItems[0], nil
}

func (m *mockIntentManager) GetResource(rn string, p string, ca string, v string, di string, i string, cn string) (hpaModel.HpaResourceRequirement, bool, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceRequirement{}, false, m.Err
	}

	return m.ResourceItems[0], false, nil
}

func (m *mockIntentManager) GetAllResources(p, ca, v, di, i, cn string) ([]hpaModel.HpaResourceRequirement, error) {
	if m.Err != nil {
		return []hpaModel.HpaResourceRequirement{}, m.Err
	}

	return m.ResourceItems, nil

}

func (m *mockIntentManager) GetResourceByName(rn, p, ca, v, di, i, cn string) (hpaModel.HpaResourceRequirement, error) {
	if m.Err != nil {
		return hpaModel.HpaResourceRequirement{}, m.Err
	}

	return m.ResourceItems[0], nil
}

func (m *mockIntentManager) DeleteResource(rn string, p string, ca string, v string, di string, i string, cn string) error {
	return nil
}
