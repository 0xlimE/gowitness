package api

import (
	"encoding/json"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
)

// ScanSessionResponse represents scan session information
type ScanSessionResponse struct {
	ID          uint   `json:"id"`
	CompanyName string `json:"company_name"`
	MainDomain  string `json:"main_domain"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time,omitempty"`
	Status      string `json:"status"`
	Notes       string `json:"notes"`
}

// ScanSessionsHandler handles requests for scan session information
//
//	@Summary		Get scan sessions information
//	@Description	Returns information about all scan sessions including target details
//	@Tags			Scan Sessions
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	ScanSessionResponse
//	@Router			/scan-sessions [get]
func (h *ApiHandler) ScanSessionsHandler(w http.ResponseWriter, r *http.Request) {
	var sessions []models.ScanSession
	if err := h.DB.Find(&sessions).Error; err != nil {
		log.Error("failed to get scan sessions", "err", err)
		http.Error(w, "Error retrieving scan sessions", http.StatusInternalServerError)
		return
	}

	response := make([]ScanSessionResponse, len(sessions))
	for i, session := range sessions {
		response[i] = ScanSessionResponse{
			ID:          session.ID,
			CompanyName: session.CompanyName,
			MainDomain:  session.MainDomain,
			StartTime:   session.StartTime.Format("2006-01-02 15:04:05"),
			Status:      session.Status,
			Notes:       session.Notes,
		}

		if session.EndTime != nil {
			response[i].EndTime = session.EndTime.Format("2006-01-02 15:04:05")
		}
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Error("failed to marshal scan sessions response", "err", err)
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
