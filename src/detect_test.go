package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func putU16(b []byte, o int, v uint16) { binary.LittleEndian.PutUint16(b[o:], v) }
func putU32(b []byte, o int, v uint32) { binary.LittleEndian.PutUint32(b[o:], v) }

func miniPE(code []byte) []byte {
	size := 0x400 + len(code) + 0x100
	b := make([]byte, size)
	copy(b, "MZ")
	putU32(b, 0x3c, 0x80)
	copy(b[0x80:], "PE\x00\x00")
	putU16(b, 0x84, 0x8664)
	putU16(b, 0x86, 1)
	putU16(b, 0x94, 240)
	putU16(b, 0x98, 0x20b)
	putU32(b, 0x104, 16)
	sh := 0x188
	copy(b[sh:], ".text")
	putU32(b, sh+8, uint32(len(code)))
	putU32(b, sh+12, 0x1000)
	putU32(b, sh+16, uint32(len(code)))
	putU32(b, sh+20, 0x400)
	putU32(b, sh+36, 0x60000020)
	copy(b[0x400:], code)
	return b
}

func emitC7(baseModRM byte, disp uint32, imm uint32) []byte {
	x := []byte{0xc7, baseModRM, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(x[2:], disp)
	binary.LittleEndian.PutUint32(x[6:], imm)
	return x
}
func emit66C7(baseModRM byte, disp uint32, imm uint16) []byte {
	x := []byte{0x66, 0xc7, baseModRM, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(x[3:], disp)
	binary.LittleEndian.PutUint16(x[7:], imm)
	return x
}
func emit6689(baseModRM byte, disp uint32) []byte {
	x := []byte{0x66, 0x89, baseModRM, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(x[3:], disp)
	return x
}

func semanticFixture(patched bool, directWord bool, reorder bool) []byte {
	one := uint32(0x03e803e8)
	three := uint32(0x0bb80bb8)
	single := uint32(3000)
	if patched {
		one = 0x27102710
		three = 0x75307530
		single = 30000
	}
	var chunks [][]byte
	chunks = append(chunks, emitC7(0x81, 0x112, one))
	if directWord {
		chunks = append(chunks, emit66C7(0x81, 0x116, uint16(single)))
	} else {
		m := []byte{0xb8, 0, 0, 0, 0}
		binary.LittleEndian.PutUint32(m[1:], single)
		chunks = append(chunks, m)
	}
	chunks = append(chunks, emitC7(0x81, 0x118, three))
	if !directWord {
		chunks = append(chunks, emit6689(0x81, 0x116))
	}
	chunks = append(chunks, emitC7(0x81, 0x11c, one), emitC7(0x81, 0x120, one), emitC7(0x81, 0x124, one))
	if reorder {
		if len(chunks) == 6 {
			chunks = [][]byte{chunks[3], chunks[0], chunks[5], chunks[1], chunks[4], chunks[2]}
		}
	}
	code := bytes.Repeat([]byte{0x90}, 80)
	for _, c := range chunks {
		code = append(code, c...)
		code = append(code, 0x90, 0x90, 0x90)
	}
	code = append(code, bytes.Repeat([]byte{0x90}, 80)...)
	return miniPE(code)
}

func TestSemanticVariants(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		patched, direct, reorder bool
	}{
		{"legacy-shaped", false, false, false},
		{"direct-word", false, true, false},
		{"reordered", false, true, true},
		{"patched", true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := semanticFixture(tc.patched, tc.direct, tc.reorder)
			r, err := detect(b, nil)
			if err != nil {
				t.Fatal(err)
			}
			if r.Detector != "Semantic layer-table detector" {
				t.Fatalf("wrong detector %q", r.Detector)
			}
			if r.Restore != tc.patched {
				t.Fatalf("restore=%v", r.Restore)
			}
			if !tc.patched {
				out, err := transform(b, r.Profile)
				if err != nil {
					t.Fatal(err)
				}
				rr, err := detect(out, nil)
				if err != nil {
					t.Fatal(err)
				}
				if !rr.Restore {
					t.Fatal("patched fixture not recognized for restore")
				}
			}
		})
	}
}

func TestSemanticRejectsAmbiguousAndPartial(t *testing.T) {
	a := semanticFixture(false, false, false)
	b := bytes.Clone(a)
	seq := a[0x400+80 : 0x400+80+90]
	b = append(b, seq...)
	putU32(b, 0x188+16, uint32(len(b)-0x400))
	putU32(b, 0x188+8, uint32(len(b)-0x400))
	if _, err := detect(b, nil); err == nil {
		t.Fatal("accepted ambiguous candidates")
	}

	p := semanticFixture(false, false, false)
	idx := bytes.Index(p, []byte{0xe8, 0x03, 0xe8, 0x03})
	if idx < 0 {
		t.Fatal("fixture")
	}
	p[idx] = 0xe9
	if _, err := detect(p, nil); err == nil {
		t.Fatal("accepted partial table")
	}
}

func TestActualReferenceAndSemanticVariation(t *testing.T) {
	path := os.Getenv("LU_TEST_EXECUTABLE")
	if path == "" {
		t.Skip("no real executable fixture")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r, err := detect(b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(r.Detector, "Exact profile:") {
		t.Fatalf("expected exact profile, got %q", r.Detector)
	}
	varied := append(bytes.Clone(b), []byte("unrelated signed/package variation")...)
	rr, err := detect(varied, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rr.Detector != "Semantic layer-table detector" {
		t.Fatalf("semantic fallback failed: %q", rr.Detector)
	}
	out, err := transform(varied, rr.Profile)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for i := range varied {
		if varied[i] != out[i] {
			changed++
		}
	}
	if changed != 22 {
		t.Fatalf("changed %d bytes, want 22", changed)
	}
	if !bytes.Equal(out[len(out)-len("unrelated signed/package variation"):], varied[len(varied)-len("unrelated signed/package variation"):]) {
		t.Fatal("unrelated bytes changed")
	}
}

func TestBuildSpecificBackupAndLegacyRestore(t *testing.T) {
	original := []byte("prefix-OLD-suffix")
	changed := []byte("prefix-NEW-suffix")
	p := profile{Original: digest(original), Patched: digest(changed), Edits: []edit{{7, []byte("OLD"), []byte("NEW")}}}
	dir := t.TempDir()
	path := filepath.Join(dir, "game.exe")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := install(path, p, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(buildBackupPath(path, p)); err != nil {
		t.Fatal("build backup missing", err)
	}
	if err := install(path, p, true); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, original) {
		t.Fatal("restore mismatch")
	}

	os.Remove(buildBackupPath(path, p))
	os.WriteFile(path, changed, 0600)
	os.WriteFile(legacyBackupPath(path), original, 0600)
	if err := install(path, p, true); err != nil {
		t.Fatal(err)
	}
}

func TestKnownProtectedBuildIdentification(t *testing.T) {
	const sha = "fec4a63cdead26f6528564f0e33c3d7fd02cd887337ed6e1aeaca79eb2843fcd"
	name, ok := protectedBuildName(sha)
	if !ok || name == "" {
		t.Fatal("known clean/protected build was not identified")
	}
	if _, ok := protectedBuildName(strings.Repeat("0", 64)); ok {
		t.Fatal("unknown hash was misidentified as a known protected build")
	}
}
