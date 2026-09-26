# FH6 Livery Limit Override — By Tilted42o

A Forza Horizon 6 livery editor modification for creators who need more room for detailed liveries, replicas, window decals, sponsor-heavy designs, and Forza Painter artwork.

## Layer limits

| Surface | Vanilla FH6 | Layer Limit Override |
| --- | ---: | ---: |
| Left / Right / Top | 3,000 | 30,000 |
| Front / Rear / Wing / Glass and other supported surfaces | 1,000 | 10,000 |

## Current version

**Layer Limit Override v2.0.0-beta4**

v2 replaces the old single-layout detector with a safer compatibility engine:

- Exact known-build profiles are checked first.
- Unknown prepared builds can use the semantic layer-table detector.
- The tool refuses to modify an executable when detection is missing, partial, or ambiguous.
- Every planned original byte is verified before installation.
- Build-specific verified backups are retained for restoration.
- Unsupported builds can generate a compatibility report without including the full game executable.

## Important — Offline Patch prerequisite

Some stock/protected FH6 executables do not expose the layer-limit routine in the form required by Layer Limit Override.

If the tool displays **MISSING OFFLINE PATCH**, close FH6 and download/apply the DVS FH6 Offline Patcher first:

https://github.com/DVS-code/FH6-Offline-Patcher/releases/tag/V1

After that patch has been applied properly, run Layer Limit Override again and select the patched `forzahorizon6.exe`.

**Huge shoutout and credit to DVS for the FH6 Offline Patcher.** DVS is a separate project and is not bundled with Layer Limit Override.

## Installation

1. Download the latest Layer Limit Override package.
2. Extract the whole ZIP to a normal folder.
3. Close Forza Horizon 6.
4. Run **Layer Limit Override v2 - By Tilted42o.exe**.
5. Select your installed `forzahorizon6.exe`.
6. If **MISSING OFFLINE PATCH** appears, use the DVS link above, apply that prerequisite properly, then rerun Layer Limit Override.
7. Confirm the detected build and apply the layer override.
8. Keep the generated backup if you want one-click restore for that game build.

## Restore

Run Layer Limit Override again and select the same patched executable. When its verified patched state is detected, the tool can restore the matching original backup.

Backups use a build-specific name:

```
forzahorizon6.exe.lu-layers-backup-<build-hash>
```

This prevents a backup from an older FH6 update from being mistaken for the current build.

## Compatibility profiles

Additional exact-build profiles can be placed in:

```
LayerLimitOverride.profiles.json
```

See [PROFILE_FORMAT.md](PROFILE_FORMAT.md) for the schema and safety requirements.

## Beta status

This is a compatibility-engine beta. It is deliberately designed to **fail closed** rather than guess.

If a prepared executable is still unsupported, no file is changed. Generate the compatibility report and include it with the FH6 version/storefront when reporting the issue.

## Credits

- **Layer Limit Override:** Tilted42o
- **DVS FH6 Offline Patcher:** DVS — https://github.com/DVS-code/FH6-Offline-Patcher

This is an unofficial community project and is not affiliated with Microsoft, Xbox, Playground Games, Turn 10, or the Forza franchise.
