package datasets

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteSSV(t *testing.T) {
	s := "072, 09:36:07.000300000;30.2;120.678;-2.1;0.3;1;0\r\n"
	p := TimePositionTransfer{72, 9*3600 + 36*60 + 7.0003, 30.2, 120.678, -2.1, 0.3, 1, 0}
	var b bytes.Buffer
	err := p.WriteSSV(&b)
	if err != nil {
		t.Fatal(err)
	}
	got := b.String()
	if got != s {
		t.Fatalf("got %#v, want %#v", got, s)
	}
}

func TestReadSSV(t *testing.T) {
	s := "072, 09:36:07.000300000; 30.2000;120.6780;-2.100; 0.30; 1; 0\r\n"
	p := TimePositionTransfer{72, 9*3600 + 36*60 + 7.0003, 30.2, 120.678, -2.1, 0.3, 1, 0}
	r := strings.NewReader(s)
	var got TimePositionTransfer
	err := (&got).ReadSSV(r)
	if err != nil {
		t.Fatal(err)
	}
	if p != got {
		t.Fatalf("got %#v, want %#v", got, p)
	}
}
