package main

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultPartySavePathUsesUserConfigDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("AppData", t.TempDir())
	} else {
		root := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", root)
		t.Setenv("HOME", root)
	}
	wantName := "party.json"
	got := defaultPartySavePath()
	if filepath.Base(got) != wantName || filepath.Dir(got) == "." {
		t.Fatalf("defaultPartySavePath()=%q, want a user-config path ending in %s", got, wantName)
	}
}
