package api

import (
	"testing"
)

func TestServer_UpdateHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		s       *Server
		name    string
		method  string
		url     string
		want    want
		comment string
	}{
		{
			name:   "Test 405 Method Not Allowed",
			url:    "/update/counter/PollCount/1",
			method: "GET",
			want: want{
				code:        405,
				contentType: "text/plain",
			},
			comment: "Принимать метрики только по протоколу HTTP методом POST",
		},
		{
			name:   "Test 200 OK",
			url:    "/update/gauge/BuckHashSys/1.4",
			method: "POST",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
			comment: "При успешном приёме gauge возвращать http.StatusOK",
		},
		{
			name:   "Test 200 OK",
			url:    "/update/counter/PollCount/1",
			method: "POST",
			want: want{
				code: 200,

				contentType: "text/plain",
			},
			comment: "При успешном приёме counter возвращать http.StatusOK",
		},
		{
			name:   "Test 404 Not Found",
			url:    "/update/counter/",
			method: "POST",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
			comment: "При попытке передать запрос без имени метрики возвращать http.StatusNotFound",
		},
		{
			name:   "Test 400 Bad Request",
			url:    "/update/wrongName/test/1",
			method: "POST",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			comment: "При попытке передать запрос с некорректным типом метрики или значением возвращать http.StatusBadRequest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//r := httptest.NewRequest(tt.method, tt.url, http.NoBody)
			//w := httptest.NewRecorder()
			//tt.s.UpdateHandler(w, r)
		})
	}
}
