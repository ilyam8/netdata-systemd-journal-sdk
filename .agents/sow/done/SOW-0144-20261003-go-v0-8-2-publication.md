# SOW-0144 - Go v0.8.2 Publication

## Status

Status: completed

Sub-state: closed in release PR #2 under the user's explicit close-out
instruction. Tag publication and public-module verification run after merge.

## Requirements

### Purpose

Complete the Go `0.8.2` publication record in release PR #2 and document the
post-merge procedure that makes `Log.CloseWithoutRetention()` consumable as
a versioned module.

### User Request

The user requested a release containing the new API through a PR with its
SOWs closed. The user explicitly directed that SOW-0144 be marked completed
and pushed to PR #2, with merge as the release handoff.

### Assistant Understanding

Facts:

- Go `0.8.2` is an additive Go-only release; Rust registry packages stay `0.8.1`.
- SOW-0143 prepares the docs and validates the already-reviewed merged API.
- The Go module is `github.com/netdata/systemd-journal-sdk/go`; root `v0.8.2`
  and submodule `go/v0.8.2` tags must peel to the same merged release commit.
- Existing release tags must not be moved, deleted, or force-pushed.
- The user overrides the initial pending-publication lifecycle: this record
  must be completed under done in PR #2 before merge. This records the chosen
  close-out timing; it does not claim tags or public downloads already exist.

Inferences:

- The user's release authorization applies after the required PR merges.

Unknowns:

- The merged release-preparation commit and proxy propagation time are not
  known before the PR merges; determine both before claiming publication.

### Acceptance Criteria

- This SOW is completed under done in PR #2, and both status summaries and
  related release records reflect the user's chosen close-out timing.
- The publication procedure requires a merged release PR, relevant CI/scanner
  gates, fresh paired annotated tags with identical targets, and a clean
  downloaded-module consumer check before publication is reported successful.
- Candidate-source API validation remains supported by SOW-0143; final tracking
  changes pass the SOW audit and whitespace checks without runtime changes.
- Records distinguish completed PR close-out from post-merge tag execution;
  no unperformed publication or public download is claimed as verified.

## Analysis

The initial split left this publication record pending until PR merge. The
user explicitly rejected that lifecycle and requires completed SOWs in the
release PR. Close this record in the PR and retain the publication procedure
as the required action after merge, without another pending SOW.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- A merged API alone is not a published Go module version; fresh correctly
  prefixed immutable tags and downloaded-module verification are required.

Evidence reviewed:

- SOW-0143 and the repository release-tagging skill define scope and sequence.
- The user's explicit instruction defines completion in PR #2 before merge.

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

1. Mark this SOW completed, move it to done, and synchronize both ledgers and
   related release records; validate, commit, and push together to PR #2.
2. After merge, confirm the merged release commit without discarding work;
   recheck source/version metadata, CI/scanners, and absent target tags.
3. Create and push both annotated tags, then verify peeled remote targets.
4. Remove the consumer's local replacement in scratch, resolve the public
   version, execute the API check, and report verified publication.

Validation plan:

- PR close-out: status/directory consistency, same-reference scan, whitespace
  checks, artifact checks, and SOW audit.
- After merge: remote tag lookup, public module metadata/download, and clean
  minimum-toolchain consumer. These checks are not reported as executed yet.

Artifact impact plan:

- AGENTS.md and runtime skills: current release workflow already applies.
- Specs and consumer docs: version guidance is delivered by SOW-0143.
- End-user/operator skills: none are declared in the repository.
- SOW lifecycle and status summaries: complete in PR #2 under the explicit
  user instruction; retain the publication procedure in this completed record.

Open-source reference evidence:

- Go module reference: `https://go.dev/ref/mod#vcs-version`.
- The exact merged SDK commit is determined after PR #2 merges.

Open decisions:

- The user requires this SOW completed in PR #2. No lifecycle decision remains.
  Release authorization is present; actual tag publication still follows merge.

## Implications And Decisions

1. Publish Go `0.8.2` after the required PR merges; do not publish Rust crates.
2. Respect immutable existing tags; stop rather than replace a conflicting tag.
3. Complete this SOW in PR #2 as the user explicitly requires. Post-merge
   publication is a delivery action, not a separate pending SOW.

## Plan

Complete and push PR close-out; after merge, verify CI/source, tag, verify
public consumption, and report the result.

## Implementation And Review Plan

- Close the tracking record now; publish sequentially after PR merge using
  the repository release skill.
- Reuse reviewed source and preparation evidence; re-review any new source
  difference before publication. No external reviewer is authorized for new work.

## Execution Log

### 2026-10-03

- Created as the concrete post-merge publication follow-up for SOW-0143.
- The user explicitly directed completion in PR #2 rather than a pending SOW.
  Moved this record to done and updated the related SOWs and both summaries.
- API, dependencies, compiler requirements, and Rust package versions are
  unchanged by this lifecycle correction.

## Validation

- Acceptance: completed status and done directory are included in PR #2 with
  synchronized release records and the user's explicit close-out decision.
- Tests: final SOW audit and whitespace checks validate this tracking change.
  Candidate-source tests and real-use evidence remain recorded in SOW-0143.
- Tag and proxy checks run after PR merge; public publication is not claimed
  by this completed PR close-out record.
- Reviewer findings: the API and release preparation passed independent review;
  this update applies the user's lifecycle instruction without runtime changes.
- Sensitive data gate: no credentials or real journal data appear here.
- Artifact maintenance: AGENTS and runtime project skills need no change for
  this explicit user override. Specs and consumer guidance belong to SOW-0143;
  no runtime contract or install-version change is introduced here. SOW-0142,
  SOW-0143, both summaries, and the PR description reflect this close-out.
- Same-failure scan: related records no longer describe SOW-0144 as currently
  open or pending; historical creation evidence is retained as history.
- End-user/operator docs: published install guidance is unchanged; the release
  procedure below and PR description explain the publication timing.
- End-user/operator skills: the declared inventory contains none.
- Follow-up mapping: the user explicitly directs the remaining tag/download
  checks to execute after merge under this completed record, without another
  pending SOW or a separate lifecycle-only commit.

## Post-Merge Publication Procedure

1. Confirm PR #2 is merged and identify its merged commit; verify relevant CI
   and scanner gates passed or have evidence-backed dispositions.
2. Recheck that `v0.8.2` and `go/v0.8.2` do not conflict with existing tags.
   Create and push fresh annotated tags at the same merged release commit;
   verify identical remote peeled targets.
3. Resolve `github.com/netdata/systemd-journal-sdk/go@v0.8.2` in a clean Go
   `1.26.2` consumer without a local replacement and with CGO disabled.
4. Repeat the one-second retention, append, 1.5-second wait, special close,
   reopen-without-retention, and readable-entry check in both naming modes.
5. Report the actual tag targets and downloaded-module results. Keep raw logs
   and caches under ignored repository-local scratch; never print credentials.

## Outcome

Publication close-out is completed in release PR #2 as explicitly instructed.
The publication procedure runs after merge; remote tags and public downloads
have not been reported as executed.

## Lessons Extracted

Record the user's required PR completion timing while keeping actual tag and
download evidence distinct from prepared publication steps.

## Followup

No work beyond the stated publication acceptance criteria is deferred.

## Regression Log

No regression is being reopened.
