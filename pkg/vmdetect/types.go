package vmdetect

import (
	"github.com/kubev2v/vm-migration-detective/internal/checks"
	"github.com/kubev2v/vm-migration-detective/internal/persistent"
)

// Re-export types that external packages need to use

// Credentials holds vCenter access credentials
type Credentials = persistent.Credentials

// DB defines the interface for persisting inspection data
type DB = persistent.DB

// CacheKey represents a unique identifier for a VM+snapshot pair
type CacheKey = persistent.CacheKey

// Concern represents a validation concern found during checks
type Concern = checks.Concern

// ConcernCategory represents the severity level of a concern
type ConcernCategory = checks.ConcernCategory

// Concern severity categories
const (
	ConcernCategoryCritical    = checks.ConcernCategoryCritical
	ConcernCategoryWarning     = checks.ConcernCategoryWarning
	ConcernCategoryInformation = checks.ConcernCategoryInformation
	ConcernCategoryAdvisory    = checks.ConcernCategoryAdvisory
	ConcernCategoryError       = checks.ConcernCategoryError
)
