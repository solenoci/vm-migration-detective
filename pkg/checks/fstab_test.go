package checks

import (
	"testing"

	"github.com/kubev2v/vm-migration-detective/pkg/types"
)

func TestValidateMigrateableFstab(t *testing.T) {
	tests := []struct {
		name              string
		inspectionData    *types.VirtInspectorXML
		expectedPassed    bool
		expectedConcernID string
	}{
		{
			name:           "nil inspection data",
			inspectionData: nil,
			expectedPassed: true,
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
			expectedPassed: true,
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
			expectedPassed: true,
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
			expectedPassed: true,
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
			expectedPassed:    false,
			expectedConcernID: "fstab-by-path-device",
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
			expectedPassed:    false,
			expectedConcernID: "fstab-by-path-device",
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
			expectedPassed: true,
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
			expectedPassed:    false,
			expectedConcernID: "fstab-by-path-device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMigrateableFstab(tt.inspectionData)

			if result.Passed != tt.expectedPassed {
				t.Errorf("ValidateMigrateableFstab() Passed = %v, want %v", result.Passed, tt.expectedPassed)
			}

			if !tt.expectedPassed && tt.expectedConcernID != "" {
				if len(result.Concerns) == 0 {
					t.Errorf("ValidateMigrateableFstab() expected concerns but got none")
				} else if result.Concerns[0].ID != tt.expectedConcernID {
					t.Errorf("ValidateMigrateableFstab() Concern ID = %v, want %v", result.Concerns[0].ID, tt.expectedConcernID)
				}
			}

			if tt.expectedPassed && len(result.Concerns) > 0 {
				t.Errorf("ValidateMigrateableFstab() expected no concerns but got %v", result.Concerns)
			}
		})
	}
}
