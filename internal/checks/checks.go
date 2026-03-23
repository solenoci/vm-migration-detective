package checks

import (
	"context"

	"github.com/kubev2v/vm-migration-detective/internal/persistent"
	pkgchecks "github.com/kubev2v/vm-migration-detective/pkg/checks"
	"github.com/kubev2v/vm-migration-detective/pkg/types"
)

// Re-export public types from pkg/checks
type Concern = pkgchecks.Concern
type ConcernCategory = pkgchecks.ConcernCategory

// Re-export concern severity categories
const (
	ConcernCategoryCritical    = pkgchecks.ConcernCategoryCritical
	ConcernCategoryWarning     = pkgchecks.ConcernCategoryWarning
	ConcernCategoryInformation = pkgchecks.ConcernCategoryInformation
	ConcernCategoryAdvisory    = pkgchecks.ConcernCategoryAdvisory
	ConcernCategoryError       = pkgchecks.ConcernCategoryError
)

// CheckResult represents the result of a validation check
type CheckResult struct {
	// Passed indicates whether the check passed (true) or found concerns (false)
	Passed bool
	// Concerns contains all issues found by this check (empty if passed)
	Concerns []Concern
	// Error contains the error message if an unexpected error occurred, nil otherwise
	Error *string `json:"error,omitempty"`
}

// InspectionParams contains all parameters needed to run an inspection
type InspectionParams struct {
	Ctx           context.Context
	VMMoref       string
	SnapshotMoref string
	DiskInfo      *types.SnapshotDiskInfo
	Inspector     persistent.InspectorInterface // Use interface from internal/persistent
}

// Check defines the interface for VM validation checks
type Check interface {
	// Run executes the check by internally creating the inspector,
	// calling the appropriate inspection, and validating the results
	Run(params InspectionParams) CheckResult
}
