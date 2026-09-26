# FH6 Livery Limit Override v2.0.0-beta4

Forza Horizon 6 livery layer limit override by Tilted42o.

## Features

- Left / Right / Top: **30,000 layers**
- Front / Rear / Wing / Glass: **10,000 layers**
- Exact known-build detection
- Semantic layer-table fallback for compatible prepared builds
- Verified build-specific backups
- Restore support
- Compatibility report generation for unsupported builds
- Fails closed if the executable cannot be verified safely

## Requirements

- Windows 10 / 11, 64-bit
- Forza Horizon 6
- Keep all files from the download together
- Some stock/protected FH6 executables require the DVS FH6 Offline Patcher first

DVS Offline Patcher:
https://github.com/DVS-code/FH6-Offline-Patcher/releases/tag/V1

## Install

1. Download `Layer-Limit-Override-v2.0.0-beta4-By-Tilted42o.zip`.
2. Extract the ZIP to a normal folder.
3. Keep the extracted files together.
4. Close Forza Horizon 6.
5. Run `Layer Limit Override v2 - By Tilted42o.exe`.
6. Select your `forzahorizon6.exe`.
7. Apply the override.

If the tool shows **MISSING OFFLINE PATCH**, apply the DVS offline patch first, then rerun Layer Limit Override and select the patched FH6 executable.

## Notes

- The tool creates a verified build-specific backup before replacing the game executable.
- Run the tool again on the same patched executable to restore its verified original backup.
- Unsupported or ambiguous builds are rejected instead of patched blindly.
- Keep `LayerLimitOverride.profiles.json` beside the tool.

## Credits

- **Tilted42o** — Layer Limit Override
- **DVS** — FH6 Offline Patcher
