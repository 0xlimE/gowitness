package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/sensepost/gowitness/pkg/shodan"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var shodanCmdOptions = struct {
	File          string
	Verbose       bool
	ScanSessionID uint
	RateLimit     int    // Rate limit for API calls (per minute)
	ProjectName   string // Project name for status updates
}{}

var shodanCmd = &cobra.Command{
	Use:   "shodan",
	Short: "Query Shodan API for IP information with IP-API/naabu fallback",
	Long: ascii.LogoHelp(ascii.Markdown(`
# scan shodan

Query Shodan API for comprehensive IP address information with automatic 
fallback to IP-API and naabu port scanning when Shodan data is unavailable.

This command takes a list of domains/IPs, resolves them to IP addresses, and:

1. **First tries Shodan API** for detailed information including:
   - Open ports and services
   - Organization and ISP information  
   - Geographic location
   - Operating system detection
   - Vulnerability information
   - Hostnames and domains
   - ASN information

2. **Falls back to IP-API + naabu** when Shodan fails or has no data:
   - IP-API.com for geolocation and ISP information
   - naabu port scanner for open port detection
   - Ensures data is always populated

This guarantees that IP intelligence is gathered regardless of Shodan API 
availability. Shodan requires an API key (SHODAN_API_KEY environment variable), 
but the command will work without it using fallback methods.

**Note**: Shodan queries consume 1 API credit each. Fallback methods are free.`)),
	Example: ascii.Markdown(`
- gowitness scan shodan -f domains.txt --write-db
- gowitness scan shodan -f targets.txt --write-db --scan-session-id 1  
- gowitness scan shodan -f hosts.txt --rate-limit 30 --verbose --write-db
- gowitness scan shodan -f ips.txt --write-db  # Works without Shodan API key`),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if shodanCmdOptions.File == "" {
			return errors.New("a file with domains/IPs must be specified")
		}

		// Check if file exists
		if _, err := os.Stat(shodanCmdOptions.File); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", shodanCmdOptions.File)
		}

		// Check if database output is specified
		if !opts.Writer.Db {
			return errors.New("--write-db flag is required for shodan scans")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("starting Shodan IP information gathering",
			"file", shodanCmdOptions.File,
			"scan-session-id", shodanCmdOptions.ScanSessionID,
			"rate-limit", shodanCmdOptions.RateLimit)

		// Update project status to running
		updateProjectStatus(shodanCmdOptions.ProjectName, "Running - (Portscanning)")

		if err := runShodanScan(); err != nil {
			log.Error("failed to complete Shodan scan", "err", err)
			// Update status to error
			updateProjectStatus(shodanCmdOptions.ProjectName, "Error - (Portscanning failed)")
			return
		}

		// Update status to complete
		updateProjectStatus(shodanCmdOptions.ProjectName, "Complete - (Portscanning)")
		log.Info("Shodan IP information gathering completed successfully")
	},
}

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

// shodanNaabuResult represents naabu port scan result for shodan command
type shodanNaabuResult struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// fetchIPAPIData fetches geolocation data from ip-api.com as fallback
func fetchIPAPIData(ip string) (*IPAPIResponse, error) {
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
func runNaabuScan(ip string) ([]int, error) {
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

		var result shodanNaabuResult
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

// createFallbackIPInfo creates IP info from fallback sources
func createFallbackIPInfo(db *gorm.DB, ip string) (*models.IPInfo, error) {
	log.Info("attempting fallback IP intelligence gathering", "ip", ip)

	// Try IP-API for geolocation
	ipApiData, err := fetchIPAPIData(ip)
	if err != nil {
		log.Warn("failed to fetch IP-API data", "ip", ip, "err", err)
		return nil, fmt.Errorf("fallback IP-API failed: %w", err)
	}

	// Try naabu for port scanning
	ports, err := runNaabuScan(ip)
	if err != nil {
		log.Warn("failed to run naabu scan", "ip", ip, "err", err)
		// Continue without port data - IP-API data is still valuable
		ports = []int{}
	} else {
		log.Info("naabu scan completed", "ip", ip, "ports_found", len(ports))
	}

	// Create IPInfo from fallback data
	ipInfo := &models.IPInfo{
		IPAddress:     ip,
		Organization:  ipApiData.Org,
		ISP:           ipApiData.ISP,
		ASN:           ipApiData.AS,
		Country:       ipApiData.Country,
		CountryCode:   ipApiData.CountryCode,
		City:          ipApiData.City,
		Region:        ipApiData.RegionName,
		Postal:        ipApiData.Zip,
		Latitude:      ipApiData.Lat,
		Longitude:     ipApiData.Lon,
		LastUpdate:    time.Now(),
		ScanSessionID: getValidShodanScanSessionID(),
	}

	// Set ports from naabu scan
	if len(ports) > 0 {
		if err := ipInfo.SetPorts(ports); err != nil {
			log.Warn("failed to set ports for IP info", "ip", ip, "err", err)
		}

		// Also create IPPort entries for consistency with Shodan data
		if err := createFallbackIPPortEntries(db, ip, ports); err != nil {
			log.Warn("failed to create IPPort entries for fallback", "ip", ip, "err", err)
		}
	}

	log.Info("created fallback IP info", "ip", ip, "source", "ip-api+naabu", "org", ipInfo.Organization)
	return ipInfo, nil
}

// createFallbackIPPortEntries creates IPPort entries for fallback scan results
func createFallbackIPPortEntries(db *gorm.DB, ip string, ports []int) error {
	sessionID := getValidShodanScanSessionID()

	for _, port := range ports {
		// Check if this IP:Port combination already exists
		var existing models.IPPort
		if err := db.Where("ip_address = ? AND port = ?", ip, port).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new IPPort entry
				ipPort := models.IPPort{
					IPAddress:     ip,
					Port:          port,
					Protocol:      "tcp", // naabu typically scans TCP ports
					State:         "open",
					Service:       "", // No service detection in fallback
					ScanSessionID: sessionID,
					IsCDN:         false,
					CDNDetected:   false,
				}

				if err := db.Create(&ipPort).Error; err != nil {
					log.Warn("failed to create fallback IPPort entry", "ip", ip, "port", port, "err", err)
				}
			}
		}
	}

	return nil
}

func runShodanScan() error {
	// Try to initialize Shodan client - it's OK if this fails, we'll use fallback
	client, err := shodan.InitFromEnv()
	if err != nil {
		log.Warn("failed to initialize Shodan client, will use fallback methods", "err", err)
		client = nil // Explicitly set to nil for clarity
	} else {
		log.Info("Shodan client initialized successfully")
	}

	// Connect to database
	db, err := database.Connection(opts.Writer.DbURI, false, opts.Writer.DbDebug)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Read hosts from file
	hosts, err := readHostsFromFile(shodanCmdOptions.File)
	if err != nil {
		return fmt.Errorf("failed to read hosts from file: %w", err)
	}

	// Resolve domains to IPs and deduplicate
	ips, err := resolveAndDeduplicateIPs(hosts)
	if err != nil {
		return fmt.Errorf("failed to resolve IPs: %w", err)
	}

	log.Info("resolved unique IP addresses", "count", len(ips))

	// Process each IP with rate limiting
	var processedCount, savedCount, skippedCount, errorCount, fallbackCount int
	rateLimiter := time.NewTicker(time.Minute / time.Duration(shodanCmdOptions.RateLimit))
	defer rateLimiter.Stop()

	for _, ip := range ips {
		// Rate limiting
		if processedCount > 0 {
			<-rateLimiter.C
		}
		processedCount++

		if shodanCmdOptions.Verbose {
			log.Info("querying Shodan for IP", "ip", ip, "progress", fmt.Sprintf("%d/%d", processedCount, len(ips)))
		}

		// Check if we already have this IP in the database
		var existing models.IPInfo
		if err := db.Where("ip_address = ?", ip).First(&existing).Error; err == nil {
			// IP already exists, skip
			skippedCount++
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("database error checking existing IP", "ip", ip, "err", err)
			errorCount++
			continue
		}

		var ipInfo *models.IPInfo
		var usedFallback bool

		// Try Shodan first if client is available
		if client != nil {
			host, err := client.GetHostMinimal(ip)
			if err != nil {
				log.Warn("failed to query Shodan for IP", "ip", ip, "err", err)
				// ipInfo remains nil, will trigger fallback
			} else {
				// Successfully got Shodan data
				ipInfo = &models.IPInfo{
					IPAddress:     host.IP,
					Organization:  host.Organization,
					ISP:           host.ISP,
					ASN:           host.ASN,
					Country:       host.Country,
					CountryCode:   host.CountryCode,
					City:          host.City,
					Region:        host.Region,
					Postal:        host.Postal,
					Latitude:      host.Latitude,
					Longitude:     host.Longitude,
					OS:            host.OS,
					LastUpdate:    host.LastUpdate.Time,
					ScanSessionID: getValidShodanScanSessionID(),
				}

				// Set array fields using helper methods
				if err := ipInfo.SetTags(host.Tags); err != nil {
					log.Warn("failed to set tags for IP", "ip", ip, "err", err)
				}
				if err := ipInfo.SetPorts(host.Ports); err != nil {
					log.Warn("failed to set ports for IP", "ip", ip, "err", err)
				}
				if err := ipInfo.SetHostnames(host.Hostnames); err != nil {
					log.Warn("failed to set hostnames for IP", "ip", ip, "err", err)
				}
				if err := ipInfo.SetDomains(host.Domains); err != nil {
					log.Warn("failed to set domains for IP", "ip", ip, "err", err)
				}
				if err := ipInfo.SetVulns(host.Vulns); err != nil {
					log.Warn("failed to set vulnerabilities for IP", "ip", ip, "err", err)
				}

				// Also create IPPort entries for open ports
				if err := createIPPortEntries(db, host); err != nil {
					log.Warn("failed to create IPPort entries", "ip", ip, "err", err)
				}
			}
		}

		// If Shodan failed or no client available, try fallback
		if ipInfo == nil {
			if fallbackInfo, err := createFallbackIPInfo(db, ip); err != nil {
				log.Error("both Shodan and fallback failed for IP", "ip", ip, "err", err)
				errorCount++
				continue
			} else {
				ipInfo = fallbackInfo
				usedFallback = true
				fallbackCount++
			}
		}

		// Save to database
		if err := db.Create(ipInfo).Error; err != nil {
			log.Warn("failed to save IP info to database", "ip", ip, "err", err)
			errorCount++
			continue
		}

		savedCount++

		if shodanCmdOptions.Verbose {
			source := "shodan"
			if usedFallback {
				source = "ip-api+naabu"
			}
			log.Info("saved IP information", "ip", ip, "organization", ipInfo.Organization, "source", source)
		}
	}

	log.Info("Shodan scan results",
		"processed", processedCount,
		"saved", savedCount,
		"skipped", skippedCount,
		"errors", errorCount,
		"fallback_used", fallbackCount)

	return nil
}

func readHostsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			hosts = append(hosts, line)
		}
	}

	return hosts, scanner.Err()
}

func resolveAndDeduplicateIPs(hosts []string) ([]string, error) {
	ipSet := make(map[string]bool)

	for _, host := range hosts {
		// Check if it's already an IP address
		if ip := net.ParseIP(host); ip != nil {
			ipSet[host] = true
			continue
		}

		// Remove protocol and port if present
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
		if colonIndex := strings.LastIndex(host, ":"); colonIndex > 0 {
			// Only remove port if it's not an IPv6 address
			if !strings.Contains(host, "]") {
				host = host[:colonIndex]
			}
		}

		// Resolve domain to IP addresses
		ips, err := net.LookupIP(host)
		if err != nil {
			log.Warn("failed to resolve host", "host", host, "err", err)
			continue
		}

		for _, ip := range ips {
			// Only include IPv4 addresses
			if ipv4 := ip.To4(); ipv4 != nil {
				ipSet[ip.String()] = true
			}
		}
	}

	// Convert set to slice
	var result []string
	for ip := range ipSet {
		result = append(result, ip)
	}

	return result, nil
}

func createIPPortEntries(db *gorm.DB, host *shodan.Host) error {
	sessionID := getValidShodanScanSessionID()

	for _, port := range host.Ports {
		// Check if this IP:Port combination already exists
		var existing models.IPPort
		if err := db.Where("ip_address = ? AND port = ?", host.IP, port).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new IPPort entry
				ipPort := models.IPPort{
					IPAddress:     host.IP,
					Port:          port,
					Protocol:      "tcp", // Shodan typically reports TCP ports
					State:         "open",
					Service:       "", // Could be enhanced with service detection from Shodan data
					ScanSessionID: sessionID,
					IsCDN:         false, // Could be enhanced with CDN detection
					CDNDetected:   false,
				}

				if err := db.Create(&ipPort).Error; err != nil {
					log.Warn("failed to create IPPort entry", "ip", host.IP, "port", port, "err", err)
				}
			}
		}
	}

	return nil
}

func getValidShodanScanSessionID() *uint {
	if shodanCmdOptions.ScanSessionID > 0 {
		return &shodanCmdOptions.ScanSessionID
	}
	return nil
}

// updateProjectStatus logs project status (admin panel removed)
func updateProjectStatus(projectName, status string) {
	if projectName == "" {
		return
	}
	log.Debug("project status update", "project", projectName, "status", status)
}

func init() {
	scanCmd.AddCommand(shodanCmd)

	shodanCmd.Flags().StringVarP(&shodanCmdOptions.File, "file", "f", "", "File containing list of domains/IPs to query (required)")
	shodanCmd.Flags().BoolVar(&shodanCmdOptions.Verbose, "verbose", false, "Enable verbose output")
	shodanCmd.Flags().UintVar(&shodanCmdOptions.ScanSessionID, "scan-session-id", 0, "Associate results with specific scan session ID")
	shodanCmd.Flags().IntVar(&shodanCmdOptions.RateLimit, "rate-limit", 60, "API calls per minute (default: 60)")
	shodanCmd.Flags().StringVar(&shodanCmdOptions.ProjectName, "project", "", "Project name for status updates (optional)")
}
