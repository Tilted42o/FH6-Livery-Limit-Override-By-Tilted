//go:build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var user32 = syscall.NewLazyDLL("user32.dll")
var dialog = syscall.NewLazyDLL("comdlg32.dll")
var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var shell32 = syscall.NewLazyDLL("shell32.dll")

func w(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func message(s string, flags uintptr) uintptr {
	r, _, _ := user32.NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(w(s))), uintptr(unsafe.Pointer(w("Layer Limit Override v2"))), flags)
	return r
}

type openFileName struct {
	Size            uint32
	Owner           uintptr
	Instance        uintptr
	Filter          *uint16
	CustomFilter    *uint16
	MaxCustomFilter uint32
	FilterIndex     uint32
	File            *uint16
	MaxFile         uint32
	FileTitle       *uint16
	MaxFileTitle    uint32
	InitialDir      *uint16
	Title           *uint16
	Flags           uint32
	FileOffset      uint16
	FileExtension   uint16
	DefaultExt      *uint16
	CustomData      uintptr
	Hook            uintptr
	TemplateName    *uint16
	Reserved        uintptr
	Reserved2       uint32
	FlagsEx         uint32
}

func choose() (string, error) {
	buf := make([]uint16, 32768)
	filter := syscall.StringToUTF16("Game executable (*.exe)")
	filter = append(filter, syscall.StringToUTF16("*.exe")...)
	filter = append(filter, 0)
	o := openFileName{Filter: &filter[0], File: &buf[0], MaxFile: uint32(len(buf)), Title: w("Select your installed Forza Horizon 6 executable"), Flags: 0x1000 | 0x800 | 0x80000 | 0x8}
	o.Size = uint32(unsafe.Sizeof(o))
	ret, _, _ := dialog.NewProc("GetOpenFileNameW").Call(uintptr(unsafe.Pointer(&o)))
	if ret == 0 {
		code, _, _ := dialog.NewProc("CommDlgExtendedError").Call()
		if code != 0 {
			return "", syscall.Errno(code)
		}
		return "", nil
	}
	return syscall.UTF16ToString(buf), nil
}

func openURL(url string) error {
	r, _, e := shell32.NewProc("ShellExecuteW").Call(
		0,
		uintptr(unsafe.Pointer(w("open"))),
		uintptr(unsafe.Pointer(w(url))),
		0, 0, 1,
	)
	if r <= 32 {
		if e != nil && e != syscall.Errno(0) {
			return e
		}
		return fmt.Errorf("Windows could not open the DVS project page (ShellExecute code %d)", r)
	}
	return nil
}

func showDVSPrerequisite(buildName, sha string) {
	msg := fmt.Sprintf(`MISSING OFFLINE PATCH

Your FH6 executable needs the offline patch before Layer Limit Override can be applied.

1. Close Forza Horizon 6.
2. Download and apply the DVS FH6 Offline Patcher from:

%s

3. After the offline patch has been applied properly, run Layer Limit Override again and select your patched forzahorizon6.exe.

Shoutout and credit to DVS for the FH6 Offline Patcher.

Open the DVS download page now?`, dvsProjectURL)
	if message(msg, 0x24) == 6 {
		if err := openURL(dvsProjectURL); err != nil {
			message(err.Error()+"\n\nCopy and paste this address into your browser:\n"+dvsProjectURL, 0x10)
		}
	}
}

func replaceFile(from, to string) error {
	ok, _, e := kernel32.NewProc("MoveFileExW").Call(uintptr(unsafe.Pointer(w(from))), uintptr(unsafe.Pointer(w(to))), 0x1|0x8)
	if ok == 0 {
		return e
	}
	return nil
}

func main() {
	var path string
	var err error
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		path, err = choose()
	}
	if err != nil {
		message(err.Error(), 0x10)
		return
	}
	if path == "" {
		return
	}

	b, err := os.ReadFile(path)
	if err != nil {
		message(err.Error(), 0x10)
		return
	}

	sha := digest(b)
	if buildName, ok := protectedBuildName(sha); ok {
		showDVSPrerequisite(buildName, sha)
		return
	}

	extra, warning := loadExternalProfiles()
	if warning != "" {
		message(warning+"\n\nBuilt-in and semantic detection will still be available.", 0x30)
	}

	r, err := detect(b, extra)
	if err != nil {
		message(err.Error(), 0x10)
		if message("Create a small compatibility report for this build?\n\nIt contains hashes, PE metadata, detector counts and a few tiny candidate code windows — never your whole game executable.", 0x24) == 6 {
			report, reportErr := createCompatibilityReport(path, r.Diag)
			if reportErr != nil {
				message("Could not create the compatibility report:\n\n"+reportErr.Error(), 0x10)
			} else {
				message("Compatibility report created:\n\n"+report+"\n\nSend this file with your FH6 version/storefront.", 0x40)
			}
		}
		return
	}

	backupName := buildBackupPath(path, r.Profile)
	prompt := fmt.Sprintf("Close Forza before continuing.\n\nApply 30,000 layers to Left / Right / Top and 10,000 to the remaining supported surfaces?\n\nDetection: %s\nBuild: %s\n\nA verified build-specific backup will be kept at:\n%s\n\nLayer Limit Override v%s - By Tilted42o", r.Detector, shortHash(r.Profile.Original), backupName, detectorVersion)
	if r.Restore {
		prompt = fmt.Sprintf("Restore the original executable from its verified backup?\n\nClose Forza first.\n\nDetection: %s\nBuild: %s", r.Detector, shortHash(r.Profile.Original))
	}
	if message(prompt, 0x21) != 1 {
		return
	}

	if err = install(path, r.Profile, r.Restore); err != nil {
		message(err.Error(), 0x10)
		return
	}
	if r.Restore {
		message("Original executable restored and verified.\n\nThe build-specific backup was retained.", 0x40)
	} else {
		message("Layer Limit Override applied and verified.\n\nLaunch Forza to use the increased layer limits.\n\nRun this tool again and choose the same executable to restore the original.", 0x40)
	}
}
