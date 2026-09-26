# Layer Limit Override v2 profile format

`LayerLimitOverride.profiles.json` uses schema 1:

```json
{
  "schema": 1,
  "profiles": [
    {
      "name": "descriptive build name",
      "original_sha256": "64 lowercase hex characters",
      "patched_sha256": "64 lowercase hex characters",
      "file_bytes": 200350720,
      "patches": [
        {"offset": 123456, "before": "e803", "after": "1027"}
      ]
    }
  ]
}
```

Safety rules enforced by the program:

- the selected file must exactly match `original_sha256` before apply or
  `patched_sha256` before restore;
- patches must be non-empty, equal length and non-overlapping;
- every `before` byte must match at its exact offset before any output is used;
- the fully patched output must exactly match `patched_sha256`;
- a verified original backup is created before replacement;
- the staged file and final installed file are hashed again;
- a changed file during preparation aborts the operation.

`file_bytes` is descriptive in beta4; the SHA-256 values are authoritative.
