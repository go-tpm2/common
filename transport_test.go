// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026, the go-tpm2/common authors. All rights reserved.

package common

import (
	"bytes"
	"testing"
)

// loopTransport is a trivial Transport stub: it echoes a fixed response
// built around whatever command code it is sent, to exercise the
// BuildCommand -> Transport.Send -> ParseResponse round trip.
type loopTransport struct {
	lastCmd []byte
}

func (l *loopTransport) Send(cmd []byte) ([]byte, error) {
	l.lastCmd = cmd
	// Echo a minimal successful response carrying one param byte.
	return BuildResponse(uint16(TagNoSessions), uint32(RCSuccess), []byte{0x42}), nil
}

// BuildResponse mirrors BuildCommand for test fixtures: it lays out a
// TPM 2.0 response header with a consistent responseSize. It lives in
// the test file because production code never builds responses.
func BuildResponse(tag uint16, rc uint32, params []byte) []byte {
	size := uint32(HeaderSize + len(params))
	var b []byte
	b = PutU16(b, tag)
	b = PutU32(b, size)
	b = PutU32(b, rc)
	return append(b, params...)
}

func TestTransportRoundTrip(t *testing.T) {
	var tr Transport = &loopTransport{}
	cmd := BuildCommand(uint16(TagNoSessions), uint32(CCGetRandom), []byte{0x00, 0x02})
	rsp, err := tr.Send(cmd)
	if err != nil {
		t.Fatalf("Send err = %v", err)
	}
	tag, rc, params, err := ParseResponse(rsp)
	if err != nil {
		t.Fatalf("ParseResponse err = %v", err)
	}
	if tag != uint16(TagNoSessions) || rc != uint32(RCSuccess) {
		t.Fatalf("tag=%#x rc=%#x", tag, rc)
	}
	if !bytes.Equal(params, []byte{0x42}) {
		t.Fatalf("params = %x", params)
	}
}
