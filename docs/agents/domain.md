# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

## Before exploring, read these

- **`GLOSSARY.md`** at the repo root.
- **`docs/adr/`**: read ADRs that touch the area you're about to work in.

If any of these files don't exist, **proceed silently**. The `/domain-modeling` skill creates them lazily when terms or decisions get resolved.

## File structure

Single-context repo layout:

```
/
├── GLOSSARY.md
├── CODING_STANDARDS.md
├── docs/adr/
├── internal/
└── web/
```

## Use the glossary's vocabulary

When your output names a domain concept (in code identifiers, UI copy, an issue title, a refactor proposal, a hypothesis, or a test name), use the term as defined in `GLOSSARY.md`.

If the concept you need isn't in the glossary yet, either reuse existing project language or note the gap for `/domain-modeling`.

## Flag ADR conflicts

If your output contradicts an existing ADR, surface it explicitly:

> _Contradicts ADR-0003 (concentric clean architecture), but worth reopening because…_
