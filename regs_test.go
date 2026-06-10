// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026, the go-tpm2/common authors. All rights reserved.

package common

import (
	"bytes"
	"testing"
)

// memRegs is a flat byte-addressed stub implementing Regs, modelling a
// little TPM register window for the byte-stream helpers.
type memRegs struct {
	mem []byte
}

func newMemRegs(n int) *memRegs { return &memRegs{mem: make([]byte, n)} }

func (m *memRegs) Read8(off uint32) uint8 { return m.mem[off] }

func (m *memRegs) Read32(off uint32) uint32 {
	v, _ := GetU32(m.mem, int(off))
	return v
}

func (m *memRegs) Write8(off uint32, v uint8) { m.mem[off] = v }

func (m *memRegs) Write32(off uint32, v uint32) {
	m.mem[off] = byte(v >> 24)
	m.mem[off+1] = byte(v >> 16)
	m.mem[off+2] = byte(v >> 8)
	m.mem[off+3] = byte(v)
}

func TestRegs8And32(t *testing.T) {
	r := newMemRegs(16)
	r.Write8(0, 0xAB)
	if got := r.Read8(0); got != 0xAB {
		t.Fatalf("Read8 = %#x", got)
	}
	r.Write32(4, 0x11223344)
	if got := r.Read32(4); got != 0x11223344 {
		t.Fatalf("Read32 = %#x", got)
	}
}

func TestWriteReadBytes(t *testing.T) {
	r := newMemRegs(32)
	data := []byte{0x10, 0x20, 0x30, 0x40, 0x50}
	WriteBytes(r, 8, data)

	out := make([]byte, len(data))
	ReadBytes(r, 8, out)
	if !bytes.Equal(out, data) {
		t.Fatalf("ReadBytes = %x want %x", out, data)
	}
}

func TestWriteReadBytesEmpty(t *testing.T) {
	r := newMemRegs(4)
	WriteBytes(r, 0, nil)
	ReadBytes(r, 0, nil)
	// Nothing to assert beyond not panicking; the loops must handle
	// zero-length slices.
}
