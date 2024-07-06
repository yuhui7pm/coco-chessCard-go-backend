package metrics

import (
	"github.com/arl/statsviz"
	"net/http"
)

// Server可视化 实时监控
func Server(add string) error {
	mux := http.NewServeMux()

	if err := statsviz.Register(mux); err != nil {
		return err
	}

	if err := http.ListenAndServe(add, mux); err != nil {
		return err
	}

	return nil
}
