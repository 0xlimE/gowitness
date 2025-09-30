package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/spf13/cobra"
)

var runCmdOptions = struct {
	ProjectPath string
	Verbose     bool
	ProjectName string // Project name for status updates
	SkipShodan  bool   // Skip Shodan scan
	SkipScreens bool   // Skip screenshot collection
}{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a complete scan workflow for a project",
	Long: ascii.LogoHelp(ascii.Markdown(`
# scan run

Execute a complete scan workflow for a project directory containing a domains.txt file.
This command orchestrates multiple gowitness commands in sequence:

1. **Shodan Intelligence Gathering**: Query Shodan API for IP information with fallback
2. **Screenshot Collection**: Capture website screenshots for all discovered domains
3. **Database Updates**: Update project status and completion tracking

The command expects a project directory structure like:
- targets/project_name/
  - domains.txt (list of domains to scan)
  - project_name.sqlite3 (database file)
  - screenshots/ (screenshot output directory)

Status updates are logged to the console for monitoring.
`)),
	Example: ascii.Markdown(`
- gowitness scan run -p targets/company_name/
- gowitness scan run -p targets/demo_project/ --project demo_project --verbose
- gowitness scan run -p targets/example/ --skip-shodan  # Screenshots only
- gowitness scan run -p targets/test/ --skip-screens    # Shodan only`),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if runCmdOptions.ProjectPath == "" {
			return errors.New("project path must be specified with -p/--path")
		}

		// Check if project directory exists
		if _, err := os.Stat(runCmdOptions.ProjectPath); os.IsNotExist(err) {
			return fmt.Errorf("project directory does not exist: %s", runCmdOptions.ProjectPath)
		}

		// Check if domains.txt exists
		domainsFile := filepath.Join(runCmdOptions.ProjectPath, "domains.txt")
		if _, err := os.Stat(domainsFile); os.IsNotExist(err) {
			return fmt.Errorf("domains.txt file not found in project directory: %s", domainsFile)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("starting complete scan workflow",
			"project_path", runCmdOptions.ProjectPath,
			"project_name", runCmdOptions.ProjectName,
			"skip_shodan", runCmdOptions.SkipShodan,
			"skip_screens", runCmdOptions.SkipScreens)

		// Update project status to running
		updateRunProjectStatus(runCmdOptions.ProjectName, "Running - (Full Scan)")

		// Execute the scan workflow
		err := executeFullScanWorkflow(runCmdOptions.ProjectPath, runCmdOptions.ProjectName)
		if err != nil {
			log.Error("scan workflow failed", "error", err)
			updateRunProjectStatus(runCmdOptions.ProjectName, "Error - (Scan failed)")
			return
		}

		// Update project status to complete
		updateRunProjectStatus(runCmdOptions.ProjectName, "Complete - (Full Scan)")

		log.Info("scan workflow completed successfully",
			"project_path", runCmdOptions.ProjectPath,
			"project_name", runCmdOptions.ProjectName)
	},
}

// ScanPhase represents a phase in the scan workflow
type ScanPhase struct {
	Name       string
	StatusName string
	Command    func(projectPath, projectName string) error
	Skip       bool
}

// executeFullScanWorkflow runs the complete scan workflow
func executeFullScanWorkflow(projectPath, projectName string) error {
	log.Info("executing full scan workflow", "project", projectName, "path", projectPath)

	// Define scan phases
	phases := []ScanPhase{
		{
			Name:       "Shodan Intelligence",
			StatusName: "Portscanning",
			Command:    executeShodanScan,
			Skip:       runCmdOptions.SkipShodan,
		},
		{
			Name:       "Screenshot Collection",
			StatusName: "Screenshotting",
			Command:    executeScreenshotScan,
			Skip:       runCmdOptions.SkipScreens,
		},
	}

	// Execute each phase
	for _, phase := range phases {
		if phase.Skip {
			log.Info("skipping scan phase", "phase", phase.Name)
			continue
		}

		log.Info("starting scan phase", "phase", phase.Name)
		updateRunProjectStatus(projectName, fmt.Sprintf("Running - (%s)", phase.StatusName))

		err := phase.Command(projectPath, projectName)
		if err != nil {
			log.Error("scan phase failed", "phase", phase.Name, "error", err)
			updateRunProjectStatus(projectName, fmt.Sprintf("Error - (%s failed)", phase.StatusName))
			return fmt.Errorf("scan phase '%s' failed: %w", phase.Name, err)
		}

		log.Info("scan phase completed", "phase", phase.Name)
		updateRunProjectStatus(projectName, fmt.Sprintf("Complete - (%s)", phase.StatusName))

		// Small delay between phases
		time.Sleep(1 * time.Second)
	}

	return nil
}

// executeShodanScan runs the Shodan intelligence gathering phase
func executeShodanScan(projectPath, projectName string) error {
	log.Info("executing Shodan scan", "project", projectName)

	domainsFile := filepath.Join(projectPath, "domains.txt")
	projectDirName := filepath.Base(projectPath)
	dbFile := filepath.Join(projectPath, fmt.Sprintf("%s.sqlite3", projectDirName))

	// Build command arguments
	args := []string{"scan", "shodan", "-f", domainsFile, "--write-db", "--write-db-uri", fmt.Sprintf("sqlite://%s", dbFile)}

	if runCmdOptions.Verbose {
		args = append(args, "--verbose")
	}

	if projectName != "" {
		args = append(args, "--project", projectName)
	}

	// Execute command
	cmd := exec.Command("./gowitness", args...)
	cmd.Dir = "." // Run from current directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Shodan scan command failed", "error", err, "output", string(output))
		return fmt.Errorf("shodan scan failed: %s", string(output))
	}

	log.Info("Shodan scan completed successfully", "project", projectName)
	return nil
}

// executeScreenshotScan runs the screenshot collection phase
func executeScreenshotScan(projectPath, projectName string) error {
	log.Info("executing screenshot scan", "project", projectName)

	domainsFile := filepath.Join(projectPath, "domains.txt")
	projectDirName := filepath.Base(projectPath)
	dbFile := filepath.Join(projectPath, fmt.Sprintf("%s.sqlite3", projectDirName))
	screenshotDir := filepath.Join(projectPath, "screenshots")

	// Ensure screenshot directory exists
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshot directory: %w", err)
	}

	// Build command arguments
	args := []string{"scan", "file", "-f", domainsFile, "--write-db", "--write-db-uri", fmt.Sprintf("sqlite://%s", dbFile), "--screenshot-path", screenshotDir}

	if runCmdOptions.Verbose {
		args = append(args, "--debug-log")
	}

	if projectName != "" {
		args = append(args, "--project", projectName)
	}

	// Execute command
	cmd := exec.Command("./gowitness", args...)
	cmd.Dir = "." // Run from current directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("screenshot scan command failed", "error", err, "output", string(output))
		return fmt.Errorf("screenshot scan failed: %s", string(output))
	}

	log.Info("screenshot scan completed successfully", "project", projectName)
	return nil
}

// updateRunProjectStatus updates the project status via admin API
func updateRunProjectStatus(projectName, status string) {
	if projectName == "" {
		return
	}

	// Use the same implementation as other scan commands
	updateProjectStatus(projectName, status)
}

func init() {
	scanCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runCmdOptions.ProjectPath, "path", "p", "", "Path to the project directory")
	runCmd.Flags().BoolVarP(&runCmdOptions.Verbose, "verbose", "v", false, "Enable verbose output")
	runCmd.Flags().StringVar(&runCmdOptions.ProjectName, "project", "", "Project name for status tracking")
	runCmd.Flags().BoolVar(&runCmdOptions.SkipShodan, "skip-shodan", false, "Skip Shodan intelligence gathering phase")
	runCmd.Flags().BoolVar(&runCmdOptions.SkipScreens, "skip-screens", false, "Skip screenshot collection phase")
}
