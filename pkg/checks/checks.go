package checks

import (
	"context"

	"github.com/kubev2v/vm-migration-detective/pkg/persistent"
	"github.com/kubev2v/vm-migration-detective/pkg/types"
	"github.com/sirupsen/logrus"
)

// CheckResult represents the result of a validation check
type CheckResult struct {
	// Valid indicates whether the check passed
	Valid bool
	// Message provides additional context about the check result
	Message string
}

// InspectionParams contains all parameters needed to run an inspection
type InspectionParams struct {
	Ctx          context.Context
	VMName       string
	SnapshotName string
	Datacenter   string
	VCenterURL   string
	Username     string
	Password     string
	DiskInfo     *types.SnapshotDiskInfo
	DB           persistent.DB // Optional database for persistent caching
	Logger       *logrus.Logger // Optional logger (can be nil)
}

// Check defines the interface for VM validation checks
type Check interface {
	// Run executes the check by internally creating the inspector,
	// calling the appropriate inspection, and validating the results
	Run(params InspectionParams) CheckResult
}
