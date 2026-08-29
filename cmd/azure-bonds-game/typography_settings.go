package main

import "golang.org/x/image/font"

// styledFace carries the player's independent weight choice through the
// existing text-rendering seam without duplicating every call site.
type styledFace struct {
	font.Face
	bold bool
}

func faceIsBold(face font.Face) bool {
	if styled, ok := face.(styledFace); ok {
		return styled.bold
	}
	return true // Legacy/test faces preserve the remake's established weight.
}

func boolSetting(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func (a *app) applyTypographySettings() {
	if len(a.localeFontPaths) != 0 {
		reading := make(map[string]font.Face, len(a.localeFontPaths))
		ui := make(map[string]font.Face, len(a.localeFontPaths))
		safeUI := make(map[string]font.Face, len(a.localeFontPaths))
		for language, path := range a.localeFontPaths {
			reading[language] = loadFace(path, float64(a.ui.settings.ReadingTextPX))
			ui[language] = loadFace(path, float64(a.ui.settings.InterfaceTextPX))
			safeSize := a.ui.settings.InterfaceTextPX
			if safeSize > 20 {
				safeSize = 20
			}
			safeUI[language] = loadFace(path, float64(safeSize))
		}
		a.localeFaces = reading
		a.localeCompactFaces = ui
		a.localeSafeCompactFaces = safeUI
	}
	language := a.ui.settings.Language
	reading := a.localeFaces[language]
	interfaceFace := a.localeCompactFaces[language]
	if reading == nil {
		reading = a.face
	}
	if interfaceFace == nil {
		interfaceFace = a.compactFace
	}
	a.face = styledFace{Face: reading, bold: boolSetting(a.ui.settings.ReadingBold, true)}
	a.compactFace = styledFace{Face: interfaceFace, bold: boolSetting(a.ui.settings.InterfaceBold, true)}
	safe := a.localeSafeCompactFaces[language]
	if safe == nil {
		safe = interfaceFace
	}
	a.safeCompactFace = styledFace{Face: safe, bold: boolSetting(a.ui.settings.InterfaceBold, true)}
}

func (a *app) selectLocaleTypography(language string) {
	if face := a.localeFaces[language]; face != nil {
		a.face = styledFace{Face: face, bold: boolSetting(a.ui.settings.ReadingBold, true)}
	}
	if face := a.localeCompactFaces[language]; face != nil {
		a.compactFace = styledFace{Face: face, bold: boolSetting(a.ui.settings.InterfaceBold, true)}
	}
	if face := a.localeSafeCompactFaces[language]; face != nil {
		a.safeCompactFace = styledFace{Face: face, bold: boolSetting(a.ui.settings.InterfaceBold, true)}
	}
}
