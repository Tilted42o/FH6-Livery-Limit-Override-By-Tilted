package main

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"sort"
)

const detectorVersion = "2.0.0-beta4"
const semanticWindow = 512

const dvsProjectURL = "https://github.com/DVS-code/FH6-Offline-Patcher/releases/tag/V1"

// Known untouched/protected retail executables. These builds do not expose the
// layer-limit initializer in patchable machine code. The DVS-prepared form of
// the same build does, so the Windows UI can give a specific prerequisite
// message instead of the generic zero-candidate error.
var knownProtectedBuilds = map[string]string{
	"fec4a63cdead26f6528564f0e33c3d7fd02cd887337ed6e1aeaca79eb2843fcd": "FH6 retail build (2026-09-21 executable)",
}

func protectedBuildName(sha string) (string, bool) {
	name, ok := knownProtectedBuilds[sha]
	return name, ok
}

type namedProfile struct {
	Name    string
	Profile profile
	Source  string
}

type semanticWrite struct {
	StoreAt    int
	ImmAt      int
	ImmLen     int
	BaseReg    int
	Disp       int
	Width      int
	Written    []byte
	SourceImm  []byte
	SourceKind string
}

type regImm struct {
	At     int
	ImmAt  int
	ImmLen int
	Reg    int
	Value  []byte
}

type regStore struct {
	At      int
	BaseReg int
	SrcReg  int
	Disp    int
	Width   int
}

type diagnostics struct {
	SHA256             string              `json:"sha256"`
	FileBytes          int                 `json:"file_bytes"`
	PEValid            bool                `json:"pe_valid"`
	Machine            uint16              `json:"machine,omitempty"`
	Timestamp          uint32              `json:"pe_timestamp,omitempty"`
	SizeOfImage        uint32              `json:"size_of_image,omitempty"`
	ExecutableSections []diagnosticSection `json:"executable_sections,omitempty"`
	ExactProfiles      int                 `json:"exact_profiles_loaded"`
	ExactMatch         string              `json:"exact_match,omitempty"`
	SemanticOriginal   int                 `json:"semantic_original_candidates"`
	SemanticPatched    int                 `json:"semantic_patched_candidates"`
	Detector           string              `json:"detector,omitempty"`
	AnchorOccurrences  int                 `json:"anchor_occurrences"`
	SupportWindows     []diagnosticWindow  `json:"support_windows,omitempty"`
	Reason             string              `json:"reason,omitempty"`
}

type diagnosticWindow struct {
	Section    string `json:"section"`
	FileOffset int    `json:"file_offset"`
	BytesHex   string `json:"bytes_hex"`
}

type diagnosticSection struct {
	Name       string `json:"name"`
	FileOffset uint32 `json:"file_offset"`
	Size       uint32 `json:"size"`
}

type semanticCandidate struct {
	edits        []edit
	state        int
	base         int
	minAt, maxAt int
}

type detectResult struct {
	Profile  profile
	Restore  bool
	Detector string
	Diag     diagnostics
}

var builtinProfiles = []namedProfile{
	{Name: "Reference FH6 build (2026-09-25)", Profile: referenceProfile, Source: "built-in"},
}

func detect(b []byte, extra []namedProfile) (detectResult, error) {
	d := diagnostics{SHA256: digest(b), FileBytes: len(b), ExactProfiles: len(builtinProfiles) + len(extra)}
	all := append([]namedProfile{}, builtinProfiles...)
	all = append(all, extra...)

	for _, np := range all {
		if d.SHA256 == np.Profile.Original {
			d.ExactMatch = np.Name
			d.Detector = "exact profile"
			return detectResult{Profile: np.Profile, Restore: false, Detector: "Exact profile: " + np.Name, Diag: d}, nil
		}
		if d.SHA256 == np.Profile.Patched {
			d.ExactMatch = np.Name
			d.Detector = "exact profile"
			return detectResult{Profile: np.Profile, Restore: true, Detector: "Exact profile: " + np.Name, Diag: d}, nil
		}
	}

	f, err := pe.NewFile(bytes.NewReader(b))
	if err != nil {
		d.Reason = "invalid Windows PE file"
		return detectResult{Diag: d}, compatibilityError(d)
	}
	defer f.Close()
	d.PEValid = true
	d.Machine = f.Machine
	d.Timestamp = f.TimeDateStamp
	if oh, ok := f.OptionalHeader.(*pe.OptionalHeader64); ok {
		d.SizeOfImage = oh.SizeOfImage
	} else {
		d.Reason = "missing PE32+ header"
		return detectResult{Diag: d}, compatibilityError(d)
	}
	if f.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
		d.Reason = "expected Windows x64"
		return detectResult{Diag: d}, compatibilityError(d)
	}

	var candidates []semanticCandidate
	originalTarget := targetBytes(false)
	patchedTarget := targetBytes(true)

	for _, s := range f.Sections {
		if s.Characteristics&0x20000000 == 0 {
			continue
		}
		start, end := uint64(s.Offset), uint64(s.Offset)+uint64(s.Size)
		if end > uint64(len(b)) {
			d.Reason = "truncated executable section"
			return detectResult{Diag: d}, compatibilityError(d)
		}
		d.ExecutableSections = append(d.ExecutableSections, diagnosticSection{Name: s.Name, FileOffset: s.Offset, Size: s.Size})
		section := b[int(start):int(end)]
		needle := []byte{0x12, 0x01, 0x00, 0x00}
		for pos := 0; pos < len(section); {
			n := bytes.Index(section[pos:], needle)
			if n < 0 {
				break
			}
			at := pos + n
			d.AnchorOccurrences++
			if len(d.SupportWindows) < 8 {
				supportLo := at - 40
				if supportLo < 0 {
					supportLo = 0
				}
				supportHi := at + 56
				if supportHi > len(section) {
					supportHi = len(section)
				}
				snip := section[supportLo:supportHi]
				interesting := bytes.Contains(snip, []byte{0xe8, 0x03}) || bytes.Contains(snip, []byte{0xb8, 0x0b}) || bytes.Contains(snip, []byte{0x10, 0x27}) || bytes.Contains(snip, []byte{0x30, 0x75})
				if interesting {
					d.SupportWindows = append(d.SupportWindows, diagnosticWindow{Section: s.Name, FileOffset: int(start) + supportLo, BytesHex: fmt.Sprintf("%x", snip)})
				}
			}
			lo := at - semanticWindow - 96
			if lo < 0 {
				lo = 0
			}
			hi := at + semanticWindow + 96
			if hi > len(section) {
				hi = len(section)
			}
			writes := scanSemanticWrites(section[lo:hi], int(start)+lo)
			for state, target := range [][]byte{originalTarget, patchedTarget} {
				cs := semanticCandidates(writes, target, state == 1)
				candidates = append(candidates, cs...)
			}
			pos = at + len(needle)
		}
	}

	candidates = dedupeCandidates(candidates)
	for _, c := range candidates {
		if c.state == 0 {
			d.SemanticOriginal++
		} else {
			d.SemanticPatched++
		}
	}

	if len(candidates) != 1 {
		d.Reason = fmt.Sprintf("found %d validated semantic layer-limit candidates; expected exactly one", len(candidates))
		return detectResult{Diag: d}, compatibilityError(d)
	}

	c := candidates[0]
	original := bytes.Clone(b)
	modified := bytes.Clone(b)
	for _, e := range c.edits {
		if e.Offset < 0 || e.Offset > len(b)-len(e.Before) || len(e.Before) != len(e.After) {
			d.Reason = "semantic patch plan contains an invalid edit"
			return detectResult{Diag: d}, compatibilityError(d)
		}
		copy(original[e.Offset:e.Offset+len(e.Before)], e.Before)
		copy(modified[e.Offset:e.Offset+len(e.After)], e.After)
	}
	p := profile{Original: digest(original), Patched: digest(modified), Edits: c.edits}

	if c.state == 0 && d.SHA256 != p.Original {
		d.Reason = "semantic original reconstruction did not match the selected file"
		return detectResult{Diag: d}, compatibilityError(d)
	}
	if c.state == 1 && d.SHA256 != p.Patched {
		d.Reason = "semantic patched reconstruction did not match the selected file"
		return detectResult{Diag: d}, compatibilityError(d)
	}

	d.Detector = "semantic layer table"
	return detectResult{Profile: p, Restore: c.state == 1, Detector: "Semantic layer-table detector", Diag: d}, nil
}

func compatibilityError(d diagnostics) error {
	return fmt.Errorf("Compatible layer-limit code could not be verified: %s. No file was changed.\n\nLayer Limit Override v%s\nFile SHA-256: %s\n\nA compatibility report can be created for this build.", d.Reason, detectorVersion, d.SHA256)
}

func targetBytes(patched bool) []byte {
	vals := []uint16{1000, 1000, 3000, 3000, 3000, 1000, 1000, 1000, 1000, 1000, 1000}
	if patched {
		vals = []uint16{10000, 10000, 30000, 30000, 30000, 10000, 10000, 10000, 10000, 10000, 10000}
	}
	out := make([]byte, len(vals)*2)
	for i, v := range vals {
		binary.LittleEndian.PutUint16(out[i*2:], v)
	}
	return out
}

func parseMemOperand(code []byte, pos int, rex byte) (base, disp, next int, ok bool) {
	if pos >= len(code) {
		return 0, 0, pos, false
	}
	modrm := code[pos]
	mod := int(modrm >> 6)
	rm := int(modrm & 7)
	if mod == 3 {
		return 0, 0, pos, false
	}
	pos++
	rexB := int(rex & 1)
	if rm == 4 {
		if pos >= len(code) {
			return 0, 0, pos, false
		}
		sib := code[pos]
		pos++
		index := int((sib >> 3) & 7)
		if index != 4 {
			return 0, 0, pos, false
		}
		rm = int(sib & 7)
	}
	base = rm + rexB*8
	switch mod {
	case 0:
		if rm == 5 {
			return 0, 0, pos, false
		}
		return base, 0, pos, true
	case 1:
		if pos >= len(code) {
			return 0, 0, pos, false
		}
		disp = int(int8(code[pos]))
		pos++
		return base, disp, pos, true
	case 2:
		if pos+4 > len(code) {
			return 0, 0, pos, false
		}
		disp = int(int32(binary.LittleEndian.Uint32(code[pos:])))
		pos += 4
		return base, disp, pos, true
	}
	return 0, 0, pos, false
}

func scanSemanticWrites(code []byte, fileBase int) []semanticWrite {
	var direct []semanticWrite
	var imms []regImm
	var stores []regStore

	for i := 0; i < len(code); i++ {
		j := i
		operand16 := false
		if j < len(code) && code[j] == 0x66 {
			operand16 = true
			j++
		}
		var rex byte
		if j < len(code) && code[j] >= 0x40 && code[j] <= 0x4f {
			rex = code[j]
			j++
		}
		if j >= len(code) {
			continue
		}
		op := code[j]

		if op >= 0xb8 && op <= 0xbf && rex&8 == 0 {
			width := 4
			if operand16 {
				width = 2
			}
			if j+1+width <= len(code) {
				reg := int(op-0xb8) + int(rex&1)*8
				v := append([]byte(nil), code[j+1:j+1+width]...)
				imms = append(imms, regImm{At: fileBase + i, ImmAt: fileBase + j + 1, ImmLen: width, Reg: reg, Value: v})
				i = j + 1 + width - 1
			}
			continue
		}

		if op == 0xc7 && rex&8 == 0 {
			if j+1 >= len(code) {
				continue
			}
			modrm := code[j+1]
			if (modrm>>3)&7 != 0 {
				continue
			}
			base, disp, n, ok := parseMemOperand(code, j+1, rex)
			if !ok {
				continue
			}
			width := 4
			if operand16 {
				width = 2
			}
			if n+width > len(code) {
				continue
			}
			imm := append([]byte(nil), code[n:n+width]...)
			direct = append(direct, semanticWrite{StoreAt: fileBase + i, ImmAt: fileBase + n, ImmLen: width, BaseReg: base, Disp: disp, Width: width, Written: imm, SourceImm: imm, SourceKind: "memory immediate"})
			i = n + width - 1
			continue
		}

		if op == 0x89 && rex&8 == 0 {
			if j+1 >= len(code) {
				continue
			}
			modrm := code[j+1]
			src := int((modrm>>3)&7) + int((rex>>2)&1)*8
			base, disp, next, ok := parseMemOperand(code, j+1, rex)
			if !ok {
				continue
			}
			width := 4
			if operand16 {
				width = 2
			}
			stores = append(stores, regStore{At: fileBase + i, BaseReg: base, SrcReg: src, Disp: disp, Width: width})
			i = next - 1
		}
	}

	out := append([]semanticWrite{}, direct...)
	for _, st := range stores {
		var best *regImm
		for k := range imms {
			ri := &imms[k]
			if ri.Reg != st.SrcReg || ri.At >= st.At || st.At-ri.At > 64 {
				continue
			}
			if ri.ImmLen < st.Width {
				continue
			}
			if best == nil || ri.At > best.At {
				best = ri
			}
		}
		if best == nil {
			continue
		}
		written := append([]byte(nil), best.Value[:st.Width]...)
		out = append(out, semanticWrite{StoreAt: st.At, ImmAt: best.ImmAt, ImmLen: best.ImmLen, BaseReg: st.BaseReg, Disp: st.Disp, Width: st.Width, Written: written, SourceImm: append([]byte(nil), best.Value...), SourceKind: "register immediate"})
	}
	return out
}

func desiredSourceAfter(w semanticWrite, patched bool) ([]byte, bool) {
	if len(w.SourceImm) == 2 {
		v := binary.LittleEndian.Uint16(w.SourceImm)
		var nv uint16
		switch v {
		case 1000:
			nv = 10000
		case 3000:
			nv = 30000
		case 10000:
			nv = 10000
		case 30000:
			nv = 30000
		default:
			return nil, false
		}
		if !patched {
			if v == 10000 {
				nv = 1000
			}
			if v == 30000 {
				nv = 3000
			}
		}
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, nv)
		return b, true
	}
	if len(w.SourceImm) == 4 {
		v := binary.LittleEndian.Uint32(w.SourceImm)
		var nv uint32
		switch v {
		case 1000:
			nv = 10000
		case 3000:
			nv = 30000
		case 0x03e803e8:
			nv = 0x27102710
		case 0x0bb80bb8:
			nv = 0x75307530
		case 10000:
			nv = 10000
		case 30000:
			nv = 30000
		case 0x27102710:
			nv = 0x27102710
		case 0x75307530:
			nv = 0x75307530
		default:
			return nil, false
		}
		if !patched {
			switch v {
			case 10000:
				nv = 1000
			case 30000:
				nv = 3000
			case 0x27102710:
				nv = 0x03e803e8
			case 0x75307530:
				nv = 0x0bb80bb8
			}
		}
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, nv)
		return b, true
	}
	return nil, false
}

func semanticCandidates(writes []semanticWrite, target []byte, currentlyPatched bool) []semanticCandidate {
	chunks := []struct{ disp, width int }{{0x112, 4}, {0x116, 2}, {0x118, 4}, {0x11c, 4}, {0x120, 4}, {0x124, 4}}
	var anchors []semanticWrite
	for _, w := range writes {
		if w.Disp == 0x112 && w.Width == 4 && bytes.Equal(w.Written, target[:4]) {
			anchors = append(anchors, w)
		}
	}
	var result []semanticCandidate
	for _, a := range anchors {
		selected := make([]semanticWrite, 0, len(chunks))
		ok := true
		minAt, maxAt := a.StoreAt, a.StoreAt
		for _, ch := range chunks {
			off := ch.disp - 0x112
			want := target[off : off+ch.width]
			var matches []semanticWrite
			for _, w := range writes {
				if w.BaseReg != a.BaseReg || w.Disp != ch.disp || w.Width != ch.width {
					continue
				}
				if abs(w.StoreAt-a.StoreAt) > semanticWindow {
					continue
				}
				if !bytes.Equal(w.Written, want) {
					continue
				}
				matches = append(matches, w)
			}
			if len(matches) != 1 {
				ok = false
				break
			}
			w := matches[0]
			selected = append(selected, w)
			if w.StoreAt < minAt {
				minAt = w.StoreAt
			}
			if w.StoreAt > maxAt {
				maxAt = w.StoreAt
			}
		}
		if !ok || maxAt-minAt > semanticWindow {
			continue
		}

		editMap := map[int]edit{}
		for _, w := range selected {
			var before, after []byte
			if currentlyPatched {
				after = append([]byte(nil), w.SourceImm...)
				b, good := desiredSourceAfter(w, false)
				if !good {
					ok = false
					break
				}
				before = b
			} else {
				before = append([]byte(nil), w.SourceImm...)
				a2, good := desiredSourceAfter(w, true)
				if !good {
					ok = false
					break
				}
				after = a2
			}
			if len(before) != len(after) {
				ok = false
				break
			}
			e := edit{Offset: w.ImmAt, Before: before, After: after}
			if old, exists := editMap[e.Offset]; exists {
				if !bytes.Equal(old.Before, e.Before) || !bytes.Equal(old.After, e.After) {
					ok = false
					break
				}
			} else {
				editMap[e.Offset] = e
			}
		}
		if !ok {
			continue
		}
		edits := make([]edit, 0, len(editMap))
		for _, e := range editMap {
			edits = append(edits, e)
		}
		sort.Slice(edits, func(i, j int) bool { return edits[i].Offset < edits[j].Offset })
		for i := 1; i < len(edits); i++ {
			if edits[i-1].Offset+len(edits[i-1].Before) > edits[i].Offset {
				ok = false
				break
			}
		}
		if !ok || len(edits) == 0 {
			continue
		}
		state := 0
		if currentlyPatched {
			state = 1
		}
		result = append(result, semanticCandidate{edits: edits, state: state, base: a.BaseReg, minAt: minAt, maxAt: maxAt})
	}
	return result
}

func dedupeCandidates(in []semanticCandidate) []semanticCandidate {
	seen := map[string]bool{}
	var out []semanticCandidate
	for _, c := range in {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%d|", c.state)
		for _, e := range c.edits {
			fmt.Fprintf(&buf, "%d:%x>%x;", e.Offset, e.Before, e.After)
		}
		k := buf.String()
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, c)
	}
	return out
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
