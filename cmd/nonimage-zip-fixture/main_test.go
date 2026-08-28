package main

import "testing"

func TestImageMemberClassification(t *testing.T) {
	images := []string{"8X8D1.DAX", "BIGPIC6.DAX", "BODY2.DAX", "CHEAD.DAX", "CPIC4.DAX", "DUNGCOM.DAX", "PIC6.DAX", "SKY.DAX", "TITLE.DAX", "WALLDEF5.DAX"}
	for _, name := range images {
		if !imageMember.MatchString(name) {
			t.Errorf("%s was not classified as image", name)
		}
	}
	nonimages := []string{"ECL1.DAX", "GEO2.DAX", "ITEMS", "MON6CHA.DAX", "GAME.OVR", "START.EXE"}
	for _, name := range nonimages {
		if imageMember.MatchString(name) {
			t.Errorf("%s was incorrectly classified as image", name)
		}
	}
}
