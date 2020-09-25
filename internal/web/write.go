package web

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/snappy"
	"github.com/ledyba/prometheus_sql_adapter/internal/repo"
	"github.com/prometheus/prometheus/prompb"
)

func Write(w http.ResponseWriter, r *http.Request) {
	in, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Failed to read body: %v", err)))
		return
	}
	out, err := snappy.Decode(nil, in)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Failed to decode body: %v", err)))
		return
	}
	req := prompb.WriteRequest{}
	err = req.Unmarshal(out)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Failed to parse request: %v", err)))
		return
	}
	err = repo.Write(&req)
	if err != nil {
		renderError(w, r, err, nil)
		return
	}
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}