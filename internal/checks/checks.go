package checks

import (
	"context"

	"github.com/kubev2v/vm-migration-detective/internal/persistent"
	"github.com/kubev2v/vm-migration-detective/pkg/types"
)

// ConcernCategory represents the severity level of a concern
type ConcernCategory string

const (
	// ConcernCategoryCritical indicates a critical issue that must be resolved
	ConcernCategoryCritical ConcernCategory = "Critical"
	// ConcernCategoryWarning indicates a warning that should be addressed
	ConcernCategoryWarning ConcernCategory = "Warning"
	// ConcernCategoryInformation indicates informational message
	ConcernCategoryInformation ConcernCategory = "Information"
	// ConcernCategoryAdvisory indicates an advisory recommendation
	ConcernCategoryAdvisory ConcernCategory = "Advisory"
	// ConcernCategoryError indicates an error occurred during the check
	ConcernCategoryError ConcernCategory = "Error"
)

// Concern represents a validation concern found during checks
type Concern struct {
	// ID is the unique identifier for this concern type
	ID string `json:"id"`
	// Category indicates the severity level of the concern
	Category ConcernCategory `json:"category"`
	// Label is a human-readable short description
	Label string `json:"label"`
	// Message provides detailed information about the concern
	Message string `json:"message"`
}

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
	Inspector     *persistent.Inspector // Shared inspector instance
}

// Check defines the interface for VM validation checks
type Check interface {
	// Run executes the check by internally creating the inspector,
	// calling the appropriate inspection, and validating the results
	Run(params InspectionParams) CheckResult
}
