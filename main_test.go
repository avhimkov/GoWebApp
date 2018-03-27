package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	// "github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testRouter := SetupRouter()

	body := bytes.NewBuffer([]byte("{\"event_status\": \"83\", \"event_name\": \"100\"}"))

	req, err := http.NewRequest("POST", "/api/v1/instructions", body)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		t.Errorf("Post hearteat failed with error %d.", err)
	}

	resp := httptest.NewRecorder()
	testRouter.ServeHTTP(resp, req)

	if resp.Code != 201 {
		t.Errorf("/api/v1/instructions failed with error code %d.", resp.Code)
	}
}
