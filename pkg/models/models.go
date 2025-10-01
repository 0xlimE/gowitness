package models

import (
	"encoding/json"
	"time"
)

// RequestType are network log types
type RequestType int

const (
	HTTP RequestType = 0
	WS
)

// Result is a Gowitness result
type Result struct {
	ID uint `json:"id" gorm:"primarykey"`

	URL                   string    `json:"url"`
	IPAddress             string    `json:"ip_address" gorm:"index"`
	ScanSessionID         *uint     `json:"scan_session_id,omitempty" gorm:"index"`
	ProbedAt              time.Time `json:"probed_at"`
	FinalURL              string    `json:"final_url"`
	ResponseCode          int       `json:"response_code"`
	ResponseReason        string    `json:"response_reason"`
	Protocol              string    `json:"protocol"`
	ContentLength         int64     `json:"content_length"`
	HTML                  string    `json:"html" gorm:"index"`
	Title                 string    `json:"title" gorm:"index"`
	PerceptionHash        string    `json:"perception_hash" gorm:"index"`
	PerceptionHashGroupId uint      `json:"perception_hash_group_id" gorm:"index"`
	Screenshot            string    `json:"screenshot"`

	// Name of the screenshot file
	Filename string `json:"file_name"`
	IsPDF    bool   `json:"is_pdf"`

	// Failed flag set if the result should be considered failed
	Failed       bool   `json:"failed"`
	FailedReason string `json:"failed_reason"`

	TLS          TLS          `json:"tls" gorm:"constraint:OnDelete:CASCADE"`
	Technologies []Technology `json:"technologies" gorm:"constraint:OnDelete:CASCADE"`

	Headers []Header     `json:"headers" gorm:"constraint:OnDelete:CASCADE"`
	Network []NetworkLog `json:"network" gorm:"constraint:OnDelete:CASCADE"`
	Console []ConsoleLog `json:"console" gorm:"constraint:OnDelete:CASCADE"`
	Cookies []Cookie     `json:"cookies" gorm:"constraint:OnDelete:CASCADE"`
}

func (r *Result) HeaderMap() map[string][]string {
	headersMap := make(map[string][]string)

	for _, header := range r.Headers {
		headersMap[header.Key] = []string{header.Value}
	}

	return headersMap
}

type TLS struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"resultid"`

	Protocol                 string       `json:"protocol"`
	KeyExchange              string       `json:"key_exchange"`
	Cipher                   string       `json:"cipher"`
	SubjectName              string       `json:"subject_name"`
	SanList                  []TLSSanList `json:"san_list" gorm:"constraint:OnDelete:CASCADE"`
	Issuer                   string       `json:"issuer"`
	ValidFrom                time.Time    `json:"valid_from"`
	ValidTo                  time.Time    `json:"valid_to"`
	ServerSignatureAlgorithm int64        `json:"server_signature_algorithm"`
	EncryptedClientHello     bool         `json:"encrypted_client_hello"`
}

type TLSSanList struct {
	ID    uint `json:"id" gorm:"primarykey"`
	TLSID uint `json:"tls_id"`

	Value string `json:"value"`
}

type Technology struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	Value string `json:"value" gorm:"index"`
}

type Header struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	Key   string `json:"key"`
	Value string `json:"value" gorm:"index"`
}

type NetworkLog struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	RequestType RequestType `json:"request_type"`
	StatusCode  int64       `json:"status_code"`
	URL         string      `json:"url"`
	RemoteIP    string      `json:"remote_ip"`
	MIMEType    string      `json:"mime_type"`
	Time        time.Time   `json:"time"`
	Content     []byte      `json:"content"`
	Error       string      `json:"error"`
}

type ConsoleLog struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	Type  string `json:"type"`
	Value string `json:"value" gorm:"index"`
}

type Cookie struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	Name         string    `json:"name"`
	Value        string    `json:"value"`
	Domain       string    `json:"domain"`
	Path         string    `json:"path"`
	Expires      time.Time `json:"expires"`
	Size         int64     `json:"size"`
	HTTPOnly     bool      `json:"http_only"`
	Secure       bool      `json:"secure"`
	Session      bool      `json:"session"`
	Priority     string    `json:"priority"`
	SourceScheme string    `json:"source_scheme"`
	SourcePort   int64     `json:"source_port"`
}

// ScanSession represents a scan session for a target company
type ScanSession struct {
	ID          uint       `json:"id" gorm:"primarykey"`
	CompanyName string     `json:"company_name" gorm:"index"`
	MainDomain  string     `json:"main_domain" gorm:"index"`
	LogoPath    string     `json:"logo_path,omitempty"` // Path to company logo file
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Status      string     `json:"status" gorm:"default:'active'"` // active, completed, cancelled
	Notes       string     `json:"notes"`
}

// IPPort represents an IP address and its open port mapping
type IPPort struct {
	ID            uint      `json:"id" gorm:"primarykey"`
	IPAddress     string    `json:"ip_address" gorm:"index;not null"`
	Port          int       `json:"port" gorm:"index;not null"`
	Protocol      string    `json:"protocol" gorm:"default:'tcp'"` // tcp, udp
	Service       string    `json:"service"`                       // e.g., "ssh", "http", "https"
	State         string    `json:"state" gorm:"default:'open'"`   // open, closed, filtered
	Banner        string    `json:"banner"`                        // service banner if available
	ScanSessionID *uint     `json:"scan_session_id,omitempty" gorm:"index"`
	DiscoveredAt  time.Time `json:"discovered_at" gorm:"autoCreateTime"`

	// CDN Detection Information
	IsCDN        bool   `json:"is_cdn" gorm:"default:false"`       // Whether this IP/host is detected as CDN
	CDNName      string `json:"cdn_name"`                          // Name of CDN provider if detected
	CDNDetected  bool   `json:"cdn_detected" gorm:"default:false"` // Whether CDN detection was performed
	OriginalHost string `json:"original_host"`                     // Original hostname that resolved to this IP

	// Unique constraint on IP+Port combination within a scan session
	// This prevents duplicate entries for the same IP:port
}

// IPInfo represents comprehensive IP address information from Shodan
type IPInfo struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	IPAddress    string    `json:"ip_address" gorm:"uniqueIndex;not null"`
	Organization string    `json:"organization"`
	ISP          string    `json:"isp"`
	ASN          string    `json:"asn"`
	Country      string    `json:"country"`
	CountryCode  string    `json:"country_code"`
	City         string    `json:"city"`
	Region       string    `json:"region"`
	Postal       string    `json:"postal"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	OS           string    `json:"os"`
	Tags         string    `json:"tags"`      // JSON string array
	Ports        string    `json:"ports"`     // JSON int array
	Hostnames    string    `json:"hostnames"` // JSON string array
	Domains      string    `json:"domains"`   // JSON string array
	Vulns        string    `json:"vulns"`     // JSON string array
	LastUpdate   time.Time `json:"last_update"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations to existing models
	ScanSessionID *uint `json:"scan_session_id,omitempty" gorm:"index"`
}

// SetTags sets the tags field from a string slice
func (ip *IPInfo) SetTags(tags []string) error {
	if tags == nil {
		ip.Tags = ""
		return nil
	}
	data, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	ip.Tags = string(data)
	return nil
}

// GetTags returns the tags as a string slice
func (ip *IPInfo) GetTags() ([]string, error) {
	if ip.Tags == "" {
		return []string{}, nil
	}
	var tags []string
	err := json.Unmarshal([]byte(ip.Tags), &tags)
	return tags, err
}

// SetPorts sets the ports field from an int slice
func (ip *IPInfo) SetPorts(ports []int) error {
	if ports == nil {
		ip.Ports = ""
		return nil
	}
	data, err := json.Marshal(ports)
	if err != nil {
		return err
	}
	ip.Ports = string(data)
	return nil
}

// GetPorts returns the ports as an int slice
func (ip *IPInfo) GetPorts() ([]int, error) {
	if ip.Ports == "" {
		return []int{}, nil
	}
	var ports []int
	err := json.Unmarshal([]byte(ip.Ports), &ports)
	return ports, err
}

// SetHostnames sets the hostnames field from a string slice
func (ip *IPInfo) SetHostnames(hostnames []string) error {
	if hostnames == nil {
		ip.Hostnames = ""
		return nil
	}
	data, err := json.Marshal(hostnames)
	if err != nil {
		return err
	}
	ip.Hostnames = string(data)
	return nil
}

// GetHostnames returns the hostnames as a string slice
func (ip *IPInfo) GetHostnames() ([]string, error) {
	if ip.Hostnames == "" {
		return []string{}, nil
	}
	var hostnames []string
	err := json.Unmarshal([]byte(ip.Hostnames), &hostnames)
	return hostnames, err
}

// SetDomains sets the domains field from a string slice
func (ip *IPInfo) SetDomains(domains []string) error {
	if domains == nil {
		ip.Domains = ""
		return nil
	}
	data, err := json.Marshal(domains)
	if err != nil {
		return err
	}
	ip.Domains = string(data)
	return nil
}

// GetDomains returns the domains as a string slice
func (ip *IPInfo) GetDomains() ([]string, error) {
	if ip.Domains == "" {
		return []string{}, nil
	}
	var domains []string
	err := json.Unmarshal([]byte(ip.Domains), &domains)
	return domains, err
}

// SetVulns sets the vulns field from a string slice
func (ip *IPInfo) SetVulns(vulns []string) error {
	if vulns == nil {
		ip.Vulns = ""
		return nil
	}
	data, err := json.Marshal(vulns)
	if err != nil {
		return err
	}
	ip.Vulns = string(data)
	return nil
}

// GetVulns returns the vulnerabilities as a string slice
func (ip *IPInfo) GetVulns() ([]string, error) {
	if ip.Vulns == "" {
		return []string{}, nil
	}
	var vulns []string
	err := json.Unmarshal([]byte(ip.Vulns), &vulns)
	return vulns, err
}
