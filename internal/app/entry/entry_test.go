// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package entry

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("regular", func(t *testing.T) {
		err := testRun([]string{"-c", filepath.Join(t.TempDir(), "config.yaml")})
		assert.NoError(t, err)
	})
	t.Run("help", func(t *testing.T) {
		err := testRun([]string{"-h"})
		assert.NoError(t, err)
	})
	t.Run("wrong args", func(t *testing.T) {
		err := testRun([]string{"-test"})
		assert.Error(t, err)
	})
	t.Run("miss config", func(t *testing.T) {
		err := testRun([]string{"-c", filepath.Join(t.TempDir(), "config.yaml"), "-disable-generate"})
		assert.Error(t, err)
	})
	t.Run("miss config", func(t *testing.T) {
		err := testRun([]string{"-c", filepath.Join(t.TempDir(), "config.yaml"), "-disable-generate"})
		assert.Error(t, err)
	})
}

func testRun(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		retErr error
		done   = make(chan struct{})
	)

	go func() {
		retErr = Run(ctx, "unit_test", args)
		close(done)
	}()

	select {
	case <-done:
		return retErr
	case <-time.After(5 * time.Second):
		cancel()
		<-done
		return retErr
	}
}
