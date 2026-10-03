# SOW-0144 - Go v0.8.2 Publication

## Status

Status: open

Sub-state: waits for the SOW-0143 release-preparation PR to merge.

## Requirements

### Purpose

Publish and verify Go `0.8.2` from the merged release-preparation PR, making
`Log.CloseWithoutRetention()` consumable as a versioned module.

### User Request

The user requested a release containing the new API, then required a PR with
the preparation SOW closed before publication.

### Assistant Understanding

Facts:

- Go `0.8.2` is an additive Go-only release; Rust registry packages stay `0.8.1`.
- SOW-0143 prepares the docs and validates the already-reviewed merged API.
- The Go module is `github.com/netdata/systemd-journal-sdk/go`; root `v0.8.2`
  and submodule `go/v0.8.2` tags must peel to the same merged release commit.
- Existing release tags must not be moved, deleted, or force-pushed.

Inferences:

- The user's release authorization applies after the required PR merges.

Unknowns:

- The merged release-preparation commit and proxy propagation time are not
  known before the PR merges; determine both before claiming publication.

### Acceptance Criteria

- Confirm the release PR merged, its docs/SOW are included, and relevant CI and
  scanner gates passed or have evidence-backed dispositions.
- Create fresh annotated root/Go `0.8.2` tags on that pushed commit and verify
  identical remote peeled targets.
- A clean CGO-disabled Go `1.26.2` consumer downloads `v0.8.2`, calls the new
  method after an archive-age wait, reopens without retention, and reads the entry.
- Record sanitized publication evidence and complete this publication SOW
  together with an operator-facing release verification record.

## Analysis

SOW-0143 deliberately closes preparation in the PR as the user requires.
Publication is tracked here because it cannot precede that PR's merge.

## Pre-Implementation Gate

Status: blocked

Problem / root-cause model:

- A merged API alone is not a published Go module version; fresh correctly
  prefixed immutable tags and downloaded-module verification are required.

Evidence reviewed:

- SOW-0143 and the repository release-tagging skill define scope and sequence.

Affected contracts and surfaces:

- Public module tags, proxy/checksum resolution, and publication evidence.

Existing patterns to reuse:

- Paired annotated tags and the existing synthetic consumer from SOW-0143.

Risk and blast radius:

- Tags can be cached permanently. Stop on a conflicting existing tag or an
  unreviewed source change; no runtime source or Rust packages are modified.

Sensitive data handling plan:

- Reuse configured Git authentication; never write credentials. Logs and
  caches stay under `.local/`; durable evidence includes only public identities
  and aggregate checks.

Implementation plan:

1. Confirm the required PR is merged and refresh master without discarding work.
2. Recheck source/version metadata, CI/scanners, and absent target tags.
3. Create and push both annotated tags, then verify peeled remote targets.
4. Remove the consumer's local replacement in scratch, resolve the public
   version, execute the API check, and record verified publication.

Validation plan:

- Remote tag lookup, public module metadata/download, clean minimum-toolchain
  consumer, whitespace checks, artifact checks, and SOW audit.

Artifact impact plan:

- AGENTS.md and runtime skills: current release workflow already applies.
- Specs and consumer docs: version guidance is delivered by SOW-0143.
- End-user/operator skills: none are declared in the repository.
- SOW lifecycle and status summaries: activate only after PR merge; complete
  with the operator-facing publication verification record.

Open-source reference evidence:

- Go module reference: `https://go.dev/ref/mod#vcs-version`.
- The exact merged SDK commit will be recorded on activation.

Open decisions:

- The user-required release PR must merge first. This is the only activation
  gate; release authorization is already present in the conversation.

## Implications And Decisions

1. Publish Go `0.8.2` after the required PR merges; do not publish Rust crates.
2. Respect immutable existing tags; stop rather than replace a conflicting tag.

## Plan

Confirm merge, verify CI/source, tag, verify public consumption, record outcome.

## Implementation And Review Plan

- Implement sequentially after PR merge using the repository release skill.
- Reuse reviewed source and preparation evidence; re-review any new source
  difference before publication. No external reviewer is authorized for new work.

## Execution Log

### 2026-10-03

- Created as the concrete post-merge publication follow-up for SOW-0143.

## Validation

- Acceptance, real-use, tag, and proxy checks run after PR merge; publication
  is not claimed by this pending record.
- Sensitive data gate: no credentials or real journal data appear here.
- Artifact maintenance: preparation docs belong to SOW-0143; activation will
  record the merged commit and release verification record.
- Same-failure scan: recheck both tag names and all active version pins.
- Specs/skills: existing scope and workflow apply; no runtime contract changes.
- End-user/operator docs: publication evidence will identify the actual tag
  targets and installation results.
- End-user/operator skills: the declared inventory contains none.
- Follow-up mapping: this file owns all publication work deferred by SOW-0143.

## Outcome

Waiting for the release-preparation PR merge.

## Lessons Extracted

Separate verified release preparation from irreversible post-merge publication.

## Followup

No work beyond the stated publication acceptance criteria is deferred.

## Regression Log

No regression is being reopened.
