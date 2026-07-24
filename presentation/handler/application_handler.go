// presentation/handler/application_handler.go
package handler

import (
	"net/http"

	"github.com/shinchi-pmtech/ringi/domain/application"
	"github.com/shinchi-pmtech/ringi/usecase"
)

type ApplicationHandler struct {
	approve *usecase.ApproveApplication
}

func NewApplicationHandler(approve *usecase.ApproveApplication) *ApplicationHandler {
	return &ApplicationHandler{approve: approve}
}

// Approve は POST /applications/approve?id=xxx&approver=yyy を処理する(説明用の最小実装)
func (h *ApplicationHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := application.ApplicationID(r.URL.Query().Get("id"))
	approverID := application.ApproverID(r.URL.Query().Get("approver"))

	if err := h.approve.Execute(id, approverID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
