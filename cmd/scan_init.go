package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/spf13/cobra"
)

var scanInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new scan session with organized target structure",
	Long: ascii.Markdown(`
# scan init

Initialize a new scan session with a dedicated target directory structure.
This creates:

- A target folder: targets/<target>/
- Target-specific database: targets/<target>/<target>.sqlite3  
- Screenshot directory: targets/<target>/screenshots/
- Scan session record with company information

The target name must contain only lowercase letters, numbers, and underscores
for folder organization, while the company name can be the full business name.

Example:
- Company: "Alm. Brand Forsikring A/S"  
- Target: "almbrand"`),
	Example: ascii.Markdown(`
- gowitness scan init --company "Alm. Brand Forsikring A/S" --target almbrand --domain almbrand.dk
- gowitness scan init -c "Acme Corporation Ltd" --target acme_corp -d acme.com`),
	RunE: scanInitCmdRunE,
}

var (
	scanInitCompanyName string
	scanInitTargetName  string
	scanInitMainDomain  string
	scanInitNotes       string
)

func scanInitCmdRunE(cmd *cobra.Command, args []string) error {
	if scanInitCompanyName == "" {
		return fmt.Errorf("company name is required (--company)")
	}
	if scanInitTargetName == "" {
		return fmt.Errorf("target name is required (--target)")
	}
	if scanInitMainDomain == "" {
		return fmt.Errorf("main domain is required (--domain)")
	}

	// Validate target name format (lowercase, numbers, underscore only)
	validTargetName := regexp.MustCompile(`^[a-z0-9_]+$`)
	if !validTargetName.MatchString(scanInitTargetName) {
		return fmt.Errorf("target name must contain only lowercase letters, numbers, and underscores (got: %s)", scanInitTargetName)
	}

	// Create target directory structure
	targetDir := filepath.Join("targets", scanInitTargetName)
	screenshotDir := filepath.Join(targetDir, "screenshots")
	dbPath := filepath.Join(targetDir, scanInitTargetName+".sqlite3")

	// Create directories
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory structure: %w", err)
	}

	log.Info("created target directory structure",
		"target-dir", targetDir,
		"screenshot-dir", screenshotDir,
		"database-path", dbPath)

	// Try to fetch company logo from Clearbit
	var logoPath string
	log.Info("attempting to fetch company logo from Clearbit", "domain", scanInitMainDomain)
	fetchedLogoPath, err := islazy.FetchClearbitLogo(scanInitMainDomain, targetDir)
	if err != nil {
		log.Warn("failed to fetch logo from Clearbit - you may need to add one manually",
			"domain", scanInitMainDomain,
			"error", err.Error(),
			"location", filepath.Join(targetDir, "logo.png"))
	} else {
		logoPath = fetchedLogoPath
		log.Info("successfully fetched company logo", "path", logoPath)
	}

	// Connect to target-specific database
	dbURI := fmt.Sprintf("sqlite://%s", dbPath)
	conn, err := database.Connection(dbURI, false, opts.Writer.DbDebug)
	if err != nil {
		return fmt.Errorf("failed to connect to target database: %w", err)
	}

	// Create new scan session
	session := &models.ScanSession{
		CompanyName: scanInitCompanyName,
		MainDomain:  scanInitMainDomain,
		LogoPath:    logoPath,
		StartTime:   time.Now(),
		Status:      "active",
		Notes:       scanInitNotes,
	}

	if err := conn.Create(session).Error; err != nil {
		return fmt.Errorf("failed to create scan session: %w", err)
	}

	log.Info("scan session initialized",
		"session-id", session.ID,
		"company", session.CompanyName,
		"target", scanInitTargetName,
		"domain", session.MainDomain,
		"database", dbPath,
		"screenshots", screenshotDir,
		"start-time", session.StartTime.Format(time.RFC3339))

	log.Info("use these settings for subsequent scans:",
		"db-uri", dbURI,
		"screenshot-path", screenshotDir)

	return nil
}

func init() {
	scanCmd.AddCommand(scanInitCmd)

	scanInitCmd.Flags().StringVarP(&scanInitCompanyName, "company", "c", "", "Full company name (required)")
	scanInitCmd.Flags().StringVar(&scanInitTargetName, "target", "", "Target folder name - lowercase, numbers, underscore only (required)")
	scanInitCmd.Flags().StringVarP(&scanInitMainDomain, "domain", "d", "", "Target company main domain (required)")
	scanInitCmd.Flags().StringVarP(&scanInitNotes, "notes", "n", "", "Optional notes about the scan session")

	// Mark required flags
	scanInitCmd.MarkFlagRequired("company")
	scanInitCmd.MarkFlagRequired("target")
	scanInitCmd.MarkFlagRequired("domain")
}
