<div align="center">

# FH6 Livery Limit Override

### By Tilted42o

**Break past Forza Horizon 6's stock livery layer limits.**

[![Version](https://img.shields.io/badge/version-v2.0.0--beta4-blue)](#downloads)
[![Windows](https://img.shields.io/badge/platform-Windows%2064--bit-lightgrey)](#installation)
[![Build](https://img.shields.io/github/actions/workflow/status/lilanimal42o/FH6-Livery-Limit-Override-By-Tilted/build-windows.yml?label=build)](../../actions)

</div>

---

## Downloads

### Recommended
**[⬇️ Download the full beta4 package](downloads/Layer-Limit-Override-v2.0.0-beta4-By-Tilted42o.zip)**

### EXE only
**[⬇️ Download Layer Limit Override](downloads/Layer-Limit-Override-v2-By-Tilted42o.exe)**

> The full ZIP is recommended because it includes the profile file, instructions, checksum, and source snapshot.

---

## What it does

| Surface | Stock FH6 | Override |
|---|---:|---:|
| Left / Right / Top | 3,000 | **30,000** |
| Front / Rear / Wing / Glass | 1,000 | **10,000** |

Built for painters making detailed drift liveries, replicas, window decals, sponsor-heavy designs, and large Forza Painter imports.

---

## Installation

1. Download the **full beta4 package** above.
2. Extract the ZIP.
3. Close **Forza Horizon 6**.
4. Run **Layer Limit Override v2 - By Tilted42o.exe**.
5. Select your `forzahorizon6.exe`.
6. Confirm the detected build and apply the override.

### Missing Offline Patch

If the program says **MISSING OFFLINE PATCH**, your FH6 executable must be prepared first.

Download the DVS FH6 Offline Patcher here:

**https://github.com/DVS-code/FH6-Offline-Patcher/releases/tag/V1**

Apply the DVS patch properly, then run Layer Limit Override again and select the patched `forzahorizon6.exe`.

**Huge shoutout and credit to DVS for the FH6 Offline Patcher.** DVS is a separate project and is not bundled with this mod.

---

## Safety

Layer Limit Override is designed to **fail closed**.

It will not modify the executable if detection is missing, incomplete, ambiguous, or if the expected bytes do not match. It also creates a verified build-specific backup before replacing the file.

Backups use this format:

`forzahorizon6.exe.lu-layers-backup-<build-hash>`

Run the tool again on the same patched executable to restore its verified original backup.

---

## Project layout

- **downloads/** — ready-to-use EXE and full ZIP
- **src/** — complete Go source
- **docs/** — technical documentation, checksum, and test results
- **LayerLimitOverride.profiles.json** — external compatibility profiles
- **.github/workflows/** — automated Windows build

---

## Beta status

Current release: **v2.0.0-beta4**

The compatibility engine supports exact known builds plus semantic layer-table detection for prepared executables. Unknown layouts are rejected instead of being patched blindly.

If an unsupported prepared build is detected, the tool can generate a small compatibility report containing hashes and detector information without including the full game executable.

---

## Credits

**Layer Limit Override:** Tilted42o  
**FH6 Offline Patcher:** DVS

This is an unofficial community project and is not affiliated with Microsoft, Xbox, Playground Games, Turn 10, or the Forza franchise.
