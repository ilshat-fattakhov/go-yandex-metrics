package main

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_updateHandler(t *testing.T) {

	type want struct {
		code int

		contentType string
	}
	type httpParams struct {
		url    string
		method string
	}
	tests := []struct {
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
				code: 405,

				contentType: "text/plain",
			},
			comment: "Принимать метрики только по протоколу HTTP методом POST",
		},
		{
			name: "Test 200 OK",
			params: httpParams{
				url:    "/update/gauge/BuckHashSys/1.4",
				method: "POST",
			},
			want: want{
				code: 200,

				contentType: "text/plain",
			},
			comment: "При успешном приёме gauge возвращать http.StatusOK",
		},
		{
			name: "Test 200 OK",
			params: httpParams{
				url:    "/update/counter/PollCount/1",
				method: "POST",
			},
			want: want{
				code: 200,

				contentType: "text/plain",
			},
			comment: "При успешном приёме counter возвращать http.StatusOK",
		},
		{
			name: "Test 404 Not Found",
			params: httpParams{
				url:    "/update/counter/",
				method: "POST",
			},
			want: want{
				code:        404,
				contentType: "text/plain",
			},
			comment: "При попытке передать запрос без имени метрики возвращать http.StatusNotFound",
		},
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.params.method, test.params.url, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			updateHandler(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			//assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
