package vmdetect

// CheckType represents the type of validation check
type CheckType string

const (
	// CheckTypeFstab validates fstab entries for migration compatibility
	CheckTypeFstab CheckType = "fstab"
	// CheckTypeDiskAccess validates that the disk is accessible (not encrypted)
	CheckTypeDiskAccess CheckType = "disk-access"
)

// AllCheckTypes returns all available check types
func AllCheckTypes() []CheckType {
	return []CheckType{
		CheckTypeFstab,
		CheckTypeDiskAccess,
	}
}

// IsValidCheckType checks if a check type is valid
func IsValidCheckType(checkType CheckType) bool {
	for _, ct := range AllCheckTypes() {
		if ct == checkType {
			return true
		}
	}
	return false
}
