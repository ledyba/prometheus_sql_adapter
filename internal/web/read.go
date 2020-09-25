package web

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

func Read(w http.ResponseWriter, r *http.Request) {
	in, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Failed to read body", zap.Error(err))
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Failed to read body: %v", err)))
		return
	}
	out, err := snappy.Decode(nil, in)
	if err != nil {
		log.Error("Failed to decode body", zap.Error(err))
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Failed to decode body: %v", err)))
		return
	}
	req := prompb.ReadRequest{}
	req.ProtoMessage()
	err = req.Unmarshal(out)
	if err != nil {
		log.Error("Failed to parse request", zap.Error(err))
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Failed to parse request: %v", err)))
		return
	}
}