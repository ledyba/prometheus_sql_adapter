package web

import (
	"fmt"
	"net/http"
)

func renderError(w http.ResponseWriter, r *http.Request, err error, more interface{}) {
	w.WriteHeader(502)
	w.Header().Set("Content-Type", "text/plain")
	body := fmt.Sprintf(`Error: %v

[[Info]]
 - Protocol: %s
`, err, r.Proto)
	if s, ok := more.(string); ok {
		body += fmt.Sprintf(`
[[Additional]]
%s`, s)
	}
	_, _ = w.Write([]byte(body))
}
