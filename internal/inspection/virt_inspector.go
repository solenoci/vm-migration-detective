package inspection

import (
	"context"
	"encoding/xml"
	"fmt"
	"os/exec"
	"time"

	"github.com/kubev2v/vm-migration-detective/pkg/types"
	"github.com/sirupsen/logrus"
)

// UseVirtV2VOpen controls whether to use virt-v2v-open (true) or nbdkit directly (false)
// Default is false (use nbdkit directly)
const UseVirtV2VOpen = false

// Inspector handles VM inspection operations
type VirtInspector struct {
	virtInspectorPath string
	timeout           time.Duration
	logger            *logrus.Logger
}

// NewInspector creates a new Inspector instance
func NewVirtInspector(virtInspectorPath string, timeout time.Duration, logger *logrus.Logger) *VirtInspector {
	if virtInspectorPath == "" {
		virtInspectorPath = "virt-inspector" // Use system PATH
	}
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	return &VirtInspector{
		virtInspectorPath: virtInspectorPath,
		timeout:           timeout,
		logger:            logger,
	}
}

func (i *VirtInspector) Inspect(
	ctx context.Context,
	vmName string,
	snapshotName string,
	vcenterURL string,
	datacenter string,
	username string,
	password string,
	diskInfo *types.SnapshotDiskInfo, // Snapshot disk info from vm_service
) (*types.VirtInspectorXML, error) {

	var nbdURLs []string
	var sessionCloser func()

	if UseVirtV2VOpen {
		i.logger.WithFields(logrus.Fields{
			"vm_name":       vmName,
			"snapshot_name": snapshotName,
			"vcenter_url":   vcenterURL,
			"datacenter":    datacenter,
		}).Info("Running virt-inspector using virt-v2v-open (VDDK + snapshot)")

		openCtx, cancel := context.WithTimeout(ctx, i.timeout)
		defer cancel()

		v2vSession, err := OpenWithVirtV2V(
			openCtx,
			vmName,
			datacenter,
			snapshotName,
			vcenterURL,
			username,
			password,
		)
		if err != nil {
			return nil, err
		}
		nbdURLs = []string{v2vSession.NBDURL}
		sessionCloser = v2vSession.Close

		// Give NBD time to initialize
		time.Sleep(4 * time.Second)
	} else {
		i.logger.WithFields(logrus.Fields{
			"vm_name":       vmName,
			"snapshot_name": snapshotName,
			"vcenter_url":   vcenterURL,
			"datacenter":    datacenter,
		}).Info("Running virt-inspector using nbdkit-vddk (VDDK + snapshot)")

		// Use diskInfo passed from vm_service (no need to query vSphere here)
		i.logger.WithFields(logrus.Fields{
			"vm_moref":       diskInfo.VMMoref,
			"snapshot_moref": diskInfo.SnapshotMoref,
			"disk_count":     len(diskInfo.DiskPaths),
			"disk_paths":     diskInfo.DiskPaths,
			"base_disk_paths": diskInfo.BaseDiskPaths,
		}).Debug("Using snapshot disk info from vm_service")

		openCtx, cancel := context.WithTimeout(ctx, i.timeout)
		defer cancel()

		// Start one NBDkit session per disk
		var nbdkitSessions []*NBDKitSession

		for idx, baseDiskPath := range diskInfo.BaseDiskPaths {
			i.logger.WithFields(logrus.Fields{
				"disk_index":     idx,
				"base_disk_path": baseDiskPath,
			}).Debug("Starting NBDkit session for disk")

			nbdkitSession, err := OpenWithNBDKitVDDK(
				openCtx,
				diskInfo.VMMoref,
				diskInfo.SnapshotMoref,
				baseDiskPath,
				vcenterURL,
				username,
				password,
				i.logger,
			)
			if err != nil {
				// Close any sessions we've already created
				for _, session := range nbdkitSessions {
					session.Close()
				}
				return nil, fmt.Errorf("failed to start NBDkit session for disk %d: %w", idx, err)
			}
			nbdkitSessions = append(nbdkitSessions, nbdkitSession)
			nbdURLs = append(nbdURLs, nbdkitSession.NBDURL)

			// Wait for NBD server to be ready (more reliable than sleep)
			if err := nbdkitSession.WaitForReady(30 * time.Second); err != nil {
				i.logger.WithError(err).WithField("disk_index", idx).Error("NBD server not ready")
				// Close all sessions
				for _, session := range nbdkitSessions {
					session.Close()
				}
				return nil, fmt.Errorf("NBD server not ready for disk %d: %w", idx, err)
			}
		}

		// Create a cleanup function that closes all sessions
		sessionCloser = func() {
			for _, session := range nbdkitSessions {
				session.Close()
			}
		}
	}
	defer sessionCloser()

	inspectCtx, cancel := context.WithTimeout(ctx, i.timeout)
	defer cancel()

	i.logger.WithFields(logrus.Fields{
		"nbd_urls":   nbdURLs,
		"disk_count": len(nbdURLs),
	}).Info("Running virt-inspector on NBD")

	// Build command with multiple -a options for all disks
	// Format must be specified before each -a parameter
	var aOptions string
	for _, url := range nbdURLs {
		aOptions += fmt.Sprintf(" --format=raw -a '%s'", url)
	}
	cmdString := fmt.Sprintf("unset LD_LIBRARY_PATH && %s%s",
		i.virtInspectorPath, aOptions)

	virtInspectorCmd := exec.CommandContext(inspectCtx, "sh", "-c", cmdString)

	output, err := virtInspectorCmd.CombinedOutput()
	outputStr := string(output)
	if err != nil {
		// Get exit code if available
		exitCode := -1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		i.logger.WithFields(logrus.Fields{
			"output":     outputStr,
			"exit_code":  exitCode,
			"nbd_urls":   nbdURLs,
			"disk_count": len(nbdURLs),
			"command":    cmdString,
		}).Error("virt-inspector failed")

		// Include output in error message for better debugging
		if outputStr != "" {
			return nil, fmt.Errorf("virt-inspector failed (exit code %d): %w\nOutput: %s", exitCode, err, outputStr)
		}
		return nil, fmt.Errorf("virt-inspector failed (exit code %d): %w", exitCode, err)
	}

	inspectionData, err := parseInspectionXML(output)
	if err != nil {
		if i.logger != nil {
			i.logger.WithFields(logrus.Fields{
				"error":  err,
				"output": outputStr,
			}).Error("Failed to parse virt-inspector XML output")
		}
		return nil, fmt.Errorf("failed to parse inspection output: %w", err)
	}

	if UseVirtV2VOpen {
		i.logger.Info("virt-v2v-open snapshot inspection completed successfully")
	} else {
		i.logger.Info("nbdkit-vddk snapshot inspection completed successfully")
	}
	return inspectionData, nil
}

// parseInspectionXML parses virt-inspector XML output and returns the native XML structure
func parseInspectionXML(xmlData []byte) (*types.VirtInspectorXML, error) {
	var xmlRoot types.VirtInspectorXML
	err := xml.Unmarshal(xmlData, &xmlRoot)
	if err != nil {
		return nil, fmt.Errorf("XML parsing error: %w", err)
	}

	if len(xmlRoot.Operatingsystems) == 0 {
		return nil, fmt.Errorf("no operating systems found in inspection output")
	}

	return &xmlRoot, nil
}
