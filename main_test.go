package main

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	//"bytes"
	//"net/http"
	//"net/http/httptest"
)

func TestSetupRouter(t *testing.T) {
	tests := []struct {
		name string
		want *gin.Engine
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetupRouter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}


//import (
//"bytes"
//"net/http"
//"net/http/httptest"
//"testing"
//
//"github.com/gin-gonic/gin"
//// "github.com/stretchr/testify/assert"
//)
//
//func Test_main(t *testing.T) {
//	gin.SetMode(gin.TestMode)
//	testRouter := SetupRouter()
//
//	body := bytes.NewBuffer([]byte("{\"event_status\": \"83\", \"event_name\": \"100\"}"))
//
//	req, err := http.NewRequest("POST", "/api/v1/instructions", body)
//	req.Header.Set("Content-Type", "application/json")
//	if err != nil {
//		t.Errorf("Post hearteat failed with error %d.", err)
//	}
//
//	resp := httptest.NewRecorder()
//	testRouter.ServeHTTP(resp, req)
//
//	if resp.Code != 201 {
//		t.Errorf("/api/v1/instructions failed with error code %d.", resp.Code)
//	}
//}

