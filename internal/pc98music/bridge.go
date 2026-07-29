// Package pc98music verifies title-specific PC-98 music bridge evidence.
package pc98music

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	GameSHA256   = "8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0"
	DriverSHA256 = "bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5"

	DriverMissingStart = 0x4000
	DriverMissingEnd   = 0x4400
)

// Anchor is a short instruction sequence independently located by IDA Pro and
// then checked against the original executable bytes.
type Anchor struct {
	Binary     string `json:"binary"`
	Label      string `json:"label"`
	FileOffset int    `json:"file_offset"`
	Bytes      []byte `json:"-"`
}

type AnchorResult struct {
	Binary     string `json:"binary"`
	Label      string `json:"label"`
	FileOffset int    `json:"file_offset"`
	Size       int    `json:"size"`
}

// SoundBIOSService describes one INT D2h client that is present in this
// MSCDRV.EXE. Names follow NEC's official PC-9800 Sound BIOS command table.
type SoundBIOSService struct {
	Command    byte   `json:"command"`
	Name       string `json:"name"`
	FileOffset int    `json:"file_offset"`
	Arguments  string `json:"arguments"`
}

type BridgeReport struct {
	GameSHA256        string             `json:"game_sha256"`
	DriverSHA256      string             `json:"driver_sha256"`
	DriverMissingFrom int                `json:"driver_missing_from"`
	DriverMissingTo   int                `json:"driver_missing_to"`
	Anchors           []AnchorResult     `json:"anchors"`
	SoundBIOSServices []SoundBIOSService `json:"sound_bios_services"`
}

var gameAnchors = []Anchor{
	{
		Binary:     "GAME.EXE",
		Label:      "MSCPLAY_MSCSTOP_VECTOR_7E_CALLS",
		FileOffset: 0x9410,
		Bytes: mustHex(
			"c606397f008a4606a2387fb07e50bf387f1e579a0b00bd08" +
				"89ec5dca02005589e5c606397f01b07e50bf387f1e579a0b00bd08",
		),
	},
	{
		Binary:     "GAME.EXE",
		Label:      "REGISTER_IMAGE_IVT_TRAMPOLINE",
		FileOffset: 0x957B,
		Bytes: mustHex(
			"551e8bec9cbb45000e5333db8edb8a5e0cd1e3d1e3c51f1e53" +
				"c57608fcad50ad8bd8ad8bc8ad8bd0ad8be8ad50ad8bf8ad50ad" +
				"8ec01f5e58facb9c0657558becc47e10fcab8bc3ab8bc1ab8bc2" +
				"ab58ab8bc6ab58ab8cd8ab58ab58ab1f5dca",
		),
	},
}

var driverAnchors = []Anchor{
	{
		Binary:     "MSCDRV.EXE",
		Label:      "VECTOR_7E_HANDLER",
		FileOffset: 0x0280,
		Bytes: mustHex(
			"505351525657551e06bb2d018edb833e060000752d84e4752650" +
				"e82d10e84a10e8300ce8910ee8370e58bbff003c0e7503bb0100" +
				"5350e8650183c404eb0490e8bc02071f5d5f5e5a595b58cf",
		),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "INSTALL_VECTOR_7E",
		FileOffset: 0x02E3,
		Bytes:      mustHex("2ec41ef20026c7078000268c4f02c3f8010000"),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "INSTALL_INT_D2",
		FileOffset: 0x12CA,
		Bytes: mustHex(
			"b8d235cd21891e21038cc0a323031eb8e0ce8ed8bb06008b17" +
				"b8d225cd211fc3",
		),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "INT_D2_INITIALIZE_AND_SHUTDOWN_CLIENTS",
		FileOffset: 0x135D,
		Bytes:      mustHex("b400cdd2b82d018ed8c3558bec8a4604e802005dc3b402cdd2c3"),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "DIRECT_YM2203_WRITE_PORTS_188_18A",
		FileOffset: 0x1275,
		Bytes: mustHex(
			"515250ba8801eca88075f85850ee8ac34242b92000e2feee585a59c3",
		),
	},
	{
		Binary:     "MSCDRV.EXE",
		Label:      "DIRECT_YM2203_READ_PORTS_188_18A",
		FileOffset: 0x12AB,
		Bytes: mustHex(
			"515250ba8801eca88075f858eeb90400e2fe504242ba8a01ec8ad8585a59c3",
		),
	},
}

var soundBIOSServices = []SoundBIOSService{
	{0x00, "INITIALIZE", 0x135D, "ES=控制／工作區段"},
	{0x02, "CLEAR", 0x1372, "無"},
	{0x10, "READREG", 0x1384, "AL=OPN 暫存器；回傳 BX"},
	{0x11, "WRITEREG", 0x13A1, "AL=OPN 暫存器，BL=資料"},
	{0x12, "SETTOUCH", 0x13B4, "AL=聲道，BL=觸鍵／閘門比"},
	{0x13, "NOTE", 0x13CA, "AL=聲道，BH=音高，BL=音長"},
	{0x14, "SETLENGTH", 0x13DD, "AL=聲道，BL=預設音長"},
	{0x16, "SETPARABLOCK", 0x140D, "AL=聲道，ES:BX=參數塊，DL=種類"},
	{0x17, "READPARA", 0x1420, "AL=聲道，BL=參數編號；回傳 BX"},
	{0x18, "WRITEPARA", 0x1438, "AL=聲道，BL=參數編號，DX=值"},
	{0x19, "ALLSTOP", 0x143D, "無"},
	{0x1A, "CONTPLAY", 0x1442, "無"},
	{0x1B, "MODUON", 0x1462, "AL=聲道"},
	{0x1C, "MODUOFF", 0x1472, "AL=聲道"},
	{0x1D, "SETINTCOND", 0x148A, "AL=聲道，ES:BX=回呼，CX=條件"},
	{0x1E, "HOLDSTATE", 0x1452, "AL=聲道，BL=維持音長"},
	{0x1F, "SETVOLUME", 0x149D, "AL=聲道，BL=音量"},
}

func mustHex(value string) []byte {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return decoded
}

func fileSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func verifyAnchors(data []byte, anchors []Anchor) ([]AnchorResult, error) {
	results := make([]AnchorResult, 0, len(anchors))
	for _, anchor := range anchors {
		end := anchor.FileOffset + len(anchor.Bytes)
		if anchor.FileOffset < 0 || end > len(data) {
			return nil, fmt.Errorf(
				"%s %s range 0x%X..0x%X is outside %d-byte input",
				anchor.Binary, anchor.Label, anchor.FileOffset, end, len(data),
			)
		}
		if !bytes.Equal(data[anchor.FileOffset:end], anchor.Bytes) {
			return nil, fmt.Errorf(
				"%s %s byte mismatch at file offset 0x%X",
				anchor.Binary, anchor.Label, anchor.FileOffset,
			)
		}
		results = append(results, AnchorResult{
			Binary:     anchor.Binary,
			Label:      anchor.Label,
			FileOffset: anchor.FileOffset,
			Size:       len(anchor.Bytes),
		})
	}
	return results, nil
}

func verifySoundBIOSServices(data []byte) error {
	for _, service := range soundBIOSServices {
		want := []byte{0xB4, service.Command, 0xCD, 0xD2}
		end := service.FileOffset + len(want)
		if service.FileOffset < 0 || end > len(data) {
			return fmt.Errorf(
				"MSCDRV.EXE Sound BIOS %s range 0x%X..0x%X is outside input",
				service.Name, service.FileOffset, end,
			)
		}
		if !bytes.Equal(data[service.FileOffset:end], want) {
			return fmt.Errorf(
				"MSCDRV.EXE Sound BIOS %s byte mismatch at file offset 0x%X",
				service.Name, service.FileOffset,
			)
		}
	}
	return nil
}

// AuditBridge verifies the exact supplied GAME.EXE and incomplete MSCDRV.EXE.
// It deliberately rejects other dumps so evidence from the known missing
// sector cannot be silently generalized.
func AuditBridge(game, driver []byte) (BridgeReport, error) {
	gameHash := fileSHA256(game)
	if gameHash != GameSHA256 {
		return BridgeReport{}, fmt.Errorf("GAME.EXE SHA-256 %s, want %s", gameHash, GameSHA256)
	}
	driverHash := fileSHA256(driver)
	if driverHash != DriverSHA256 {
		return BridgeReport{}, fmt.Errorf("MSCDRV.EXE SHA-256 %s, want %s", driverHash, DriverSHA256)
	}
	gameResults, err := verifyAnchors(game, gameAnchors)
	if err != nil {
		return BridgeReport{}, err
	}
	driverResults, err := verifyAnchors(driver, driverAnchors)
	if err != nil {
		return BridgeReport{}, err
	}
	if err := verifySoundBIOSServices(driver); err != nil {
		return BridgeReport{}, err
	}
	return BridgeReport{
		GameSHA256:        gameHash,
		DriverSHA256:      driverHash,
		DriverMissingFrom: DriverMissingStart,
		DriverMissingTo:   DriverMissingEnd,
		Anchors:           append(gameResults, driverResults...),
		SoundBIOSServices: append([]SoundBIOSService(nil), soundBIOSServices...),
	}, nil
}
