---
name: project-agent-orchestration
description: "Mandatory workflow when planning, implementing, validating, reviewing, or closing SOW-driven work in this repository."
---
# Project Agent Orchestration

## Purpose

Keep implementation and review reproducible, evidence-based, and bounded to
this repository.

## Scope

Use this skill when:

- creating or updating implementation SOWs;
- writing prompts for external reviewers;
- deciding whether a phase can advance;
- recording review findings, fixes, validation, or production-grade status.

Do not use this skill for:

- trivial wording-only changes that do not require a SOW;
- end-user SDK behavior decisions covered by the journal compatibility skill.

## Mandatory Knowledge

- The project manager implements the active SOW, performs local validation, and
  owns the outcome. Evidence: `AGENTS.md`.
- External review is the default, user-waivable technical review gate. Before
  submitting code to GitHub, recommend external review when the current change
  set has not already been reviewed. Evidence: `AGENTS.md`.
- Do not run external reviewers until the user authorizes them. Once
  authorized, follow the standing authorization and any limits for the current
  conversation. Evidence: the system-wide `external-reviewers` skill.
- The system-wide `external-reviewers` skill is the sole source of truth for
  available reviewers, model selection, harnesses, prompts, authorization
  scope, retries, and failure handling. Do not copy those details here.
- Current review cadence is whole-SOW batching: finish the complete active SOW,
  run local validation, update SOW evidence, then review the entire SOW and
  changed surface as one meaningful batch when authorized. Do not review
  trivial edits or partial churn. Evidence: `AGENTS.md`.
- CRITICAL: Do not make changes outside this repository. This applies to all assistants and external reviewers.
- The only write exception outside the repository is `/tmp`; prefer `.local/` inside this repository for scratch files.
- After each implementation chunk is implemented, reviewed, and verified, prefer committing that chunk before starting the next chunk. Stage explicit files only; never use `git add -A` or `git add .`.
- Reviewer agents must run read-only. Use the system-wide
  `external-reviewers` skill rather than repository-local harness instructions.
- Read-only dependency metadata commands can still write package caches. Prompts that allow dependency inspection must either forbid dependency-fetching commands or require cache/output variables under `.local/` or `/tmp`, including `GOMODCACHE`, `GOCACHE`, `GOPATH`, `npm_config_cache`, `PIP_CACHE_DIR`, `CARGO_HOME`, and equivalent tool caches.
- Journal work must not probe the live host journal. External-agent prompts for journal compatibility work must forbid `systemd-cat`, `logger`, live `journalctl` without `--file` or a repository-local `--directory`, writes to `/var/log/journal` or `/run/log/journal`, and any systemd command that changes host journal state.
- Core SDK runtime work must preserve the four-layer runtime-purity split from `AGENTS.md`: core file-format SDK, systemd/journald compatibility layer, optional identity helper, and optional writer-lock helper. Prompts must not ask agents to put host identity discovery or cooperating-writer locking back into core reader/writer paths.
- Core reader/writer runtime prompts must forbid external programs and host-observation sources in core code, including `/proc`, `/host/proc`, `/etc/machine-id`, platform registries, `sysctl`, `system_profiler`, `ps`, shell commands, subprocess APIs, and equivalent mechanisms. These are allowed only in explicitly named optional helper code and tests for those helpers.

Canonical external-agent prompt block:

```text
CRITICAL REPOSITORY BOUNDARY:
- Do not make changes outside this repository for any reason.
- Repository path: current repository root.
- You may inspect external references read-only when the task requires it.
- Write, edit, delete, move, reset, checkout, install, generate, cache, or format nothing outside this repository.
- The only write exception outside the repository is /tmp.
- Prefer .local/ inside this repository for scratch work, generated temporary files, cloned references, logs, and working notes.
```

## Best Practices

- Split work into small SOWs with one concrete deliverable and clear acceptance gates.
- Work on exactly one active SOW at a time.
- Record external-review authorization or waiver in the active SOW.
- Use the system-wide `external-reviewers` skill at meaningful review
  boundaries after whole-SOW implementation and local validation.
- Keep reviewer prompts neutral: include the original request, SOW filename, changed scope, validation commands, the canonical repository-boundary block, and ask for unwanted side effects and security issues.
- For SOWs touching runtime purity, ask reviewers to verify that core SDK code has no implicit host identity discovery, subprocess execution, or automatic writer locking, and that optional helpers are opt-in and documented separately.
- For dependency research or package metadata checks, include explicit cache redirection instructions before allowing commands such as `go get`, `go list`, `npm view`, `npm pack`, `pip download`, `pip index`, `cargo metadata`, or `cargo doc`.
- Under standing authorization, repeat review after material fixes using the
  same whole-SOW scope until no blocking issues remain.

## Bad Practices

- Do not let any assistant or external agent edit outside this repository.
- Do not run external reviewers before user authorization.
- Do not hard-code reviewer names, model identifiers, or harness commands in
  repository instructions.
- Do not let reviewers make changes; reviewers must be read-only.
- Do not let external agents run package-manager commands with default caches, because they can write under home directories even when the visible command output is read-only.
- Do not narrow follow-up reviewer prompts to only the last fix; keep the original review scope and add fix notes.
- Do not advance a SOW on "mostly ok" or unresolved production-grade doubts.

## Workflow Checklist

1. Confirm the active SOW has a completed pre-implementation gate.
2. Implement the approved SOW plan and record changed files and evidence.
3. Complete local validation and update the SOW.
4. If external review is not authorized, recommend it at the whole-SOW boundary
   and before GitHub submission. If the user waives it, record the waiver.
5. When external review is authorized, use the system-wide
   `external-reviewers` skill and record every finding with its disposition.
6. Fix verified blocking findings and, under standing authorization, re-review
   after material fixes until phase gates are satisfied.
7. Run the project-local audit and record results before closing.
8. If the audit fails, repair the issue inside this repository, rerun the audit, and record the clean result before closing.
9. Prefer committing the verified chunk before starting the next work chunk.

## Validation Checklist

Before claiming a phase is ready to advance:

- Active SOW records implementation ownership, validation, review
  authorization or waiver, reviewer runs when authorized, findings, and
  dispositions.
- External review has either completed with blocking findings resolved and
  re-reviewed, or the user waiver is recorded.
- `.agents/sow/audit.sh` has been run and relevant findings are resolved or recorded.
- No durable artifact contains raw sensitive data.
- No changed file sits outside this repository.
- Verified chunks are committed with explicit path staging before the next chunk starts unless the SOW records why commit was skipped.

## Evidence

- `AGENTS.md`: project roles, SOW gates, repository boundary, and review policy.
- System-wide `external-reviewers` skill: current reviewer availability,
  selection, harness, prompt, authorization, retry, and failure rules.

## Update Rules

Update this skill when:

- the user changes repository review policy or cadence;
- the system-wide reviewer skill changes its repository-facing contract;
- a review cycle exposes a missed orchestration failure mode;
- repository boundary or scratch-space policy changes.
