# AGENTS.md

These rules are MANDATORY for opencode and any AI agent operating in this repository. Read this file before executing any command.

## Non-Negotiable Rules

1. **High-Care Processing Standard** — Every task must be processed with the highest standard of care and caution. Double-check before acting.

2. **No Unauthorized Changes** — Never destroy, delete, rearrange, or modify anything without explicit user approval. This includes files, directories, git history, and configuration.

3. **No Over-Assumption** — Stay strictly focused on the user's request. Do NOT assume the user wants changes to other parts of the system beyond the explicit request. When in doubt, ask first.

4. **Never Reduce or Break Features** — Reducing existing features, causing features to stop working, or introducing regressions is STRICTLY FORBIDDEN. All changes must be backward-compatible and additive only, unless the user explicitly approves a breaking change.

5. **Git Operations Require Explicit Permission** — `git add`, `git commit`, `git push`, and pushing tags (e.g. `git push origin <tag>`) may ONLY be performed when the user explicitly requests them. Never perform these operations proactively or as a follow-up without being asked.

6. **Keep Knowledge Graph Updated** — After any codebase changes (new features, refactoring, bug fixes), run `graphify update .` to update the knowledge graph. This ensures the graph reflects the current state of the codebase for accurate queries and analysis.

7. **SKILL.md Changes Are Feature-Only** — Changes to `.agents/skills/dcd-documents/SKILL.md` are permitted ONLY as feature additions or feature removals (e.g., introducing a new DCD construct, or removing support for an existing one). Restyling, rewording, restructuring, or any other non-feature edits to that file are not permitted without explicit user approval.

8. **Every Change Requires User Double-Check** — Every change must be double-checked by the user before it is considered done. After making any change, present the result for user review and wait for approval before committing, pushing, or treating the task as complete.

9. **Planning Is Mandatory Before Changes** — Planning MUST be done before starting any change. Present a plan (e.g., a todo list or a short written plan) to the user and get approval before executing any modification.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
