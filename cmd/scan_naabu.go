package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var naabuCmdOptions = struct {
	File          string
	TopPorts      string
	CustomPorts   string
	Rate          int
	Threads       int
	Timeout       int
	ExcludeCDN    bool
	DisplayCDN    bool
	Verbose       bool
	ScanSessionID uint
	OutputFile    string
}{}

// NaabuResult represents a single port scan result from naabu JSON output
type NaabuResult struct {
	Host     string `json:"host"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	CDN      bool   `json:"cdn"`
	CDNName  string `json:"cdn-name"`
	Protocol string `json:"protocol"`
}

var naabuCmd = &cobra.Command{
	Use:   "naabu",
	Short: "Run naabu port scanner against a list of domains",
	Long: ascii.LogoHelp(ascii.Markdown(`
# scan naabu

Run naabu port scanner against a list of domains and store the results in the 
IPPort table. This command does NOT perform web screenshots - it only does 
port scanning and populates the port information in the database.

The command automatically excludes CDN/WAF services from full port scans to 
avoid scanning CDN infrastructure (only scans ports 80,443 for CDN hosts).

**Note**: This command requires naabu to be installed. Run 'make prerequisites' 
to install naabu and its dependencies.`)),
	Example: ascii.Markdown(`
- gowitness scan naabu -f domains.txt --write-db
- gowitness scan naabu -f targets.txt --top-ports 1000 --write-db --scan-session-id 1
- gowitness scan naabu -f hosts.txt --custom-ports "22,80,443,8080" --rate 500 --write-db
- gowitness scan naabu -f domains.txt --exclude-cdn --display-cdn --verbose --write-db`),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if naabuCmdOptions.File == "" {
			return errors.New("a file with domains must be specified")
		}

		// Check if file exists
		if _, err := os.Stat(naabuCmdOptions.File); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", naabuCmdOptions.File)
		}

		// Check if naabu is installed
		if _, err := exec.LookPath("naabu"); err != nil {
			return errors.New("naabu is not installed. Please run 'make prerequisites' to install it")
		}

		// Check if database output is specified
		if !opts.Writer.Db {
			return errors.New("--write-db flag is required for naabu scans")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("starting naabu port scan",
			"file", naabuCmdOptions.File,
			"exclude-cdn", naabuCmdOptions.ExcludeCDN,
			"scan-session-id", naabuCmdOptions.ScanSessionID)

		// Create temporary output file for naabu results
		tempFile := naabuCmdOptions.OutputFile
		if tempFile == "" {
			tempFile = fmt.Sprintf("naabu_results_%d.json", time.Now().Unix())
		}
		defer func() {
			if naabuCmdOptions.OutputFile == "" {
				os.Remove(tempFile) // Clean up temp file if we created it
			}
		}()

		// Build naabu command
		naabuArgs := buildNaabuCommand(tempFile)

		// Execute naabu
		if err := executeNaabu(naabuArgs); err != nil {
			log.Error("failed to execute naabu", "err", err)
			return
		}

		// Parse results and save to database
		if err := parseAndSaveResults(tempFile); err != nil {
			log.Error("failed to parse and save naabu results", "err", err)
			return
		}

		log.Info("naabu port scan completed successfully")
	},
}

func buildNaabuCommand(outputFile string) []string {
	args := []string{
		"-l", naabuCmdOptions.File,
		"-json",
		"-o", outputFile,
		"-display-cdn", // Always enable CDN detection for database storage
	}

	// Always exclude CDN by default for safety
	if naabuCmdOptions.ExcludeCDN {
		args = append(args, "-exclude-cdn")
	}

	if naabuCmdOptions.Verbose {
		args = append(args, "-verbose")
	}

	// Port selection
	if naabuCmdOptions.CustomPorts != "" {
		args = append(args, "-p", naabuCmdOptions.CustomPorts)
	} else if naabuCmdOptions.TopPorts != "" {
		args = append(args, "-top-ports", naabuCmdOptions.TopPorts)
	}

	// Performance settings
	if naabuCmdOptions.Rate > 0 {
		args = append(args, "-rate", fmt.Sprintf("%d", naabuCmdOptions.Rate))
	}

	if naabuCmdOptions.Threads > 0 {
		args = append(args, "-c", fmt.Sprintf("%d", naabuCmdOptions.Threads))
	}

	if naabuCmdOptions.Timeout > 0 {
		args = append(args, "-timeout", fmt.Sprintf("%d", naabuCmdOptions.Timeout))
	}

	return args
}

func executeNaabu(args []string) error {
	log.Info("executing naabu", "args", strings.Join(args, " "))

	cmd := exec.Command("naabu", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func parseAndSaveResults(filename string) error {
	// Connect to database
	db, err := database.Connection(opts.Writer.DbURI, false, opts.Writer.DbDebug)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Read naabu results file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read results file: %w", err)
	}

	// Parse JSON lines
	lines := strings.Split(string(data), "\n")
	var savedCount int
	var skippedCount int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var result NaabuResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			log.Warn("failed to parse naabu result line", "line", line, "err", err)
			skippedCount++
			continue
		}

		// Create IPPort entry
		ipPort := models.IPPort{
			IPAddress:     result.IP,
			Port:          result.Port,
			Protocol:      result.Protocol, // Use protocol from naabu result
			State:         "open",
			ScanSessionID: getValidScanSessionID(),
			IsCDN:         result.CDN,
			CDNName:       result.CDNName,
			CDNDetected:   true, // We always run CDN detection
			OriginalHost:  result.Host,
		}

		// Check if this IP:Port combination already exists
		var existing models.IPPort
		if err := db.Where("ip_address = ? AND port = ?", result.IP, result.Port).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Not found, create new record
				if err := db.Create(&ipPort).Error; err != nil {
					log.Warn("failed to save port result", "ip", result.IP, "port", result.Port, "err", err)
					skippedCount++
					continue
				}
				savedCount++
			} else {
				log.Warn("database error checking for existing port", "ip", result.IP, "port", result.Port, "err", err)
				skippedCount++
				continue
			}
		} else {
			// Record already exists, skip
			skippedCount++
		}
	}

	log.Info("naabu results processed", "saved", savedCount, "skipped", skippedCount)
	return nil
}

func getValidScanSessionID() *uint {
	if naabuCmdOptions.ScanSessionID > 0 {
		return &naabuCmdOptions.ScanSessionID
	}
	return nil
}

func init() {
	scanCmd.AddCommand(naabuCmd)

	naabuCmd.Flags().StringVarP(&naabuCmdOptions.File, "file", "f", "", "File containing list of domains/hosts to scan (required)")
	naabuCmd.Flags().StringVar(&naabuCmdOptions.TopPorts, "top-ports", "100", "Top ports to scan [100,1000,full]")
	naabuCmd.Flags().StringVar(&naabuCmdOptions.CustomPorts, "custom-ports", "", "Custom ports to scan (e.g., '22,80,443,8080')")
	naabuCmd.Flags().IntVar(&naabuCmdOptions.Rate, "rate", 500, "Packets to send per second")
	naabuCmd.Flags().IntVar(&naabuCmdOptions.Threads, "threads", 25, "Number of concurrent threads")
	naabuCmd.Flags().IntVar(&naabuCmdOptions.Timeout, "timeout", 1000, "Timeout in milliseconds")
	naabuCmd.Flags().BoolVar(&naabuCmdOptions.ExcludeCDN, "exclude-cdn", true, "Skip full port scans for CDN/WAF (only scan 80,443)")
	naabuCmd.Flags().BoolVar(&naabuCmdOptions.DisplayCDN, "display-cdn", false, "Display CDN detection information")
	naabuCmd.Flags().BoolVar(&naabuCmdOptions.Verbose, "verbose", false, "Enable verbose output")
	naabuCmd.Flags().UintVar(&naabuCmdOptions.ScanSessionID, "scan-session-id", 0, "Associate results with specific scan session ID")
	naabuCmd.Flags().StringVar(&naabuCmdOptions.OutputFile, "output", "", "File to save naabu JSON results (optional, uses temp file by default)")
}
