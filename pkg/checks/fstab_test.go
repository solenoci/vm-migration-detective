package checks

import (
	"testing"

	"github.com/kubev2v/vm-migration-detective/pkg/types"
)

func TestValidateMigrateableFstab(t *testing.T) {
	tests := []struct {
		name           string
		inspectionData *types.VirtInspectorXML
		expectedValid  bool
		expectedMsg    string
	}{
		{
			name:           "nil inspection data",
			inspectionData: nil,
			expectedValid:  true,
			expectedMsg:    "No inspection data provided",
		},
		{
			name: "no mountpoints",
			inspectionData: &types.VirtInspectorXML{
				Operatingsystems: []types.VirtInspectorOS{
					{
						Name:   "linux",
						Distro: "rhel",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{},
						},
					},
				},
			},
			expectedValid: true,
			expectedMsg:   "Fstab is migrateable - no /dev/disk/by-path/ entries found",
		},
		{
			name: "valid fstab with UUID",
			inspectionData: &types.VirtInspectorXML{
				Operatingsystems: []types.VirtInspectorOS{
					{
						Name:   "linux",
						Distro: "rhel",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{
								{
									Device:     "/dev/disk/by-uuid/1234-5678",
									MountPoint: "/",
								},
								{
									Device:     "/dev/disk/by-uuid/abcd-efgh",
									MountPoint: "/boot",
								},
							},
						},
					},
				},
			},
			expectedValid: true,
			expectedMsg:   "Fstab is migrateable - no /dev/disk/by-path/ entries found",
		},
		{
			name: "valid fstab with device names",
			inspectionData: &types.VirtInspectorXML{
				Operatingsystems: []types.VirtInspectorOS{
					{
						Name:   "linux",
						Distro: "rhel",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{
								{
									Device:     "/dev/sda1",
									MountPoint: "/",
								},
								{
									Device:     "/dev/sda2",
									MountPoint: "/boot",
								},
							},
						},
					},
				},
			},
			expectedValid: true,
			expectedMsg:   "Fstab is migrateable - no /dev/disk/by-path/ entries found",
		},
		{
			name: "invalid fstab with by-path",
			inspectionData: &types.VirtInspectorXML{
				Operatingsystems: []types.VirtInspectorOS{
					{
						Name:   "linux",
						Distro: "rhel",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{
								{
									Device:     "/dev/disk/by-path/pci-0000:00:10.0-scsi-0:0:0:0-part1",
									MountPoint: "/",
								},
							},
						},
					},
				},
			},
			expectedValid: false,
			expectedMsg:   "Fstab contains /dev/disk/by-path/ entries which are not migrateable. Found device: /dev/disk/by-path/pci-0000:00:10.0-scsi-0:0:0:0-part1 mounted at: /",
		},
		{
			name: "mixed valid and invalid entries",
			inspectionData: &types.VirtInspectorXML{
				Operatingsystems: []types.VirtInspectorOS{
					{
						Name:   "linux",
						Distro: "rhel",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{
								{
									Device:     "/dev/disk/by-uuid/1234-5678",
									MountPoint: "/boot",
								},
								{
									Device:     "/dev/disk/by-path/pci-0000:00:10.0-scsi-0:0:0:0-part2",
									MountPoint: "/data",
								},
							},
						},
					},
				},
			},
			expectedValid: false,
			expectedMsg:   "Fstab contains /dev/disk/by-path/ entries which are not migrateable. Found device: /dev/disk/by-path/pci-0000:00:10.0-scsi-0:0:0:0-part2 mounted at: /data",
		},
		{
			name: "multiple operating systems - all valid",
			inspectionData: &types.VirtInspectorXML{
				Operatingsystems: []types.VirtInspectorOS{
					{
						Name:   "linux",
						Distro: "rhel",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{
								{
									Device:     "/dev/sda1",
									MountPoint: "/",
								},
							},
						},
					},
					{
						Name:   "linux",
						Distro: "ubuntu",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{
								{
									Device:     "/dev/disk/by-uuid/abcd-1234",
									MountPoint: "/",
								},
							},
						},
					},
				},
			},
			expectedValid: true,
			expectedMsg:   "Fstab is migrateable - no /dev/disk/by-path/ entries found",
		},
		{
			name: "multiple operating systems - one invalid",
			inspectionData: &types.VirtInspectorXML{
				Operatingsystems: []types.VirtInspectorOS{
					{
						Name:   "linux",
						Distro: "rhel",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{
								{
									Device:     "/dev/sda1",
									MountPoint: "/",
								},
							},
						},
					},
					{
						Name:   "linux",
						Distro: "ubuntu",
						Mountpoints: types.VirtInspectorMountpoints{
							Mountpoint: []types.VirtInspectorMountpoint{
								{
									Device:     "/dev/disk/by-path/pci-0000:00:10.0-scsi-0:0:0:0",
									MountPoint: "/mnt",
								},
							},
						},
					},
				},
			},
			expectedValid: false,
			expectedMsg:   "Fstab contains /dev/disk/by-path/ entries which are not migrateable. Found device: /dev/disk/by-path/pci-0000:00:10.0-scsi-0:0:0:0 mounted at: /mnt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMigrateableFstab(tt.inspectionData)

			if result.Valid != tt.expectedValid {
				t.Errorf("ValidateMigrateableFstab() Valid = %v, want %v", result.Valid, tt.expectedValid)
			}

			if result.Message != tt.expectedMsg {
				t.Errorf("ValidateMigrateableFstab() Message = %v, want %v", result.Message, tt.expectedMsg)
			}
		})
	}
}
