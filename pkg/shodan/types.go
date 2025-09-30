package shodan

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ShodanTime is a custom time type to handle Shodan's timestamp format
type ShodanTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for Shodan timestamps
func (st *ShodanTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}

	// Try different timestamp formats that Shodan might use
	formats := []string{
		"2006-01-02T15:04:05.000000",      // Shodan's typical format
		"2006-01-02T15:04:05",             // Without microseconds
		time.RFC3339,                      // Standard RFC3339
		time.RFC3339Nano,                  // RFC3339 with nanoseconds
		"2006-01-02T15:04:05Z",           // UTC format
		"2006-01-02T15:04:05.000000Z",    // UTC with microseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			st.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse time %q", s)
}

// MarshalJSON implements custom JSON marshaling
func (st ShodanTime) MarshalJSON() ([]byte, error) {
	if st.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(st.Time.Format(time.RFC3339))
}

// Host represents a host record from Shodan API
type Host struct {
	IP           string     `json:"ip_str"`
	Organization string     `json:"org,omitempty"`
	ISP          string     `json:"isp,omitempty"`
	ASN          string     `json:"asn,omitempty"`
	Country      string     `json:"country_name,omitempty"`
	CountryCode  string     `json:"country_code,omitempty"`
	City         string     `json:"city,omitempty"`
	Region       string     `json:"region_code,omitempty"`
	Postal       string     `json:"postal_code,omitempty"`
	Latitude     float64    `json:"latitude,omitempty"`
	Longitude    float64    `json:"longitude,omitempty"`
	Ports        []int      `json:"ports,omitempty"`
	Hostnames    []string   `json:"hostnames,omitempty"`
	Domains      []string   `json:"domains,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	OS           string     `json:"os,omitempty"`
	LastUpdate   ShodanTime `json:"last_update,omitempty"`
	Data         []Service  `json:"data,omitempty"`
	Vulns        []string   `json:"vulns,omitempty"`
}

// Service represents a service running on a port
type Service struct {
	Port      int               `json:"port"`
	Transport string            `json:"transport"`
	Product   string            `json:"product,omitempty"`
	Version   string            `json:"version,omitempty"`
	Banner    string            `json:"data,omitempty"`
	Timestamp ShodanTime        `json:"timestamp,omitempty"`
	Location  ServiceLocation   `json:"location,omitempty"`
	HTTP      *HTTPInfo         `json:"http,omitempty"`
	SSL       *SSLInfo          `json:"ssl,omitempty"`
	Opts      map[string]string `json:"opts,omitempty"`
}

// ServiceLocation represents the geolocation of a service
type ServiceLocation struct {
	Country     string  `json:"country_name,omitempty"`
	CountryCode string  `json:"country_code,omitempty"`
	City        string  `json:"city,omitempty"`
	Region      string  `json:"region_code,omitempty"`
	Postal      string  `json:"postal_code,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

// HTTPInfo represents HTTP-specific information
type HTTPInfo struct {
	Status     int               `json:"status,omitempty"`
	Title      string            `json:"title,omitempty"`
	Server     string            `json:"server,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	HTML       string            `json:"html,omitempty"`
	Redirects  []string          `json:"redirects,omitempty"`
	Components map[string]string `json:"components,omitempty"`
}

// SSLInfo represents SSL/TLS certificate information
type SSLInfo struct {
	Versions    []string         `json:"versions,omitempty"`
	Cipher      SSLCipher        `json:"cipher,omitempty"`
	Certificate SSLCertificate   `json:"cert,omitempty"`
	Chain       []SSLCertificate `json:"chain,omitempty"`
}

// SSLCipher represents SSL cipher information
type SSLCipher struct {
	Version string `json:"version,omitempty"`
	Bits    int    `json:"bits,omitempty"`
	Name    string `json:"name,omitempty"`
}

// SSLCertificate represents SSL certificate information
type SSLCertificate struct {
	Subject     SSLSubject `json:"subject,omitempty"`
	Issuer      SSLSubject `json:"issuer,omitempty"`
	Serial      string     `json:"serial,omitempty"`
	Fingerprint string     `json:"fingerprint,omitempty"`
	Expired     bool       `json:"expired,omitempty"`
	ValidFrom   ShodanTime `json:"valid_from,omitempty"`
	ValidUntil  ShodanTime `json:"valid_until,omitempty"`
}

// SSLSubject represents SSL certificate subject/issuer information
type SSLSubject struct {
	CN string `json:"CN,omitempty"`
	C  string `json:"C,omitempty"`
	O  string `json:"O,omitempty"`
	OU string `json:"OU,omitempty"`
	L  string `json:"L,omitempty"`
	ST string `json:"ST,omitempty"`
}

// APIInfo represents Shodan API account information
type APIInfo struct {
	QueryCredits int    `json:"query_credits"`
	ScanCredits  int    `json:"scan_credits"`
	Telnet       bool   `json:"telnet"`
	Plan         string `json:"plan"`
	HTTPS        bool   `json:"https"`
	Unlocked     bool   `json:"unlocked"`
}
