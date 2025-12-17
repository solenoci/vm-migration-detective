package persistent

// This package provides a public API bridge to the internal persistent package.

import (
	"github.com/kubev2v/vm-migration-detective/internal/persistent"
)

// Re-export persistent types
type (
	Inspector   = persistent.Inspector
	Credentials = persistent.Credentials
	CacheKey    = persistent.CacheKey
	DB          = persistent.DB
)

// Re-export constructor functions
var (
	NewInspector = persistent.NewInspector
)
