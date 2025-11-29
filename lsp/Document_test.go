package main

import (
	"os"
	"strings"
	"testing"

	"github.com/tliron/commonlog"
)

var TEST_DOCUMENT = Document{
	uri: "",
	lines: []string{
		"Hello",
		"world",
		"Hello",
		"world",
		"Hello",
		"world",
		"Hello",
		"world",
	},
}
var FULL_LINES = strings.Join(TEST_DOCUMENT.lines, "\n")

func ConfigureLogger() {
	log_path := os.Stdout.Name()
	commonlog.Configure(1, &log_path)
	log = commonlog.GetLogger("Test")
	log.SetMaxLevel(commonlog.Debug)
}

func TestDocumentInsert(t *testing.T) {
	ConfigureLogger()
	num_lines := len(TEST_DOCUMENT.lines)
	start_line := strings.Clone(TEST_DOCUMENT.lines[0])
	end_line := strings.Clone(TEST_DOCUMENT.lines[1])
	TEST_DOCUMENT.editRange(0, 5, 1, 0, "\n")
	if len(TEST_DOCUMENT.lines) != num_lines+1 {
		log.Infof("Lines: %d", len(TEST_DOCUMENT.lines))
		t.Error("Number of lines mismatch!")
	}
	if TEST_DOCUMENT.lines[2] != end_line {
		t.Error("Existing line deleted on insertion!")
	}
	if start_line != TEST_DOCUMENT.lines[0] {
		t.Error("Start line broken!!")
	}
	if TEST_DOCUMENT.lines[1] != "" {
		t.Error("Unexpected value in the inserted line!")
	}
}

func TestDocumentInsertJ(t *testing.T) {
	ConfigureLogger()
	num_lines := len(TEST_DOCUMENT.lines)
	start_line := strings.Clone(TEST_DOCUMENT.lines[0])
	end_line := strings.Clone(TEST_DOCUMENT.lines[1])
	TEST_DOCUMENT.editRange(0, 5, 1, 0, "\n")
	TEST_DOCUMENT.editRange(1, 0, 1, 0, "j")
	if len(TEST_DOCUMENT.lines) != num_lines+1 {
		t.Error("Number of lines mismatch!")
	}
	if TEST_DOCUMENT.lines[2] != end_line {
		t.Error("Existing line deleted on insertion!")
	}
	if start_line != TEST_DOCUMENT.lines[0] {
		t.Error("Start line broken!!")
	}
	if TEST_DOCUMENT.lines[1] != "j" {
		t.Error("Unexpected value in the inserted line!")
	}
}

func TestDocumentInsertJJ(t *testing.T) {
	ConfigureLogger()
	num_lines := len(TEST_DOCUMENT.lines)
	start_line := strings.Clone(TEST_DOCUMENT.lines[0])
	end_line := strings.Clone(TEST_DOCUMENT.lines[1])
	TEST_DOCUMENT.editRange(0, 5, 1, 0, "\n")
	TEST_DOCUMENT.editRange(1, 0, 1, 0, "j")
	TEST_DOCUMENT.editRange(1, 1, 1, 1, "j")
	if len(TEST_DOCUMENT.lines) != num_lines+1 {
		t.Error("Number of lines mismatch!")
	}
	if TEST_DOCUMENT.lines[2] != end_line {
		t.Error("Existing line deleted on insertion!")
	}
	if start_line != TEST_DOCUMENT.lines[0] {
		t.Error("Start line broken!!")
	}
	if TEST_DOCUMENT.lines[1] != "jj" {
		t.Error("Unexpected value in the inserted line!")
	}
}

func TestDocumentRemoveJ(t *testing.T) {
	num_lines := len(TEST_DOCUMENT.lines)
	start_line := strings.Clone(TEST_DOCUMENT.lines[0])
	end_line := strings.Clone(TEST_DOCUMENT.lines[1])
	TEST_DOCUMENT.editRange(0, 5, 1, 0, "\n")
	TEST_DOCUMENT.editRange(1, 0, 1, 0, "j")
	TEST_DOCUMENT.editRange(1, 0, 1, 1, "")
	if len(TEST_DOCUMENT.lines) != num_lines+1 {
		t.Error("Number of lines mismatch!")
	}
	if TEST_DOCUMENT.lines[2] != end_line {
		t.Error("Existing line deleted on insertion!")
	}
	if start_line != TEST_DOCUMENT.lines[0] {
		t.Error("Start line broken!!")
	}
	if TEST_DOCUMENT.lines[1] != "" {
		t.Error("Unexpected value in the inserted line!")
	}
}
