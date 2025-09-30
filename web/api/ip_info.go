package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/gorm"
)

// IPAPIResponse represents response from ip-api.com
type IPAPIResponse struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Message     string  `json:"message,omitempty"`
}

// NaabuResult represents naabu port scan result
type NaabuResult struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// IPPortInfo represents port information for an IP
type IPPortInfo struct {
	ID            uint   `json:"id"`
	Port          int    `json:"port"`
	Protocol      string `json:"protocol"`
	Service       string `json:"service"`
	State         string `json:"state"`
	Banner        string `json:"banner"`
	ScanSessionID *uint  `json:"scan_session_id,omitempty"`
	DiscoveredAt  string `json:"discovered_at"`
	IsCDN         bool   `json:"is_cdn"`
	CDNName       string `json:"cdn_name"`
	CDNDetected   bool   `json:"cdn_detected"`
	OriginalHost  string `json:"original_host"`
}

// DomainInfo represents domain information associated with an IP
type DomainInfo struct {
	ID             uint   `json:"id"`
	URL            string `json:"url"`
	FinalURL       string `json:"final_url"`
	Title          string `json:"title"`
	ResponseCode   int    `json:"response_code"`
	ResponseReason string `json:"response_reason"`
	Protocol       string `json:"protocol"`
	Screenshot     string `json:"screenshot"`
	Filename       string `json:"file_name"`
	Failed         bool   `json:"failed"`
	FailedReason   string `json:"failed_reason"`
	ProbedAt       string `json:"probed_at"`
	ScanSessionID  *uint  `json:"scan_session_id,omitempty"`
}

// IPInfoResponse represents the complete response for IP information
type IPInfoResponse struct {
	IPAddress    string       `json:"ip_address"`
	OpenPorts    []IPPortInfo `json:"open_ports"`
	TotalPorts   int          `json:"total_ports"`
	Domains      []DomainInfo `json:"domains"`
	TotalDomains int          `json:"total_domains"`
	ScanSessions []uint       `json:"scan_sessions"` // List of scan session IDs this IP was seen in

	// Enhanced Shodan information
	ShodanInfo *ShodanInfo `json:"shodan_info,omitempty"`
}

// ShodanInfo represents Shodan data for an IP address
type ShodanInfo struct {
	Organization  string   `json:"organization,omitempty"`
	ISP           string   `json:"isp,omitempty"`
	ASN           string   `json:"asn,omitempty"`
	Country       string   `json:"country,omitempty"`
	CountryCode   string   `json:"country_code,omitempty"`
	City          string   `json:"city,omitempty"`
	Region        string   `json:"region,omitempty"`
	Postal        string   `json:"postal,omitempty"`
	Latitude      float64  `json:"latitude,omitempty"`
	Longitude     float64  `json:"longitude,omitempty"`
	OS            string   `json:"os,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Ports         []int    `json:"ports,omitempty"`
	Hostnames     []string `json:"hostnames,omitempty"`
	ShodanDomains []string `json:"shodan_domains,omitempty"`
	Vulns         []string `json:"vulns,omitempty"`
	LastUpdate    string   `json:"last_update,omitempty"`
	UpdatedAt     string   `json:"updated_at,omitempty"`
}

// fetchIPAPIData fetches geolocation data from ip-api.com as fallback
func (h *ApiHandler) fetchIPAPIData(ip string) (*IPAPIResponse, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query", ip)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from IP-API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IP-API response: %w", err)
	}

	var ipApiResp IPAPIResponse
	if err := json.Unmarshal(body, &ipApiResp); err != nil {
		return nil, fmt.Errorf("failed to parse IP-API response: %w", err)
	}

	if ipApiResp.Status == "fail" {
		return nil, fmt.Errorf("IP-API error: %s", ipApiResp.Message)
	}

	return &ipApiResp, nil
}

// runNaabuScan runs naabu port scanner for the given IP
func (h *ApiHandler) runNaabuScan(ip string) ([]int, error) {
	// Check if naabu is available
	if _, err := exec.LookPath("naabu"); err != nil {
		return nil, fmt.Errorf("naabu not found: %w", err)
	}

	// Run naabu with top 100 ports and JSON output
	cmd := exec.Command("naabu", "-host", ip, "-top-ports", "100", "-json", "-silent")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("naabu execution failed: %w", err)
	}

	// Parse naabu output (JSON lines)
	ports := []int{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var result NaabuResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			log.Warn("failed to parse naabu line", "line", line, "err", err)
			continue
		}

		if result.IP == ip {
			ports = append(ports, result.Port)
		}
	}

	return ports, nil
}

// isValidIPAddress checks if the given string is a valid IP address
func isValidIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}

// storeFallbackIPData stores IP information gathered from fallback sources
func (h *ApiHandler) storeFallbackIPData(ipAddress string, ipApiData *IPAPIResponse, ports []int) error {
	// Check if IP info already exists
	var existingIPInfo models.IPInfo
	if err := h.DB.Where("ip_address = ?", ipAddress).First(&existingIPInfo).Error; err == nil {
		// Already exists, don't overwrite Shodan data
		log.Debug("IP info already exists, not overwriting", "ip", ipAddress)
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing IP info: %w", err)
	}

	// Create new IP info from IP-API data
	ipInfo := models.IPInfo{
		IPAddress:    ipAddress,
		Organization: ipApiData.Org,
		ISP:          ipApiData.ISP,
		ASN:          ipApiData.AS,
		Country:      ipApiData.Country,
		CountryCode:  ipApiData.CountryCode,
		City:         ipApiData.City,
		Region:       ipApiData.RegionName,
		Postal:       ipApiData.Zip,
		Latitude:     ipApiData.Lat,
		Longitude:    ipApiData.Lon,
		LastUpdate:   time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Set ports from naabu scan
	if len(ports) > 0 {
		if err := ipInfo.SetPorts(ports); err != nil {
			log.Warn("failed to set ports for IP info", "ip", ipAddress, "err", err)
		}
	}

	// Save to database
	if err := h.DB.Create(&ipInfo).Error; err != nil {
		return fmt.Errorf("failed to save fallback IP info: %w", err)
	}

	log.Info("stored fallback IP data", "ip", ipAddress, "source", "ip-api+naabu")
	return nil
}

// IPInfoHandler handles IP information requests
//
//	@Summary		Get information about an IP address
//	@Description	Returns comprehensive information about an IP address including open ports and associated domains
//	@Tags			IP Information
//	@Accept			json
//	@Produce		json
//	@Param			ip	path		string	true	"The IP address to get information for"
//	@Success		200	{object}	IPInfoResponse
//	@Router			/ip/{ip} [get]
func (h *ApiHandler) IPInfoHandler(w http.ResponseWriter, r *http.Request) {
	ipAddress := chi.URLParam(r, "ip")
	if ipAddress == "" {
		http.Error(w, "IP address parameter is required", http.StatusBadRequest)
		return
	}

	var response IPInfoResponse
	response.IPAddress = ipAddress

	// Get open ports for this IP
	var ipPorts []models.IPPort
	if err := h.DB.Where("ip_address = ?", ipAddress).Find(&ipPorts).Error; err != nil {
		log.Error("failed to get IP ports", "err", err, "ip", ipAddress)
		http.Error(w, "Error retrieving port information", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response.OpenPorts = make([]IPPortInfo, len(ipPorts))
	scanSessionSet := make(map[uint]bool)

	for i, port := range ipPorts {
		response.OpenPorts[i] = IPPortInfo{
			ID:            port.ID,
			Port:          port.Port,
			Protocol:      port.Protocol,
			Service:       port.Service,
			State:         port.State,
			Banner:        port.Banner,
			ScanSessionID: port.ScanSessionID,
			DiscoveredAt:  port.DiscoveredAt.Format("2006-01-02 15:04:05"),
			IsCDN:         port.IsCDN,
			CDNName:       port.CDNName,
			CDNDetected:   port.CDNDetected,
			OriginalHost:  port.OriginalHost,
		}

		// Track scan sessions
		if port.ScanSessionID != nil {
			scanSessionSet[*port.ScanSessionID] = true
		}
	}
	response.TotalPorts = len(ipPorts)

	// Get domains associated with this IP
	var domains []models.Result
	if err := h.DB.Where("ip_address = ?", ipAddress).Find(&domains).Error; err != nil {
		log.Error("failed to get domains for IP", "err", err, "ip", ipAddress)
		http.Error(w, "Error retrieving domain information", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response.Domains = make([]DomainInfo, len(domains))
	for i, domain := range domains {
		response.Domains[i] = DomainInfo{
			ID:             domain.ID,
			URL:            domain.URL,
			FinalURL:       domain.FinalURL,
			Title:          domain.Title,
			ResponseCode:   domain.ResponseCode,
			ResponseReason: domain.ResponseReason,
			Protocol:       domain.Protocol,
			Screenshot:     domain.Screenshot,
			Filename:       domain.Filename,
			Failed:         domain.Failed,
			FailedReason:   domain.FailedReason,
			ProbedAt:       domain.ProbedAt.Format("2006-01-02 15:04:05"),
			ScanSessionID:  domain.ScanSessionID,
		}

		// Track scan sessions from domains too
		if domain.ScanSessionID != nil {
			scanSessionSet[*domain.ScanSessionID] = true
		}
	}
	response.TotalDomains = len(domains)

	// Convert scan session set to slice
	response.ScanSessions = make([]uint, 0, len(scanSessionSet))
	for sessionID := range scanSessionSet {
		response.ScanSessions = append(response.ScanSessions, sessionID)
	}

	// Get Shodan information for this IP, with fallback to IP-API and naabu
	var ipInfo models.IPInfo
	needsFallback := false

	if err := h.DB.Where("ip_address = ?", ipAddress).First(&ipInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			needsFallback = true
		} else {
			// Log error but don't fail the request
			log.Warn("failed to get IP info from database", "err", err, "ip", ipAddress)
			needsFallback = true
		}
	} else {
		// Check if we have minimal data (might be from fallback source)
		if ipInfo.Organization == "" && ipInfo.ISP == "" && ipInfo.Country == "" {
			needsFallback = true
		}
	}

	// If we need fallback data, try to gather it
	if needsFallback {
		log.Info("attempting fallback IP intelligence gathering", "ip", ipAddress)

		// Validate IP address
		if !isValidIPAddress(ipAddress) {
			log.Warn("invalid IP address for fallback lookup", "ip", ipAddress)
		} else {
			// Try IP-API for geolocation
			ipApiData, err := h.fetchIPAPIData(ipAddress)
			if err != nil {
				log.Warn("failed to fetch IP-API data", "ip", ipAddress, "err", err)
			}

			// Try naabu for port scanning (only if no ports already exist)
			var ports []int
			var existingPorts []models.IPPort
			if err := h.DB.Where("ip_address = ?", ipAddress).Find(&existingPorts).Error; err == nil && len(existingPorts) == 0 {
				if scanPorts, err := h.runNaabuScan(ipAddress); err != nil {
					log.Warn("failed to run naabu scan", "ip", ipAddress, "err", err)
				} else {
					ports = scanPorts
					log.Info("naabu scan completed", "ip", ipAddress, "ports_found", len(ports))
				}
			}

			// Store fallback data if we got any
			if ipApiData != nil {
				if err := h.storeFallbackIPData(ipAddress, ipApiData, ports); err != nil {
					log.Error("failed to store fallback IP data", "ip", ipAddress, "err", err)
				} else {
					// Re-fetch the newly stored data
					if err := h.DB.Where("ip_address = ?", ipAddress).First(&ipInfo).Error; err != nil {
						log.Warn("failed to re-fetch stored IP info", "err", err, "ip", ipAddress)
					}
				}
			}
		}
	}

	// If we have IP info (either from Shodan or fallback), populate response
	if ipInfo.IPAddress != "" {
		shodanInfo := &ShodanInfo{
			Organization: ipInfo.Organization,
			ISP:          ipInfo.ISP,
			ASN:          ipInfo.ASN,
			Country:      ipInfo.Country,
			CountryCode:  ipInfo.CountryCode,
			City:         ipInfo.City,
			Region:       ipInfo.Region,
			Postal:       ipInfo.Postal,
			Latitude:     ipInfo.Latitude,
			Longitude:    ipInfo.Longitude,
			OS:           ipInfo.OS,
			LastUpdate:   ipInfo.LastUpdate.Format("2006-01-02 15:04:05"),
			UpdatedAt:    ipInfo.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// Get array fields using helper methods
		if tags, err := ipInfo.GetTags(); err == nil {
			shodanInfo.Tags = tags
		}
		if ports, err := ipInfo.GetPorts(); err == nil {
			shodanInfo.Ports = ports
		}
		if hostnames, err := ipInfo.GetHostnames(); err == nil {
			shodanInfo.Hostnames = hostnames
		}
		if domains, err := ipInfo.GetDomains(); err == nil {
			shodanInfo.ShodanDomains = domains
		}
		if vulns, err := ipInfo.GetVulns(); err == nil {
			shodanInfo.Vulns = vulns
		}

		response.ShodanInfo = shodanInfo
	}

	// Return JSON response
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Error("failed to marshal IP info response", "err", err)
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
