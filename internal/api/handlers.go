package api

import (
	"bytes"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)

	allMetrics := getAllMetrics(s)
	tpl := `{{.}}`
	t, err := template.New("All Metrics").Parse(tpl)
	if err != nil {
		log.Println("an error occured parsing template: %w", err)
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, allMetrics)
	if err != nil {
		log.Println("an error occured converting template data: %w", err)
	}

	html := doc.String()
	_, err = w.Write([]byte(html))
	if err != nil {
		log.Println("an error occured writing to browser: %w", err)
	}
}

func getAllMetrics(s *Server) string {
	html := "<h3>Gauge:</h3>"
	for mName, mValue := range s.store.Gauge {
		html += (mName + ":" + strconv.FormatFloat(mValue, 'f', -1, 64) + "<br>")
	}
	html += "<h3>Counter:</h3>"
	for mName, mValue := range s.store.Counter {
		html += (mName + ":" + strconv.FormatInt(mValue, 10) + "<br>")
	}

	return html
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	mValue, err := s.getSingleMetric(mType, mName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)

	tpl := `{{.}}`
	t, err := template.New("Single Metric").Parse(tpl)
	if err != nil {
		log.Println("an error occured parsing template: %w", err)
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, mValue)
	if err != nil {
		log.Println("an error occured converting template data: %w", err)
	}
	html := doc.String()
	_, err = w.Write([]byte(html))
	if err != nil {
		log.Println("an error occured writing to browser: %w", err)
	}
}

func (s *Server) getSingleMetric(mType, mName string) (string, error) {
	var html string
	var ErrItemNotFound = errors.New("item not found")

	switch mType {
	case "gauge":
		if mValue, ok := s.store.Gauge[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatFloat(mValue, 'f', -1, 64)
			return html, nil
		}
	case "counter":
		if mValue, ok := s.store.Counter[mName]; !ok {
			return "", ErrItemNotFound
		} else {
			html = strconv.FormatInt(mValue, 10)
			return html, nil
		}
	default:
		return "", nil
	}
}

func (s *Server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mType := chi.URLParam(r, "mtype")
	mName := chi.URLParam(r, "mname")
	mValue := chi.URLParam(r, "mvalue")

	if mType == "gauge" || mType == "counter" {
		s.Save(mType, mName, mValue, w)
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}

func (s *Server) Save(mType, mName, mValue string, w http.ResponseWriter) {
	switch mType {
	case "counter":
		s.saveCounter(mName, mValue, w)
	case "gauge":
		s.saveGauge(mName, mValue, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *Server) saveCounter(mName, mValue string, w http.ResponseWriter) {
	vFloat64, err := strconv.ParseFloat(mValue, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vInt64 := int64(vFloat64)
	s.store.Counter[mName] += vInt64
}

func (s *Server) saveGauge(mName, mValue string, w http.ResponseWriter) {
	vFloat64, err := strconv.ParseFloat(mValue, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.store.Gauge[mName] = vFloat64
}

func (a *Agent) SaveMetrics() error {
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	a.store.Gauge["Alloc"] = float64(m.Alloc)
	a.store.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
	a.store.Gauge["Frees"] = float64(m.Frees)
	a.store.Gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
	a.store.Gauge["GCSys"] = float64(m.GCSys)
	a.store.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
	a.store.Gauge["HeapIdle"] = float64(m.HeapIdle)
	a.store.Gauge["HeapInuse"] = float64(m.HeapInuse)
	a.store.Gauge["HeapObjects"] = float64(m.HeapObjects)
	a.store.Gauge["HeapReleased"] = float64(m.HeapReleased)
	a.store.Gauge["HeapSys"] = float64(m.HeapSys)
	a.store.Gauge["LastGC"] = float64(m.LastGC)
	a.store.Gauge["Lookups"] = float64(m.Lookups)
	a.store.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
	a.store.Gauge["MCacheSys"] = float64(m.MCacheSys)
	a.store.Gauge["MSpanInuse"] = float64(m.MSpanInuse)
	a.store.Gauge["MSpanSys"] = float64(m.MSpanSys)
	a.store.Gauge["Mallocs"] = float64(m.Mallocs)
	a.store.Gauge["NextGC"] = float64(m.NextGC)
	a.store.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
	a.store.Gauge["NumGC"] = float64(m.NumGC)
	a.store.Gauge["OtherSys"] = float64(m.OtherSys)
	a.store.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	a.store.Gauge["StackInuse"] = float64(m.StackInuse)
	a.store.Gauge["StackSys"] = float64(m.StackSys)
	a.store.Gauge["Sys"] = float64(m.Sys)
	a.store.Gauge["TotalAlloc"] = float64(m.TotalAlloc)
	a.store.Gauge["RandomValue"] = rand.Float64()
	a.store.Counter["PollCount"]++

	return nil
}

func (a *Agent) SendMetrics() error {
	c := http.Client{Timeout: time.Duration(1) * time.Second}

	for n, v := range a.store.Gauge {
		value := strconv.FormatFloat(v, 'f', -1, 64)

		sendURL, err := url.JoinPath("http://", a.cfg.Host, "update", "gauge", n, value)
		if err != nil {
			log.Println("failed to join path parts: %w", err)
			return nil
		}

		req, err := http.NewRequest(http.MethodPost, sendURL, http.NoBody)
		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Content-Length", "0")

		if err != nil {
			continue
		}

		resp, err := c.Do(req)
		if err != nil {
			log.Println("failed to do a request: %w", err)
			return nil
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Println("failed to close response body: %w", err)
				return
			}
		}()
		if resp.StatusCode != http.StatusOK {
			log.Println("unexpected response code while sending gauge metrics: %w", resp.StatusCode)
			return nil
		}
	}

	for n, v := range a.store.Counter {
		value := strconv.Itoa(int(v))
		sendURL, err := url.JoinPath("http://", a.cfg.Host, "update", "counter", n, value)
		if err != nil {
			log.Println("failed to join path parts: %w", err)
			return nil
		}
		req, err := http.NewRequest(http.MethodPost, sendURL, http.NoBody)
		if err != nil {
			log.Println("failed to do an HTTP POST request: %w", err)
			return nil
		}

		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Content-Length", "0")

		resp, err := c.Do(req)
		if err != nil {
			log.Println("failed to do a request: %w", err)
			return nil
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Println("failed to close response body: %w", err)
				return
			}
		}()
		if resp.StatusCode != http.StatusOK {
			log.Println("unexpected response code while sending counter metrics: %w", resp.StatusCode)
			return nil
		}

		a.store.Counter["PollCount"] = 0
	}
	return nil
}
