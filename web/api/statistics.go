package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"golang.org/x/net/publicsuffix"
)

type statisticsResponse struct {
	DbSize        int64                     `json:"dbsize"`
	Results       int64                     `json:"results"`
	Headers       int64                     `json:"headers"`
	NetworkLogs   int64                     `json:"networklogs"`
	ConsoleLogs   int64                     `json:"consolelogs"`
	ResponseCodes []*statisticsResponseCode `json:"response_code_stats"`
	DomainStats   *domainStatistics         `json:"domain_stats"`
	IPStats       *ipStatistics             `json:"ip_stats"`
	TargetInfo    *targetInformation        `json:"target_info"`
}

type targetInformation struct {
	CompanyName   string `json:"company_name"`
	MainDomain    string `json:"main_domain"`
	LogoPath      string `json:"logo_path,omitempty"`
	ScanStartTime string `json:"scan_start_time"`
	ScanStatus    string `json:"scan_status"`
	Notes         string `json:"notes"`
}

type statisticsResponseCode struct {
	Code  int   `json:"code"`
	Count int64 `json:"count"`
}

type domainStatistics struct {
	UniqueApexDomains int64         `json:"unique_apex_domains"`
	TotalSubdomains   int64         `json:"total_subdomains"`
	TotalDomains      int64         `json:"total_domains"`
	ApexDomains       []*apexDomain `json:"apex_domains"`
}

type apexDomain struct {
	Domain     string       `json:"domain"`
	IsApex     bool         `json:"is_apex"`
	ResultID   uint         `json:"result_id,omitempty"`
	Subdomains []*subdomain `json:"subdomains"`
	Count      int64        `json:"count"`
}

type subdomain struct {
	Domain   string `json:"domain"`
	ResultID uint   `json:"result_id"`
	URL      string `json:"url"`
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
}

type ipStatistics struct {
	UniqueIPs    int64      `json:"unique_ips"`
	TotalResults int64      `json:"total_results"`
	IPList       []*ipEntry `json:"ip_list"`
}

type ipEntry struct {
	IPAddress    string           `json:"ip_address"`
	DomainCount  int64            `json:"domain_count"`
	FirstSeen    string           `json:"first_seen"`
	LastSeen     string           `json:"last_seen"`
	SampleDomain string           `json:"sample_domain"`
	ResultID     uint             `json:"result_id"`
	Domains      []*ipDomainEntry `json:"domains"`
}

type ipDomainEntry struct {
	Domain   string `json:"domain"`
	ResultID uint   `json:"result_id"`
	URL      string `json:"url"`
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
}

// extractApexDomain extracts the apex domain from a URL using the public suffix list
// This properly handles country-code TLDs like .co.uk, .com.au, etc.
func extractApexDomain(inputURL string) string {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return ""
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return ""
	}

	// Use the public suffix list to get the effective TLD (eTLD)
	// This handles complex TLDs like .co.uk, .com.au properly
	etld, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		// If parsing fails, fall back to simple logic for basic cases
		parts := strings.Split(hostname, ".")
		if len(parts) >= 2 {
			return strings.Join(parts[len(parts)-2:], ".")
		}
		return hostname
	}

	return etld
}

// StatisticsHandler returns database statistics
//
//	@Summary		Database statistics
//	@Description	Get database statistics.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	statisticsResponse
//	@Router			/statistics [get]
func (h *ApiHandler) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	response := &statisticsResponse{}

	if err := h.DB.Raw("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()").
		Take(&response.DbSize).Error; err != nil {

		log.Error("an error occured getting database size", "err", err)
		return
	}

	if err := h.DB.Model(&models.Result{}).Count(&response.Results).Error; err != nil {
		log.Error("an error occured counting results", "err", err)
		return
	}

	if err := h.DB.Model(&models.Header{}).Count(&response.Headers).Error; err != nil {
		log.Error("an error occured counting headers", "err", err)
		return
	}

	if err := h.DB.Model(&models.NetworkLog{}).Count(&response.NetworkLogs).Error; err != nil {
		log.Error("an error occured counting network logs", "err", err)
		return
	}

	if err := h.DB.Model(&models.ConsoleLog{}).Count(&response.ConsoleLogs).Error; err != nil {
		log.Error("an error occured counting console logs", "err", err)
		return
	}

	var counts []*statisticsResponseCode
	if err := h.DB.Model(&models.Result{}).
		Select("response_code as code, count(*) as count").
		Group("response_code").Scan(&counts).Error; err != nil {
		log.Error("failed counting response codes", "err", err)
		return
	}

	response.ResponseCodes = counts

	// Calculate domain statistics
	domainStats, err := h.calculateDomainStatistics()
	if err != nil {
		log.Error("failed calculating domain statistics", "err", err)
		return
	}
	response.DomainStats = domainStats

	// Calculate IP statistics
	ipStats, err := h.calculateIPStatistics()
	if err != nil {
		log.Error("failed calculating IP statistics", "err", err)
		return
	}
	response.IPStats = ipStats

	// Get target information from the most recent scan session
	targetInfo, err := h.getTargetInformation()
	if err != nil {
		log.Warn("failed getting target information", "err", err)
		// Don't fail the entire request, just leave target info empty
	} else {
		response.TargetInfo = targetInfo
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// calculateDomainStatistics calculates comprehensive domain statistics
func (h *ApiHandler) calculateDomainStatistics() (*domainStatistics, error) {
	var results []models.Result
	if err := h.DB.Select("id, url").Find(&results).Error; err != nil {
		return nil, err
	}

	// Map to group domains by apex domain
	apexDomainMap := make(map[string]*apexDomain)
	totalSubdomains := int64(0)

	for _, result := range results {
		parsedURL, err := url.Parse(result.URL)
		if err != nil {
			continue
		}

		hostname := parsedURL.Hostname()
		if hostname == "" {
			continue
		}

		apexDomainName := extractApexDomain(result.URL)
		if apexDomainName == "" {
			continue
		}

		// Initialize apex domain if not exists
		if _, exists := apexDomainMap[apexDomainName]; !exists {
			apexDomainMap[apexDomainName] = &apexDomain{
				Domain:     apexDomainName,
				IsApex:     false,
				Subdomains: make([]*subdomain, 0),
				Count:      0,
			}
		}

		apex := apexDomainMap[apexDomainName]
		apex.Count++

		// Check if this is the apex domain itself or a subdomain
		if hostname == apexDomainName {
			// This is the apex domain - add it as a "subdomain" entry for protocol/port grouping

			// Extract protocol and port from URL
			protocol := parsedURL.Scheme
			port := parsedURL.Port()
			if port == "" {
				// Set default ports for common schemes
				switch protocol {
				case "http":
					port = "80"
				case "https":
					port = "443"
				default:
					port = "unknown"
				}
			}

			// Add apex domain as a subdomain entry for protocol/port display
			apex.Subdomains = append(apex.Subdomains, &subdomain{
				Domain:   hostname,
				ResultID: result.ID,
				URL:      result.URL,
				Protocol: protocol,
				Port:     port,
			})

			// Mark as apex and set a result ID if not already set
			apex.IsApex = true
			if apex.ResultID == 0 {
				apex.ResultID = result.ID
			}
		} else {
			// This is a subdomain
			totalSubdomains++

			// Extract protocol and port from URL
			protocol := parsedURL.Scheme
			port := parsedURL.Port()
			if port == "" {
				// Set default ports for common schemes
				switch protocol {
				case "http":
					port = "80"
				case "https":
					port = "443"
				default:
					port = "unknown"
				}
			}

			apex.Subdomains = append(apex.Subdomains, &subdomain{
				Domain:   hostname,
				ResultID: result.ID,
				URL:      result.URL,
				Protocol: protocol,
				Port:     port,
			})
		}
	}

	// Convert map to slice and sort by count (descending)
	apexDomains := make([]*apexDomain, 0, len(apexDomainMap))
	for _, apex := range apexDomainMap {
		apexDomains = append(apexDomains, apex)
	}

	// Simple bubble sort by count (descending)
	for i := 0; i < len(apexDomains)-1; i++ {
		for j := 0; j < len(apexDomains)-i-1; j++ {
			if apexDomains[j].Count < apexDomains[j+1].Count {
				apexDomains[j], apexDomains[j+1] = apexDomains[j+1], apexDomains[j]
			}
		}
	}

	return &domainStatistics{
		UniqueApexDomains: int64(len(apexDomainMap)),
		TotalSubdomains:   totalSubdomains,
		TotalDomains:      int64(len(results)),
		ApexDomains:       apexDomains,
	}, nil
}

// calculateIPStatistics calculates comprehensive IP address statistics
func (h *ApiHandler) calculateIPStatistics() (*ipStatistics, error) {
	var results []models.Result
	if err := h.DB.Select("id, url, ip_address, probed_at").Where("ip_address != ''").Find(&results).Error; err != nil {
		return nil, err
	}

	// Map to group results by IP address
	ipMap := make(map[string]*ipEntry)

	for _, result := range results {
		if result.IPAddress == "" {
			continue
		}

		parsedURL, err := url.Parse(result.URL)
		if err != nil {
			continue
		}

		hostname := parsedURL.Hostname()
		if hostname == "" {
			continue
		}

		// Extract protocol and port from URL
		protocol := parsedURL.Scheme
		port := parsedURL.Port()
		if port == "" {
			// Set default ports for common schemes
			switch protocol {
			case "http":
				port = "80"
			case "https":
				port = "443"
			default:
				port = "unknown"
			}
		}

		// Initialize IP entry if not exists
		if _, exists := ipMap[result.IPAddress]; !exists {
			ipMap[result.IPAddress] = &ipEntry{
				IPAddress:    result.IPAddress,
				DomainCount:  0,
				FirstSeen:    result.ProbedAt.Format("2006-01-02 15:04:05"),
				LastSeen:     result.ProbedAt.Format("2006-01-02 15:04:05"),
				SampleDomain: hostname,
				ResultID:     result.ID,
				Domains:      make([]*ipDomainEntry, 0),
			}
		}

		ipEntry := ipMap[result.IPAddress]
		ipEntry.DomainCount++

		// Add domain entry
		ipEntry.Domains = append(ipEntry.Domains, &ipDomainEntry{
			Domain:   hostname,
			ResultID: result.ID,
			URL:      result.URL,
			Protocol: protocol,
			Port:     port,
		})

		// Update first/last seen times
		currentProbed := result.ProbedAt.Format("2006-01-02 15:04:05")
		if currentProbed < ipEntry.FirstSeen {
			ipEntry.FirstSeen = currentProbed
		}
		if currentProbed > ipEntry.LastSeen {
			ipEntry.LastSeen = currentProbed
		}
	}

	// Convert map to slice and sort by domain count (descending)
	ipList := make([]*ipEntry, 0, len(ipMap))
	for _, ip := range ipMap {
		ipList = append(ipList, ip)
	}

	// Simple bubble sort by domain count (descending)
	for i := 0; i < len(ipList)-1; i++ {
		for j := 0; j < len(ipList)-i-1; j++ {
			if ipList[j].DomainCount < ipList[j+1].DomainCount {
				ipList[j], ipList[j+1] = ipList[j+1], ipList[j]
			}
		}
	}

	return &ipStatistics{
		UniqueIPs:    int64(len(ipMap)),
		TotalResults: int64(len(results)),
		IPList:       ipList,
	}, nil
}

// getTargetInformation retrieves target information from the most recent scan session
func (h *ApiHandler) getTargetInformation() (*targetInformation, error) {
	var session models.ScanSession
	if err := h.DB.Order("start_time DESC").First(&session).Error; err != nil {
		return nil, err
	}

	return &targetInformation{
		CompanyName:   session.CompanyName,
		MainDomain:    session.MainDomain,
		LogoPath:      session.LogoPath,
		ScanStartTime: session.StartTime.Format("2006-01-02 15:04:05"),
		ScanStatus:    session.Status,
		Notes:         session.Notes,
	}, nil
}
