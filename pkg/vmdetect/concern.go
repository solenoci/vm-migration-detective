package vmdetect

import "github.com/kubev2v/vm-migration-detective/pkg/checks"

// CheckResult represents the result of a single validation check
type CheckResult struct {
	// CheckType indicates which check was performed
	CheckType CheckType `json:"check_type"`
	// Passed indicates whether the check passed (true) or found concerns (false)
	Passed bool `json:"passed"`
	// Concerns contains all issues found by this check (empty if passed)
	Concerns []checks.Concern `json:"concerns,omitempty"`
	// Error contains the error message if an unexpected error occurred
	Error *string `json:"error,omitempty"`
}
