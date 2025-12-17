package checks

import (
	"strings"
	"time"

	"github.com/kubev2v/vm-migration-detective/pkg/persistent"
	"github.com/kubev2v/vm-migration-detective/pkg/types"
)

// FstabCheck validates fstab entries for migration compatibility
type FstabCheck struct{}

// NewFstabCheck creates a new FstabCheck instance
func NewFstabCheck() *FstabCheck {
	return &FstabCheck{}
}

// Run executes the fstab validation check
// It internally creates the inspector, calls the inspection, and validates the fstab data
func (c *FstabCheck) Run(params InspectionParams) CheckResult {
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

	// Call the inspection using VirtInspector (fstab data comes from virt-inspector)
	inspectionData, err := inspector.InspectWithVirt(
		params.Ctx,
		params.VMName,
		params.SnapshotName,
		params.Datacenter,
		params.DiskInfo,
	)
	if err != nil {
		return CheckResult{
			Valid:   false,
			Message: "Failed to run inspection: " + err.Error(),
		}
	}

	// Validate the fstab data
	return ValidateMigrateableFstab(inspectionData)
}

// ValidateMigrateableFstab checks if the VM's fstab is migrateable
// Returns false if fstab created by path (/dev/disk/by-path/), otherwise returns true
func ValidateMigrateableFstab(inspectionData *types.VirtInspectorXML) CheckResult {
	if inspectionData == nil {
		return CheckResult{
			Valid:   true,
			Message: "No inspection data provided",
		}
	}

	// Check all operating systems in the inspection data
	for _, os := range inspectionData.Operatingsystems {
		// Check all mountpoints for path-based device references
		for _, mountpoint := range os.Mountpoints.Mountpoint {
			if strings.HasPrefix(mountpoint.Device, "/dev/disk/by-path/") {
				return CheckResult{
					Valid: false,
					Message: "Fstab contains /dev/disk/by-path/ entries which are not migrateable. " +
						"Found device: " + mountpoint.Device + " mounted at: " + mountpoint.MountPoint,
				}
			}
		}
	}

	return CheckResult{
		Valid:   true,
		Message: "Fstab is migrateable - no /dev/disk/by-path/ entries found",
	}
}
