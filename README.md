<div align="center">

# FH6 Livery Limit Override

### By Tilted42o

**More room for serious Forza Horizon 6 painters.**

[![Windows](https://img.shields.io/badge/platform-Windows%2064--bit-lightgrey)](#installation)
[![Build](https://img.shields.io/github/actions/workflow/status/Tilted42o/FH6-Livery-Limit-Override-By-Tilted/build-windows.yml?label=build)](../../actions)

</div>

---

# Choose Your Edition

## Sticker Only Edition

For painters who only need the increased in-editor sticker/surface limits.

- **Left / Right / Top:** 30,000 layers
- **Other supported sticker surfaces:** 10,000 layers
- Keeps the normal saved decal/vinyl-group limit

### [Download Sticker Only Edition](https://github.com/Tilted42o/FH6-Livery-Limit-Override-By-Tilted/releases/tag/v2.0.0-beta4)

---

## Sticker + Decal Edition

For painters who also need larger saved decals / vinyl groups.

- **Left / Right / Top:** 30,000 layers
- **Other supported sticker surfaces:** 10,000 layers
- **Saved decal / vinyl-group limit:** 6,500 direct child entries
- Can upgrade an existing Sticker Only installation
- Can migrate the experimental beta5 30,000 decal cap down to the release 6,500 cap

### [Download Sticker + Decal Edition v1.0.0](https://github.com/Tilted42o/FH6-Livery-Limit-Override-By-Tilted/releases/tag/sticker-decals-v1.0.0)

> **Performance warning:** decals/vinyl groups above Forza Horizon 6's original 3,000-layer limit can lower frame rates while inside the livery editor and can make saving take longer. Based on testing so far, this appears to affect the editor only; no noticeable performance impact has been observed after leaving the editor and returning to normal driving/gameplay.

---

## Which one should I use?

Use **Sticker Only** if you just want the larger surface limits and do not need saved decals/vinyl groups above the stock limit.

Use **Sticker + Decal** if you create or import larger saved decals/vinyl groups and want the 6,500 limit as well.

Both editions keep the same 30,000 / 10,000 sticker-surface limits.

## Installation

1. Open the release page for the edition you want.
2. Download the single ZIP uploaded under **Assets**.
3. Extract the **entire ZIP** to one folder.
4. Keep all included files together.
5. Close Forza Horizon 6.
6. Run the included Layer Limit Override EXE.
7. Select your installed `forzahorizon6.exe`.
8. Read the confirmation screen and apply the override.

> **Do not pull the Override EXE out by itself.** The included files are meant to stay together.

## Missing Offline Patch

If Layer Limit Override displays **MISSING OFFLINE PATCH**, your FH6 executable needs the DVS offline patch first.

**DVS FH6 Offline Patcher:**  
https://github.com/DVS-code/FH6-Offline-Patcher/releases/tag/V1

Apply the DVS patch properly, then run Layer Limit Override again and select the patched `forzahorizon6.exe`.

**Huge shoutout and credit to DVS for the FH6 Offline Patcher.** DVS is a separate project and is not bundled with Layer Limit Override.

## Safety

Layer Limit Override is designed to fail closed rather than guess.

If detection is missing, incomplete, ambiguous, or the expected bytes do not match, the game executable is not modified. A verified build-specific original backup is kept for restore.

## Credits

**Layer Limit Override:** Tilted42o  
**FH6 Offline Patcher:** DVS

---

> GitHub automatically adds **Source code (zip)** and **Source code (tar.gz)** to every release. Those are GitHub-generated repository snapshots. Download the actual Layer Limit Override ZIP listed as the release asset.

This is an unofficial community project and is not affiliated with Microsoft, Xbox, Playground Games, Turn 10, or the Forza franchise.
