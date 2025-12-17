# vm-migration-detective

VM validations library for migration preparation.

## Overview

This library provides tools and utilities for inspecting and validating virtual machines before migration. It includes support for various inspection methods including virt-inspector and virt-v2v-inspector with VDDK integration.

## Structure

- **pkg/types**: Public types and data structures
  - `types.go`: Core types
    - `SnapshotDiskInfo`: VM snapshot disk information for VDDK access
  - `virt_inspector.go`: virt-inspector XML data structures
    - `VirtInspectorXML`: Root structure for virt-inspector output
    - OS information, applications, filesystems, mountpoints, drives
  - `virt_v2v_inspector.go`: virt-v2v-inspector XML data structures
    - `VirtV2VInspectorXML`: Root structure for virt-v2v-inspector output
    - OS information and firmware details

- **pkg/inspection**: Public bridge to inspection functionality (temporary for external usage)
  - Re-exports internal inspection types and functions

- **internal/inspection**: Core inspection implementation
  - `virt_inspector.go`: libguestfs virt-inspector integration with NBDKit/VDDK
  - `virt_v2v_inspector.go`: virt-v2v-inspector integration with VDDK direct access
  - `nbdkit_vddk.go`: NBDKit server with VDDK plugin for snapshot access
  - `virt_v2v_open.go`: virt-v2v-open NBD server integration

## Usage

Import the library in your Go project:

```go
import (
    "github.com/kubev2v/vm-migration-detective/pkg/inspection"
    "github.com/kubev2v/vm-migration-detective/pkg/types"
)
```

### Inspecting a VM with virt-inspector

```go
inspector := inspection.NewVirtInspector("", 30*time.Minute, logger)
inspectionData, err := inspector.Inspect(
    ctx,
    vmName,
    snapshotName,
    vcenterURL,
    datacenter,
    username,
    password,
    diskInfo,
)
// inspectionData is *types.VirtInspectorXML with OS info, apps, filesystems, etc.
```

### Inspecting a VM with virt-v2v-inspector

```go
inspector := inspection.NewVirtV2vInspector("", 30*time.Minute, logger)
inspectionData, err := inspector.Inspect(
    ctx,
    vmName,
    snapshotName,
    vcenterURL,
    datacenter,
    username,
    password,
    diskInfo,
    "no_verify=1", // SSL verification option
)
// inspectionData is *types.VirtV2VInspectorXML with OS and firmware info
```

## Development

See the Makefile for available targets:
- `make validate-all`: Run all validations (lint, format check, tidy check)
- `make lint`: Run golangci-lint
- `make format`: Format code with goimports
- `make check-format`: Check if code formatting is correct
- `make tidy`: Tidy go modules
- `make tidy-check`: Check if go.mod and go.sum are tidy
- `make verify`: Verify the code compiles
- `make clean`: Clean build artifacts and downloaded tools

## Requirements

- Go 1.24.0 or later
- libguestfs tools (virt-inspector)
- virt-v2v tools (virt-v2v-inspector)
- NBDKit with VDDK plugin (optional, for VDDK support)
