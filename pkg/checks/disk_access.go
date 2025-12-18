package checks

import (
	"strings"
	"time"

	"github.com/kubev2v/vm-migration-detective/pkg/persistent"
)

// DiskAccessCheck validates that the disk is accessible (not encrypted)
type DiskAccessCheck struct{}

// NewDiskAccessCheck creates a new DiskAccessCheck instance
func NewDiskAccessCheck() *DiskAccessCheck {
	return &DiskAccessCheck{}
}

// Run executes the disk access validation check
// It tries to run virt-inspector and checks if the disk is encrypted
// Returns false (not valid) if encrypted disk error is detected, true otherwise
func (c *DiskAccessCheck) Run(params InspectionParams) CheckResult {
	// Create persistent inspector internally
	credentials := persistent.Credentials{
		VCenterURL: params.VCenterURL,
		Username:   params.Username,
		Password:   params.Password,
	}

	inspector := persistent.NewInspector(
		"",             // Use system PATH for virt-inspector
		"",             // Use system PATH for virt-v2v-inspector
		30*time.Minute, // Timeout
		credentials,
		params.Logger, // Logger from params
		params.DB,     // Optional database for persistent caching
	)

	// Try to run the inspection
	_, err := inspector.InspectWithVirt(
		params.Ctx,
		params.VMName,
		params.SnapshotName,
		params.Datacenter,
		params.DiskInfo,
	)

	if err != nil {
		// Check if the error message indicates an encrypted disk
		// The inspection function already detects encrypted disks and returns a specific error message
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "disk encryption detected") ||
			strings.Contains(errStr, "encrypted disk") ||
			strings.Contains(errStr, "cannot access encrypted") {
			// Encrypted disk is a known validation failure, not an unexpected error
			return CheckResult{
				Valid:   false,
				Message: err.Error(),
				Error:   nil,
			}
		}

		// Other errors are unexpected errors
		errMsg := err.Error()
		return CheckResult{
			Valid:   false,
			Message: "Unexpected error",
			Error:   &errMsg,
		}
	}

	// Inspection succeeded - disk is accessible
	return CheckResult{
		Valid:   true,
		Message: "Disk is accessible and can be inspected",
	}
}
