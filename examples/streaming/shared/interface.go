// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package shared

import (
	"context"
)

type Streamer interface {
	Configure(ctx context.Context, path string, chunkSize int64) error
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, b []byte) error
}
