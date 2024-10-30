package api

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/Mantelijo/spike-backend/internal/data"
)

func NewHttpApi(addr, port string, ds *data.DataStore) *HttpApi {
	return &HttpApi{
		r:         http.NewServeMux(),
		dataStore: ds,
		addr:      addr,
		port:      port,
	}
}

type HttpApi struct {
	addr, port string
	r          *http.ServeMux
	dataStore  *data.DataStore
}

// Start starts the http api server
func (h *HttpApi) Start() error {
	bindAddr := net.JoinHostPort(h.addr, h.port)
	h.registerRoutes()

	slog.Info("starting http api", slog.String("addr", bindAddr))
	return http.ListenAndServe(bindAddr, h.r)
}

func (h *HttpApi) registerRoutes() {
	h.r.HandleFunc("POST /widgets", h.createWidget)
	h.r.HandleFunc("DELETE /widgets/{serial_num}", h.removeWidget)
	h.r.HandleFunc("PUT /widgets/associations", h.associateWidget)
}
