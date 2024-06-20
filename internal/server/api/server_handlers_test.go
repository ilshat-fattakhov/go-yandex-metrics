package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-playground/assert"
	"go.uber.org/zap"
)

func TestServer_PingHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		s    *Server
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.PingHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_IndexHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		s    *Server
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.IndexHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_GetHandler(t *testing.T) {
	type args struct {
		lg *zap.Logger
	}
	tests := []struct {
		name string
		s    *Server
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetHandler(tt.args.lg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.GetHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_UpdateHandler(t *testing.T) {
	type args struct {
		lg *zap.Logger
	}
	type want struct {
		code int

		contentType string
	}
	type httpParams struct {
		url    string
		method string
	}
	tests := []struct {
		s       *Server
		args    args
		name    string
		params  httpParams
		want    want
		comment string
	}{
		{
			name: "Test 405 Method Not Allowed",
			params: httpParams{
				url:    "/update/counter/PollCount/1",
				method: "GET",
			},
			want: want{
				code:        405,
				contentType: "text/plain",
			},
			comment: "Принимать метрики только по протоколу HTTP методом POST",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.params.method, tt.params.url, http.NoBody)
			rr := httptest.NewRecorder()
			handler := tt.s.UpdateHandler(tt.s.logger)
			handler.ServeHTTP(rr, req)
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Handler returned wrong status code. Expected: %d. Got: %d.", http.StatusOK, status)
			}
			tt.s.UpdateHandler(tt.args.lg)
			assert.Equal(t, tt.want.code, rr.Code)
		})
	}
}
