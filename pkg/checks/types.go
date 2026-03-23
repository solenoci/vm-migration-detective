package checks

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
