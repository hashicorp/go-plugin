// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !windows
// +build !windows

package plugin

var serverListener = serverListener_unix

func isSupportUnix() bool {
	return true
}
