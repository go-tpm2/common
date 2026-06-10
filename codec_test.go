// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026, the go-tpm2/common authors. All rights reserved.

package common

import (
	"bytes"
	"testing"
)

func TestPutGetU8(t *testing.T) {
	b := PutU8(nil, 0xAB)
	if !bytes.Equal(b, []byte{0xAB}) {
		t.Fatalf("PutU8 = %x", b)
	}
	v, ok := GetU8(b, 0)
	if !ok || v != 0xAB {
		t.Fatalf("GetU8 = %#x ok=%v", v, ok)
	}
	if _, ok := GetU8(b, 1); ok {
		t.Fatal("GetU8 past end should fail")
	}
	if _, ok := GetU8(b, -1); ok {
		t.Fatal("GetU8 negative off should fail")
	}
}

func TestPutGetU16(t *testing.T) {
	b := PutU16(nil, 0x1234)
	if !bytes.Equal(b, []byte{0x12, 0x34}) {
		t.Fatalf("PutU16 big-endian = %x", b)
	}
	v, ok := GetU16(b, 0)
	if !ok || v != 0x1234 {
		t.Fatalf("GetU16 = %#x ok=%v", v, ok)
	}
	if _, ok := GetU16(b, 1); ok {
		t.Fatal("GetU16 overrun should fail")
	}
	if _, ok := GetU16(b, -1); ok {
		t.Fatal("GetU16 negative off should fail")
	}
}

func TestPutGetU32(t *testing.T) {
	b := PutU32(nil, 0x12345678)
	if !bytes.Equal(b, []byte{0x12, 0x34, 0x56, 0x78}) {
		t.Fatalf("PutU32 big-endian = %x", b)
	}
	v, ok := GetU32(b, 0)
	if !ok || v != 0x12345678 {
		t.Fatalf("GetU32 = %#x ok=%v", v, ok)
	}
	if _, ok := GetU32(b, 1); ok {
		t.Fatal("GetU32 overrun should fail")
	}
	if _, ok := GetU32(b, -1); ok {
		t.Fatal("GetU32 negative off should fail")
	}
}

func TestPutGetU64(t *testing.T) {
	b := PutU64(nil, 0x0102030405060708)
	want := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if !bytes.Equal(b, want) {
		t.Fatalf("PutU64 big-endian = %x", b)
	}
	v, ok := GetU64(b, 0)
	if !ok || v != 0x0102030405060708 {
		t.Fatalf("GetU64 = %#x ok=%v", v, ok)
	}
	if _, ok := GetU64(b, 1); ok {
		t.Fatal("GetU64 overrun should fail")
	}
	if _, ok := GetU64(b, -1); ok {
		t.Fatal("GetU64 negative off should fail")
	}
}

func TestBuildCommand(t *testing.T) {
	params := []byte{0xDE, 0xAD}
	cmd := BuildCommand(uint16(TagNoSessions), uint32(CCGetRandom), params)
	want := []byte{
		0x80, 0x01, // tag
		0x00, 0x00, 0x00, 0x0C, // commandSize = 10 + 2
		0x00, 0x00, 0x01, 0x7B, // commandCode GetRandom
		0xDE, 0xAD, // params
	}
	if !bytes.Equal(cmd, want) {
		t.Fatalf("BuildCommand = %x want %x", cmd, want)
	}
}

func TestBuildCommandNoParams(t *testing.T) {
	cmd := BuildCommand(uint16(TagNoSessions), uint32(CCSelfTest), nil)
	if len(cmd) != HeaderSize {
		t.Fatalf("len = %d want %d", len(cmd), HeaderSize)
	}
	size, _ := GetU32(cmd, 2)
	if int(size) != HeaderSize {
		t.Fatalf("commandSize = %d want %d", size, HeaderSize)
	}
}

func TestParseResponseRoundTrip(t *testing.T) {
	params := []byte{0x01, 0x02, 0x03}
	// Construct a valid response: tag | size | rc | params.
	var rsp []byte
	rsp = PutU16(rsp, uint16(TagSessions))
	rsp = PutU32(rsp, uint32(HeaderSize+len(params)))
	rsp = PutU32(rsp, uint32(RCSuccess))
	rsp = append(rsp, params...)

	tag, rc, got, err := ParseResponse(rsp)
	if err != nil {
		t.Fatalf("ParseResponse err = %v", err)
	}
	if tag != uint16(TagSessions) {
		t.Fatalf("tag = %#x", tag)
	}
	if rc != uint32(RCSuccess) {
		t.Fatalf("rc = %#x", rc)
	}
	if !bytes.Equal(got, params) {
		t.Fatalf("params = %x want %x", got, params)
	}
}

func TestParseResponseShort(t *testing.T) {
	_, _, _, err := ParseResponse([]byte{0x80, 0x01, 0x00})
	if err != ErrShortBuffer {
		t.Fatalf("err = %v want ErrShortBuffer", err)
	}
}

func TestParseResponseSizeMismatch(t *testing.T) {
	// Declares size 99 but buffer is only HeaderSize bytes.
	var rsp []byte
	rsp = PutU16(rsp, uint16(TagNoSessions))
	rsp = PutU32(rsp, 99)
	rsp = PutU32(rsp, 0)
	_, _, _, err := ParseResponse(rsp)
	if err != ErrSizeMismatch {
		t.Fatalf("err = %v want ErrSizeMismatch", err)
	}
}

func TestMarshalUnmarshalTPM2B(t *testing.T) {
	payload := []byte{0xCA, 0xFE, 0xBA, 0xBE}
	m := MarshalTPM2B(payload)
	want := append([]byte{0x00, 0x04}, payload...)
	if !bytes.Equal(m, want) {
		t.Fatalf("MarshalTPM2B = %x want %x", m, want)
	}

	// Append trailing bytes to verify rest is returned correctly.
	trailer := []byte{0x99, 0x88}
	buf := append(append([]byte{}, m...), trailer...)
	val, rest, err := UnmarshalTPM2B(buf)
	if err != nil {
		t.Fatalf("UnmarshalTPM2B err = %v", err)
	}
	if !bytes.Equal(val, payload) {
		t.Fatalf("val = %x want %x", val, payload)
	}
	if !bytes.Equal(rest, trailer) {
		t.Fatalf("rest = %x want %x", rest, trailer)
	}
}

func TestMarshalTPM2BEmpty(t *testing.T) {
	m := MarshalTPM2B(nil)
	if !bytes.Equal(m, []byte{0x00, 0x00}) {
		t.Fatalf("MarshalTPM2B(nil) = %x", m)
	}
}

func TestUnmarshalTPM2BShortSize(t *testing.T) {
	_, _, err := UnmarshalTPM2B([]byte{0x00})
	if err != ErrShortBuffer {
		t.Fatalf("err = %v want ErrShortBuffer", err)
	}
}

func TestUnmarshalTPM2BShortPayload(t *testing.T) {
	// Declares size 5 but only 2 payload bytes present.
	_, _, err := UnmarshalTPM2B([]byte{0x00, 0x05, 0xAA, 0xBB})
	if err != ErrShortBuffer {
		t.Fatalf("err = %v want ErrShortBuffer", err)
	}
}
