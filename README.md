<div align="center">

# FH6 Livery Limit Override

### By Tilted42o

**Break past Forza Horizon 6's stock livery layer limits.**

[![Version](https://img.shields.io/badge/version-v2.0.0--beta4-blue)](https://github.com/lilanimal42o/FH6-Livery-Limit-Override-By-Tilted/releases/latest)
[![Windows](https://img.shields.io/badge/platform-Windows%2064--bit-lightgrey)](#installation)
[![Build](https://img.shields.io/github/actions/workflow/status/lilanimal42o/FH6-Livery-Limit-Override-By-Tilted/build-windows.yml?label=build)](../../actions)

## ⬇️ DOWNLOAD

### **[Download Layer Limit Override v2.0.0-beta4](https://github.com/lilanimal42o/FH6-Livery-Limit-Override-By-Tilted/releases/latest/download/Layer-Limit-Override-v2.0.0-beta4-By-Tilted42o.zip)**

**Download the ZIP above, extract it, and keep the included files together.**

</div>

---

## Layer Limits

| Surface | Stock FH6 | Layer Limit Override |
|---|---:|---:|
| Left / Right / Top | 3,000 | **30,000** |
| Front / Rear / Wing / Glass | 1,000 | **10,000** |

Built for detailed liveries, replicas, window decals, sponsor-heavy designs, and large Forza Painter imports.

## Installation

1. Download **Layer-Limit-Override-v2.0.0-beta4-By-Tilted42o.zip** from the link above.
2. Extract the ZIP to a normal folder.
3. **Keep everything in that folder together.**
4. Close Forza Horizon 6.
5. Run **Layer Limit Override v2 - By Tilted42o.exe** from the extracted folder.
6. Select your `forzahorizon6.exe`.
7. Confirm the detected build and apply the override.

> **Do not pull the EXE out by itself.** The release is designed to be used as the complete extracted folder.

## Missing Offline Patch

If Layer Limit Override displays **MISSING OFFLINE PATCH**, your FH6 executable needs the DVS offline patch first.

Download it here:

**https://github.com/DVS-code/FH6-Offline-Patcher/releases/tag/V1**

Apply the DVS patch properly, then run Layer Limit Override again and select the patched `forzahorizon6.exe`.

**Huge shoutout and credit to DVS for the FH6 Offline Patcher.**

DVS is a separate project and is not bundled with Layer Limit Override.

## Safety

Layer Limit Override is designed to fail closed rather than guess.

If detection is missing, incomplete, ambiguous, or the expected bytes do not match, the game executable is not modified.

A verified build-specific backup is created before replacement:

`forzahorizon6.exe.lu-layers-backup-<build-hash>`

Run the tool again on the same patched executable to restore its verified original backup.

## Beta Status

Current release: **v2.0.0-beta4**

If an unsupported prepared build is detected, Layer Limit Override can generate a small compatibility report containing hashes and detector information without including the full game executable.

## Credits

**Layer Limit Override:** Tilted42o  
**FH6 Offline Patcher:** DVS

---

> GitHub automatically shows **Source code (zip)** and **Source code (tar.gz)** on every release. Those are GitHub-generated repository snapshots, **not the Layer Limit Override download**. Use the download button at the top of this page.

This is an unofficial community project and is not affiliated with Microsoft, Xbox, Playground Games, Turn 10, or the Forza franchise.
