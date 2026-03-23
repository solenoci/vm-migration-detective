package vmdetect

import (
	"context"
	"fmt"
	"time"

	"github.com/kubev2v/vm-migration-detective/internal/checks"
	"github.com/kubev2v/vm-migration-detective/internal/persistent"
	"github.com/kubev2v/vm-migration-detective/internal/vsphere"
	"github.com/kubev2v/vm-migration-detective/pkg/types"
	"github.com/sirupsen/logrus"
)

// CheckRunner orchestrates validation checks on VMs
type CheckRunner struct {
	inspector   persistent.InspectorInterface
	credentials Credentials
	logger      *logrus.Logger
}

// CheckRunnerConfig contains configuration for creating a CheckRunner
type CheckRunnerConfig struct {
	// Credentials for vCenter access (required)
	Credentials Credentials
	// VDDKLibDir is the path to VDDK library directory (required, cannot be empty)
	VDDKLibDir string

	// VirtInspectorPath is the path to virt-inspector executable (optional, uses system PATH if nil)
	VirtInspectorPath *string
	// VirtV2vInspectorPath is the path to virt-v2v-inspector executable (optional, uses system PATH if nil)
	VirtV2vInspectorPath *string
	// Timeout for inspection operations (optional, defaults to 30 minutes if nil)
	Timeout *time.Duration
	// Logger for logging (optional, can be nil)
	Logger *logrus.Logger
	// DB for persistent caching (optional, can be nil for memory-only caching)
	DB DB
}

// NewCheckRunner creates a new CheckRunner with an internally managed inspector instance
// Returns an error if required configuration is missing or invalid
func NewCheckRunner(config CheckRunnerConfig) (*CheckRunner, error) {
	// Validate required credentials
	if config.Credentials.VCenterURL == "" {
		return nil, fmt.Errorf("credentials.VCenterURL is required")
	}
	if config.Credentials.Username == "" {
		return nil, fmt.Errorf("credentials.Username is required")
	}
	if config.Credentials.Password == "" {
		return nil, fmt.Errorf("credentials.Password is required")
	}

	// Validate required VDDKLibDir
	if config.VDDKLibDir == "" {
		return nil, fmt.Errorf("VDDKLibDir is required and cannot be empty")
	}

	// Extract optional string parameters
	virtInspectorPath := ""
	if config.VirtInspectorPath != nil {
		virtInspectorPath = *config.VirtInspectorPath
	}

	virtV2vInspectorPath := ""
	if config.VirtV2vInspectorPath != nil {
		virtV2vInspectorPath = *config.VirtV2vInspectorPath
	}

	// Set default timeout if not provided
	timeout := 30 * time.Minute
	if config.Timeout != nil {
		timeout = *config.Timeout
	}

	// Create the inspector internally
	inspector := persistent.NewInspector(
		virtInspectorPath,
		virtV2vInspectorPath,
		timeout,
		config.Credentials,
		config.Logger,
		config.DB,
		config.VDDKLibDir,
	)

	return &CheckRunner{
		inspector:   inspector,
		credentials: config.Credentials,
		logger:      config.Logger,
	}, nil
}

// RunChecksParams contains parameters for running checks
type RunChecksParams struct {
	Ctx           context.Context
	VMMoref       string
	SnapshotMoref string
}

// RunChecksResult contains the results of all checks
type RunChecksResult struct {
	// Results contains individual check results
	Results []CheckResult `json:"results"`
	// AllConcerns aggregates all concerns from all checks
	AllConcerns []Concern `json:"all_concerns"`
	// Passed indicates if all checks passed (no concerns found)
	Passed bool `json:"passed"`
}

// RunChecks executes validation checks on a VM snapshot
// If checkTypes is empty, all checks are run. Otherwise, only specified checks are executed.
func (r *CheckRunner) RunChecks(params RunChecksParams, checkTypes ...CheckType) (*RunChecksResult, error) {
	// Validate required parameters
	if params.Ctx == nil {
		return nil, fmt.Errorf("params.Ctx is required")
	}
	if params.VMMoref == "" {
		return nil, fmt.Errorf("params.VMMoref is required")
	}
	if params.SnapshotMoref == "" {
		return nil, fmt.Errorf("params.SnapshotMoref is required")
	}

	// Get snapshot disk info from vSphere
	diskInfo, err := r.getSnapshotDiskInfo(params.Ctx, params.VMMoref, params.SnapshotMoref)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot disk info: %w", err)
	}

	// Determine which checks to run
	checksToRun := checkTypes
	if len(checksToRun) == 0 {
		// Run all checks by default
		checksToRun = AllCheckTypes()
	}

	// Create inspection params with the shared inspector
	inspectionParams := checks.InspectionParams{
		Ctx:           params.Ctx,
		VMMoref:       params.VMMoref,
		SnapshotMoref: params.SnapshotMoref,
		DiskInfo:      diskInfo,
		Inspector:     r.inspector,
	}

	results := make([]CheckResult, 0, len(checksToRun))
	allConcerns := []Concern{}
	allPassed := true

	for _, checkType := range checksToRun {
		var check checks.Check
		var result CheckResult

		switch checkType {
		case CheckTypeFstab:
			check = checks.NewFstabCheck()
		case CheckTypeDiskAccess:
			check = checks.NewDiskAccessCheck()
		default:
			// Unknown check type, skip
			continue
		}

		// Run the check
		checkResult := check.Run(inspectionParams)

		// Convert checks.CheckResult to vmdetect.CheckResult
		result = CheckResult{
			CheckType: checkType,
			Passed:    checkResult.Passed,
			Concerns:  checkResult.Concerns,
			Error:     checkResult.Error,
		}

		results = append(results, result)
		allConcerns = append(allConcerns, result.Concerns...)

		if !result.Passed {
			allPassed = false
		}
	}

	return &RunChecksResult{
		Results:     results,
		AllConcerns: allConcerns,
		Passed:      allPassed,
	}, nil
}

// getSnapshotDiskInfo queries vSphere for snapshot disk information
func (r *CheckRunner) getSnapshotDiskInfo(ctx context.Context, vmMoref, snapshotMoref string) (*types.SnapshotDiskInfo, error) {
	// Create vSphere client
	vsphereClient, err := vsphere.NewClient(
		ctx,
		r.credentials.VCenterURL,
		r.credentials.Username,
		r.credentials.Password,
		true, // insecure - accept self-signed certificates
		r.logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vSphere: %w", err)
	}
	defer vsphereClient.Close()

	// Get snapshot disk info
	info, err := vsphereClient.GetSnapshotDiskInfo(ctx, vmMoref, snapshotMoref)
	if err != nil {
		return nil, err
	}

	// Convert internal type to public type
	return &types.SnapshotDiskInfo{
		VMMoref:             info.VMMoref,
		SnapshotMoref:       info.SnapshotMoref,
		DiskPaths:           nil, // Library queries these internally
		BaseDiskPaths:       nil, // Library queries these internally
		ComputeResourcePath: info.ComputeResourcePath,
	}, nil
}
