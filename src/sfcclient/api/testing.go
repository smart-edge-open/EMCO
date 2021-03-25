// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package api

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

func executeRequest(request *http.Request, router *mux.Router) *http.Response {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	resp := recorder.Result()
	return resp
}
