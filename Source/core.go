package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type edit struct {
	Offset        int
	Before, After []byte
}

type profile struct {
	Original, Patched string
	Edits             []edit
}

func unhex(s string) []byte {
	b, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return b
}

func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

var referenceProfile = profile{
	Original: "1bd6e4260c71d394b8ce42bac188c200eaeb8a88a6c8e1d2dca0aa3f6965661d",
	Patched:  "b7a9230e649d1658120876eafa1372000a81bc86c2957e87d65229585a77cdaa",
	Edits: []edit{
		{10207841, unhex("e803e803"), unhex("10271027")},
		{10207846, unhex("b80b0000"), unhex("30750000")},
		{10207856, unhex("b80bb80b"), unhex("30753075")},
		{10207873, unhex("e803e803"), unhex("10271027")},
		{10207883, unhex("e803e803"), unhex("10271027")},
		{10207893, unhex("e803e803"), unhex("10271027")},
	},
}

func transform(b []byte, p profile) ([]byte, error) {
	if digest(b) != p.Original {
		return nil, errors.New("This executable changed after detection or does not match the selected compatibility profile. No file was changed.")
	}
	out := bytes.Clone(b)
	for _, e := range p.Edits {
		if e.Offset < 0 || e.Offset > len(out)-len(e.Before) || len(e.Before) != len(e.After) || !bytes.Equal(out[e.Offset:e.Offset+len(e.Before)], e.Before) {
			return nil, errors.New("Patch bytes do not match the validated plan; stopped without changing the game file.")
		}
		copy(out[e.Offset:], e.After)
	}
	if digest(out) != p.Patched {
		return nil, errors.New("Patched file verification failed before installation.")
	}
	return out, nil
}

func shortHash(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

func buildBackupPath(path string, p profile) string {
	return path + ".lu-layers-backup-" + shortHash(p.Original)
}

func legacyBackupPath(path string) string { return path + ".lu-layers-backup" }

func backup(path string, b []byte, p profile) error {
	name := buildBackupPath(path, p)
	f, e := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if os.IsExist(e) {
		saved, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		if digest(saved) != p.Original {
			return errors.New("An existing build-specific backup does not match. It will not be overwritten.")
		}
		return nil
	}
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return fmt.Errorf("Backup could not be completed; original is unchanged: %w", e)
	}
	saved, e := os.ReadFile(name)
	if e != nil {
		return e
	}
	if digest(saved) != p.Original {
		return errors.New("Backup verification failed; original is unchanged.")
	}
	return nil
}

func readVerifiedBackup(path string, p profile) ([]byte, string, error) {
	candidates := []string{buildBackupPath(path, p), legacyBackupPath(path)}
	for _, name := range candidates {
		b, err := os.ReadFile(name)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, "", err
		}
		if digest(b) == p.Original {
			return b, name, nil
		}
	}
	return nil, "", errors.New("No verified original backup for this exact game build was found.")
}

func install(path string, p profile, restore bool) error {
	info, e := os.Lstat(path)
	if e != nil {
		return e
	}
	if !info.Mode().IsRegular() {
		return errors.New("Select a regular executable file.")
	}
	current, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	want := p.Original
	if restore {
		want = p.Patched
	}
	if digest(current) != want {
		return errors.New("The executable changed after detection. No file was changed.")
	}

	var out []byte
	if restore {
		out, _, e = readVerifiedBackup(path, p)
		if e != nil {
			return e
		}
	} else {
		out, e = transform(current, p)
		if e != nil {
			return e
		}
		if e = backup(path, current, p); e != nil {
			return e
		}
	}

	f, e := os.CreateTemp(filepath.Dir(path), ".lu-layers-pending-*.exe")
	if e != nil {
		return e
	}
	temp := f.Name()
	defer os.Remove(temp)
	_, e = f.Write(out)
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	check, e := os.ReadFile(temp)
	if e != nil {
		return e
	}
	if digest(check) != digest(out) {
		return errors.New("Staged executable verification failed.")
	}
	check, e = os.ReadFile(path)
	if e != nil {
		return e
	}
	if digest(check) != want {
		return errors.New("The game file changed during preparation; stopped.")
	}
	if e = replaceFile(temp, path); e != nil {
		return fmt.Errorf("Could not replace the executable. Close Forza and check folder write access. Your backup is retained. %w", e)
	}
	check, e = os.ReadFile(path)
	if e != nil {
		return e
	}
	if digest(check) != digest(out) {
		return errors.New("Final verification failed. Restore the retained original backup before launching.")
	}
	return nil
}
