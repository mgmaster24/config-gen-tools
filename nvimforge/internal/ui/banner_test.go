package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintBanner_Shown(t *testing.T) {
	var buf bytes.Buffer
	PrintBanner(&buf, true)
	if buf.Len() == 0 {
		t.Error("expected banner output when show=true")
	}
	if !strings.Contains(buf.String(), "█") {
		t.Error("expected banner output to contain block-drawing characters")
	}
}

func TestPrintBanner_Hidden(t *testing.T) {
	var buf bytes.Buffer
	PrintBanner(&buf, false)
	if buf.Len() != 0 {
		t.Errorf("expected no output when show=false, got %q", buf.String())
	}
}
