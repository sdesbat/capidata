package postgresprotocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func msg(kind byte, body []byte) []byte {
	out := []byte{kind, 0, 0, 0, 0}
	binary.BigEndian.PutUint32(out[1:5], uint32(len(body)+4))
	return append(out, body...)
}

func TestParserAndBind(t *testing.T) {
	parse := msg('P', append(append([]byte("s1\x00UPDATE accounts SET balance=$1 WHERE id=$2\x00"), 0, 0), 0, 0))
	bindBody := []byte("\x00s1\x00")
	bindBody = append(bindBody, 0, 0) // no parameter format codes: text defaults
	bindBody = append(bindBody, 0, 2) // two parameters
	bindBody = append(bindBody, 0, 0, 0, 3, '1', '0', '0')
	bindBody = append(bindBody, 0, 0, 0, 2, '4', '2')
	bindBody = append(bindBody, 0, 0) // no result format codes
	msgs, err := (&Parser{}).Feed(append(parse, msg('B', bindBody)...))
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Fatalf("messages: %d", len(msgs))
	}
	p, err := ParseParse(msgs[0])
	if err != nil || p.Query == "" {
		t.Fatalf("parse: %+v %v", p, err)
	}
	b, err := ParseBind(msgs[1])
	if err != nil || len(b.Parameters) != 2 {
		t.Fatalf("bind: %+v %v", b, err)
	}
	if !bytes.Equal(b.Parameters[0], []byte("100")) {
		t.Fatalf("parameters: %q", b.Parameters)
	}
}
