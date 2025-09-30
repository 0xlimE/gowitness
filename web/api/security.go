package api

import (
	"encoding/json"
	"net/http"
)

// SecurityStatus represents the current security configuration
type SecurityStatus struct {
	PasswordEnabled bool   `json:"password_enabled"`
	ServerInfo      string `json:"server_info,omitempty"`
}

// SecurityStatusHandler returns the current security status
// @Summary Get Security Status
// @Description Get the current security configuration of the server
// @Tags security
// @Accept json
// @Produce json
// @Success 200 {object} SecurityStatus
// @Router /security/status [get]
func (api *ApiHandler) SecurityStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Check if we have password protection enabled by looking for the auth cookie requirement
	// In a real implementation, this would check server configuration

	// For now, we'll check if the request has an auth cookie to determine if password protection is active
	_, err := r.Cookie("gowitness_auth")
	hasPassword := err == nil // If cookie exists, password protection is likely enabled

	status := SecurityStatus{
		PasswordEnabled: hasPassword,
		ServerInfo:      "gowitness v3 web interface",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
