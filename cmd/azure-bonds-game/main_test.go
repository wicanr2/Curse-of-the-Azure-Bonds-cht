package main

import (
	"reflect"
	"testing"
)

func TestWrapTextLinesUsesUnicodeRunesAndLineLimit(t *testing.T) {
	got := wrapTextLines("熔岩池中有火蜥蜴\n第二段", 4, 3)
	want := []string{"熔岩池中", "有火蜥蜴", "第二段"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wrapTextLines()=%q, want %q", got, want)
	}
}

func TestWrapTextLinesRejectsInvalidLayout(t *testing.T) {
	if got := wrapTextLines("中文", 0, 2); got != nil {
		t.Fatalf("zero-width wrap=%q, want nil", got)
	}
	if got := wrapTextLines("中文", 2, 0); got != nil {
		t.Fatalf("zero-line wrap=%q, want nil", got)
	}
}
