package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const profilesFilename = "LayerLimitOverride.profiles.json"

type profileFile struct {
	Schema   int           `json:"schema"`
	Profiles []jsonProfile `json:"profiles"`
}

type jsonProfile struct {
	Name           string      `json:"name"`
	OriginalSHA256 string      `json:"original_sha256"`
	PatchedSHA256  string      `json:"patched_sha256"`
	FileBytes      int         `json:"file_bytes,omitempty"`
	Patches        []jsonPatch `json:"patches"`
}

type jsonPatch struct {
	Offset int    `json:"offset"`
	Before string `json:"before"`
	After  string `json:"after"`
}

func validateSHA(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func convertJSONProfile(j jsonProfile, source string) (namedProfile, error) {
	if strings.TrimSpace(j.Name) == "" {
		return namedProfile{}, errors.New("profile name is empty")
	}
	j.OriginalSHA256 = strings.ToLower(strings.TrimSpace(j.OriginalSHA256))
	j.PatchedSHA256 = strings.ToLower(strings.TrimSpace(j.PatchedSHA256))
	if !validateSHA(j.OriginalSHA256) || !validateSHA(j.PatchedSHA256) {
		return namedProfile{}, errors.New("invalid SHA-256")
	}
	if j.OriginalSHA256 == j.PatchedSHA256 {
		return namedProfile{}, errors.New("original and patched hashes are identical")
	}
	if len(j.Patches) == 0 {
		return namedProfile{}, errors.New("profile has no patches")
	}

	p := profile{Original: j.OriginalSHA256, Patched: j.PatchedSHA256}
	for _, jp := range j.Patches {
		before, err := hex.DecodeString(jp.Before)
		if err != nil {
			return namedProfile{}, fmt.Errorf("bad before bytes at offset %d", jp.Offset)
		}
		after, err := hex.DecodeString(jp.After)
		if err != nil {
			return namedProfile{}, fmt.Errorf("bad after bytes at offset %d", jp.Offset)
		}
		if jp.Offset < 0 || len(before) == 0 || len(before) != len(after) {
			return namedProfile{}, fmt.Errorf("invalid patch at offset %d", jp.Offset)
		}
		p.Edits = append(p.Edits, edit{Offset: jp.Offset, Before: before, After: after})
	}
	sort.Slice(p.Edits, func(i, j int) bool { return p.Edits[i].Offset < p.Edits[j].Offset })
	for i := 1; i < len(p.Edits); i++ {
		if p.Edits[i-1].Offset+len(p.Edits[i-1].Before) > p.Edits[i].Offset {
			return namedProfile{}, errors.New("overlapping patches")
		}
	}
	return namedProfile{Name: j.Name, Profile: p, Source: source}, nil
}

func profileSearchPath() string {
	exe, err := os.Executable()
	if err != nil {
		return profilesFilename
	}
	return filepath.Join(filepath.Dir(exe), profilesFilename)
}

func loadExternalProfiles() ([]namedProfile, string) {
	path := profileSearchPath()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ""
		}
		return nil, fmt.Sprintf("Could not read %s: %v", path, err)
	}
	var pf profileFile
	if err = json.Unmarshal(b, &pf); err != nil {
		return nil, fmt.Sprintf("Compatibility profile file is invalid and was ignored: %v", err)
	}
	if pf.Schema != 1 {
		return nil, fmt.Sprintf("Compatibility profile file uses unsupported schema %d and was ignored.", pf.Schema)
	}
	var out []namedProfile
	for i, j := range pf.Profiles {
		np, err := convertJSONProfile(j, "external")
		if err != nil {
			return nil, fmt.Sprintf("Compatibility profile #%d is invalid and the profile file was ignored: %v", i+1, err)
		}
		out = append(out, np)
	}
	return out, ""
}
