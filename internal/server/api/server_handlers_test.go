package api

import (
	logger "go-yandex-metrics/internal/server/middleware"
	"log"
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
	type want struct {
		code        int
		contentType string
	}
	type httpParams struct {
		url    string
		method string
	}
	tests := []struct {
		s       *Server
		name    string
		params  httpParams
		want    want
		comment string
	}{
		{
			name: "Test 400 Bad Request",
			params: httpParams{
				url:    "/update/wrongName/test/1",
				method: "POST",
			},
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			comment: "При попытке передать запрос с некорректным типом метрики или значением возвращать http.StatusBadRequest",
		},
	}
	lg, err := logger.InitLogger()
	if err != nil {
		log.Fatal("failed to init logger: %w", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.params.method, tt.params.url, http.NoBody)
			rr := httptest.NewRecorder()

			handler := tt.s.UpdateHandler(lg)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.code, rr.Code)
		})
	}
}
