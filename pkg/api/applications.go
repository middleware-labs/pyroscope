package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pyroscope-io/pyroscope/pkg/model/appmetadata"
	"github.com/pyroscope-io/pyroscope/pkg/server/httputils"
)

type ApplicationListerAndDeleter interface {
	List(context.Context) ([]appmetadata.ApplicationMetadata, error)
	Delete(ctx context.Context, name string) error
}

type ApplicationsHandler struct {
	svc       ApplicationListerAndDeleter
	httpUtils httputils.Utils
}

func NewApplicationsHandler(svc ApplicationListerAndDeleter, httpUtils httputils.Utils) *ApplicationsHandler {
	return &ApplicationsHandler{
		svc:       svc,
		httpUtils: httpUtils,
	}
}

func (h *ApplicationsHandler) GetApps(w http.ResponseWriter, r *http.Request) {
	apps, err := h.svc.List(r.Context())

	filteredApps := make([]appmetadata.ApplicationMetadata, 0)

	for _, app := range apps {
		if strings.HasPrefix(app.FQName, r.URL.Query().Get("uid")) {
			filteredApps = append(filteredApps, app)
		}
	}

	if err != nil {
		h.httpUtils.HandleError(r, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.httpUtils.WriteResponseJSON(r, w, filteredApps)
}

type DeleteAppInput struct {
	Name string `json:"name"`
}

func (h *ApplicationsHandler) DeleteApp(w http.ResponseWriter, r *http.Request) {
	var payload DeleteAppInput

	fmt.Println("calling delete app")
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		h.httpUtils.HandleError(r, w, httputils.JSONError{Err: err})
		return
	}

	err = h.svc.Delete(r.Context(), payload.Name)
	if err != nil {
		h.httpUtils.HandleError(r, w, err)
		return
	}
}
