package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Mantelijo/spike-backend/internal/dto"
)

type createWidgetRequest struct {
	Name         string   `json:"name"`
	SerialNumber string   `json:"serial_number"`
	Ports        []string `json:"ports"`
}

// parseRequest is a helper to quickly read and parse requests into structs and
// prints errors if action was not performed successfuly. Do not use it in
// production. Return true if request was parsed successfully.
func parseRequest[T any](r *http.Request, w http.ResponseWriter, v *T, handlerName string) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("reading request body",
			slog.String("handler", handlerName),
			slog.Any("error", err),
		)
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return false
	}

	if err := json.Unmarshal(body, v); err != nil {
		slog.Error("parsing request body",
			slog.String("handler", handlerName),
			slog.Any("error", err),
		)
		http.Error(w, "could not parse request body", http.StatusBadRequest)
		return false
	}
	return true
}

func (api *HttpApi) createWidget(w http.ResponseWriter, r *http.Request) {
	req := &createWidgetRequest{}
	if !parseRequest(r, w, req, "createWidget") {
		return
	}
	// TODO: validate request

	err := api.dataStore.CreateWidget(&dto.Widget{
		Name:         req.Name,
		SerialNumber: req.SerialNumber,
		PortBitmap:   dto.WidgetPortFromStrings(req.Ports),
	})
	if err != nil {
		slog.Error("creating widget",
			slog.Any("error", err),
		)
		http.Error(w, "could not create widget", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Widget created")
}

func (api *HttpApi) removeWidget(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
	// DB DELETE widget with automatic widget conns deletion
	// Delete from cache
}

type associateWidgetRequest struct {
	PortType string `json:"port_type"`
	WidgetSn string `json:"widget_serial_num"`
	PeerSn   string `json:"peer_widget_serial_num"`
}

func (api *HttpApi) associateWidget(w http.ResponseWriter, r *http.Request) {
	req := &associateWidgetRequest{}
	if !parseRequest(r, w, req, "associateWidget") {
		return
	}

	wc := &dto.WidgetConnections{
		SerialNumber: req.WidgetSn,
	}
	port := dto.WidgetPortFromString(req.PortType)
	switch port {
	case dto.P:
		wc.P_PeerSerialNumber = req.PeerSn
	case dto.Q:
		wc.Q_PeerSerialNumber = req.PeerSn
	case dto.R:
		wc.R_PeerSerialNumber = req.PeerSn
	default:
		http.Error(w, "invalid port type: options are P, R or Q", http.StatusBadRequest)
	}

	err := api.dataStore.AssociateConnections(wc)
	if err != nil {
		slog.Error("setting widget association",
			slog.Any("error", err),
		)
		http.Error(w, "could not create widget association", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Widget association updated")
}
