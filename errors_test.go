// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026, the go-tpm2/common authors. All rights reserved.

package common

import "testing"

func TestErrorString(t *testing.T) {
	e := Error("boom")
	if e.Error() != "boom" {
		t.Fatalf("Error() = %q", e.Error())
	}
	// Sentinel comparison must work across the error interface.
	var err error = ErrShortBuffer
	if err != ErrShortBuffer {
		t.Fatal("sentinel comparison failed")
	}
}
