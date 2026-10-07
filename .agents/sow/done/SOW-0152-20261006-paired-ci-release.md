# SOW-0152 - Paired CI Release

## Status

Status: completed

`completed` is the successful terminal status. `done` is a directory name, not a status value.

Sub-state: manual Rust CI publication and the paired-tag maintainer handoff
are implemented, validated and reviewed. CI uses crates.io Trusted Publishing;
authorized maintainers push tags with existing access. The selected source,
Rust-first order and final Go verification are preserved. Owner setup and
actual publication belong to pending SOW-0151.

## Requirements

### Purpose

Publish the Rust crates through CI and keep Rust, repository and Go releases
on the same reviewed source commit.

### User Request

The user asked whether a CI publisher could be created and selected option A:
a manual Release button that publishes all eight Rust crates, then creates the
repository and Go tags at the same merged commit. The implementation belongs
in a separate PR with its SOW completed before submission.

The user subsequently clarified the responsibilities: CI uses crates.io
Trusted Publishing, while a maintainer with tag-push permission pushes the
tags using existing GitHub access. Deliver that simpler handoff, usable by
multiple authorized maintainers, with no additional GitHub App/account token.

### Assistant Understanding

Facts:

- Planning master `044de252f612` had checks and wiki publishing, but no
  crates.io publisher. PR #6 has since merged as `fba9d45d53044` and is
  integrated into this implementation branch.
- All eight public crates already exist at `0.8.2`; none has `0.9.0` as of
  2026-10-06. Trusted Publishing supports these existing crates.
- PR #6 prepares `0.9.0`, completes SOW-0150, and tracks actual publication in
  pending SOW-0151. This SOW delivers reusable automation, not a public release.
- The Go module lives in `go/` and needs `go/vVERSION` as well as `vVERSION`.
- Supported compiler minimums remain Rust 1.91 and Go 1.26.2.

Inferences:

- Selecting an explicit merged source SHA avoids changing release source when
  master advances while a partial publication is being resumed.
- Rust-first publication preserves the existing release failure contract.

Unknowns:

- The crate owner must configure Trusted Publishing, and a repository
  administrator must configure the release environment before activation.
  These account settings cannot be supplied through a source-code PR. The
  workflow will check the environment and fail clearly when setup is missing.

### Acceptance Criteria

- A manual canonical-master workflow accepts a stable version and a full
  merged source SHA; ordinary PR/push runs execute helper tests only.
- Validate source identity, workspace versions, published dependency pins,
  active install examples, paired tag targets, and Go module identity before
  publication. Run both language suites using declared compiler minimums.
- Dry-run and publish only the eight public crates in dependency order. Obtain
  a fresh temporary publishing permission for each new crate upload.
- Verify index availability, non-yanked version, archive checksum and clean
  VCS SHA. Resume only versions already published from the selected source.
- Build an exact-version Rust registry consumer and report its selected source
  before handing off annotated paired tags to a permitted maintainer. Document
  atomic creation/push using existing access and provide verification of
  canonical targets, matching Rust archive sources and exact Go consumption.
- Meaningful local tests cover partial publication, mismatched artifacts,
  incomplete indexing, wrong source/tag targets, and publication/tag failures
  without uploading crates or touching public release tags.
- Operator guide, release skill, product spec and SOW ledger describe the
  delivered contract; local audit and independent review pass.

## Analysis

Sources checked:

- `.github/workflows/{docs-examples,wiki,coverage}.yml`.
- `rust/Cargo.toml`, eight public manifests, `rust/Cargo.lock`, `go/go.mod`.
- `.agents/skills/project-release-tagging/SKILL.md`, product-scope spec,
  SOW-0145 publication evidence, SOW-0066 stable-release scope.
- PR #6 SOW-0150 and pending SOW-0151 at `6fe321b36567`.
- Official Cargo publish/package documentation, crates.io Trusted Publishing
  documentation and official `rust-lang/crates-io-auth-action`.
- Canonical repository workflows, environments, rulesets and remote tags.

Current state:

- No release environment is configured in the canonical repository.
- SOW-0066 is a separate eventual stable 1.0.0 release, not this automation.
- The preparation/publication split in PR #6 avoids overlapping implementation
  with this reusable workflow SOW. Actual publication stays with SOW-0151.

Risks:

- Public crate uploads cannot be replaced; a later crate can fail after earlier
  uploads succeed. Verify existing artifacts before skipping them on a rerun.
- Tags at the wrong commit can be cached by Go consumers. Reject conflicts,
  give the maintainer the exact CI source, document atomic pair pushes, and
  verify that Rust archives and both tags identify that source.
- Registry indexing and Go proxy propagation take time. Use bounded waits and
  report the exact incomplete step rather than claiming success.
- Temporary publishing permission lasts 30 minutes. Refresh per crate after
  its dry-run, with no publishing permission during language validation.

## Pre-Implementation Gate

Status: ready

The original scope passed this gate before implementation. Review exposed a
permission choice needed to preserve the approved recovery contract. The
user's subsequent clarification assigns tag creation to an authorized maintainer
using existing access; record the correction plan below before resuming code.

Problem / root-cause model:

- The repository has no CI publisher, so release preparation cannot make Rust
  packages consumable automatically. Separate manual tagging previously
  allowed Rust/Go source to differ. CI now publishes Rust from an explicitly
  selected merged commit and records it for the maintainer's matching tag
  handoff; final operator verification checks the shared source.

Evidence reviewed:

- Exact manifests and release skill establish the eight-package order.
- Cargo 1.91 supports locked packaging and workspace package selection.
- Official crates.io docs require an existing crate and owner-created binding
  to repository, workflow filename and optional environment.
- The official authentication action's `v1.0.5` resolves to
  `c6f97d42243bad5fab37ca0427f495c86d5b1a18`; the floating `v1` alias was absent.
- GitHub documents manual-dispatch default-branch availability and environment
  branch policies. Existing workflows establish pinned actions and non-cancelling
  concurrency conventions.

Affected contracts and surfaces:

- `.github/workflows/release.yml`, release helper and its tests, `RELEASING.md`,
  root README link, release-tagging skill, product-scope spec and SOW ledger.
- SDK runtime APIs and journal formats do not change.

Existing patterns to reuse:

- Pinned checkout/Rust/Go actions from docs-examples workflow.
- Canonical-repository guard from wiki workflow.
- Python standard-library test/helper style from `tests/docs/`.
- Runtime path discovery and ignored `.local/` caches; visible commands with
  preserved exit status from coverage scripts.

Risk and blast radius:

- Release operations affect immutable public artifacts only when a maintainer
  starts the configured manual workflow. Local validation uses controlled
  test registries and repository-owned Git repositories, never public uploads.
- Only the Rust publishing job obtains temporary crate permission after both
  language suites pass. Maintainers perform tag writes with their existing
  access. Workflow code remains at the dispatch master SHA while
  release source is checked out separately at the selected immutable SHA.

Sensitive data handling plan:

- No permanent crates.io token is requested, read or committed. Temporary
  action output is passed only through the publishing step environment.
- CI uses the built-in GitHub token for repository/environment reads only.
  It holds no tag-push permission or additional account credential. Maintainer
  tag pushes use normal Git authentication outside this workflow/helper.
- Durable evidence contains package/version/source hashes and aggregate test
  outcomes only; no personal names or account credentials.

Implementation plan:

1. Record approved option A and implement input/source/metadata/registry/tag
   checks in one standard-library Python helper with narrow CI-only mutations.
2. Add one workflow with PR helper checks, minimum-toolchain validation,
   serial per-crate dry-run/authentication/upload, fresh Rust consumption and
   a source receipt for maintainer tag creation. Add operator Go verification.
3. Add controlled tests for actual Git tag operations and simulated registry
   publication failures/resumption; run local minimum-toolchain checks.
4. Update operator guide, release skill, spec and status ledger. Review the
   complete SOW, run audit, complete/move the SOW, commit explicit paths and
   submit the separate PR. Do not publish a version as part of this SOW.

Validation plan:

- Standard-library unit/integration tests with controlled registry responses,
  archive contents and task-owned local Git repositories.
- Workflow actionlint, Python lint, meaningful command-error checks, and
  smoke checks against actual repository manifests and minimum toolchains.
- Full Rust and Go suites for the source used in the local smoke check.
- Local SOW audit, diff check, same-pattern scan and independent final review;
  recommend external review at the complete local-validation boundary.

Artifact impact plan:

- AGENTS.md: release behavior belongs in spec/skill; roles and workflow rules
  are already suitable and need no change.
- Runtime project skills: add CI publication and immutable resumption contract
  to the release-tagging skill.
- Specs: add the paired CI release operational contract.
- End-user/operator docs: add root `RELEASING.md` and a README link.
- End-user/operator skills: none are shipped by this project.
- SOW lifecycle: finish SOW-0152 with the workflow; actual 0.9.0 activation and
  publication remain in PR #6's pending SOW-0151, not falsely completed here.
- SOW-status.md: reflect current state, then completed automation outcome.

Open-source reference evidence:

- `rust-lang/crates-io-auth-action @ c6f97d42243bad5fab37ca0427f495c86d5b1a18`:
  `action.yml`, `README.md`. Inspected through official GitHub API read-only.
- Cargo docs: `https://doc.rust-lang.org/cargo/commands/cargo-publish.html`
  and `https://doc.rust-lang.org/cargo/commands/cargo-package.html`.
- Trusted Publishing: `https://crates.io/docs/trusted-publishing` and its
  official source in `rust-lang/crates.io`,
  `svelte/src/routes/docs/trusted-publishing/+page.svelte`.

Open decisions:

- Release trigger/scope resolved by user selection A on 2026-10-06.
- The user's earlier explicit external-review authorization applies to later
  meaningful milestones in this conversation. Review the complete locally
  validated candidate under that standing authorization before submission.
- Decision 2, superseded by maintainer-owned tag creation: GitHub permission for native atomic tag pushes to an
  older selected source after master's workflow changes. Options presented:
  A, a dedicated GitHub App installed only on this repository, with Contents
  and Workflows write permissions and a temporary token per tag job
  (recommended, long-term-best); B, a repository-limited account token with
  those permissions (surgical, simpler setup, account/expiry dependency).
  The user clarified that a maintainer with existing permission pushes tags.
  CI only needs Trusted Publishing for crates.io. No App or stored GitHub
  account token is required by the delivered workflow. Any authorized
  maintainer can perform the documented tag handoff after Rust CI succeeds.

## Implications And Decisions

1. 2026-10-06: user selected A, the long-term-best manual paired Release
   workflow, over tag-triggered Rust publication and manual Rust-only CI.
   Rust is verified first; then both annotated tags identify the same source.
2. Implementation uses official Trusted Publishing with a `release`
   environment limited to canonical master, an explicit source SHA, bounded
   registry waits and verified partial-release resumption. These deliver the
   approved failure/recovery contract without SDK runtime changes.
3. Separate PR and completed implementation SOW. Existing SOW-0151 owns actual
   release execution after preparation and workflow merge/account setup.
4. Review amendment: the built-in GitHub token cannot provide the Workflows
   permission needed for all supported historical native tag pushes. Retain
   the selected source and atomic paired-tag contract; resolve decision 2 rather
   than silently restrict source selection or create tags separately.
5. 2026-10-06: user rejected an App dedicated to this step and requested that
   multiple administrators can release. An account-token interpretation was
   prepared locally, then superseded before review/submission by decision 6.
6. 2026-10-06: user clarified that CI needs crates.io Trusted Publishing and
   an authorized maintainer pushes tags with existing GitHub access. Apply the
   simpler operator handoff: CI validates both languages and publishes Rust;
   after success, a maintainer pushes both annotated tags to the selected
   source and verifies Go consumption. Remove automatic tag writes and all
   additional GitHub account-token/App setup. This supersedes the automatic
   tag component of the initial option A while preserving its release order,
   language validation, shared source and recoverability.

### Maintainer Tag Handoff Correction Gate

- Problem and evidence: the reviewed built-in token cannot satisfy all
  historical native tag pushes. Retain the atomic pair and verified-source
  contract, and move the actual tag creation/push to a maintainer using
  existing access. Remove CI's tag-write implementation and its permission
  setup, eliminating the native-push permission gap in CI.
- Affected surfaces: remove the tag workflow job and the temporary tag-setup
  code; retain read-only tag/source checks and exact-version Go verification
  as an operator helper. Update tests, operator guide, release skill, product spec,
  this SOW/ledger, and the newly merged pending publication SOW-0151.
- Reuse: retain environment branch checks, CI-only crate mutation commands,
  registry resumption, immutable tag validation and exact-version consumers.
  Maintainers use normal native annotated atomic Git pushes.
- Risk: Rust publication can finish before the maintainer creates tags. The
  CI summary must explicitly hand off the selected source/version; only the
  execution SOW can claim the complete paired release after tag and Go checks.
  A conflicting tag stops verification and requires a version/source decision.
- Implementation: remove the tag job, CI tag command and account-permission
  helper. Write a Rust-success summary with the source and required tag handoff.
  Provide a read-only verify-go helper for exact paired-target and consumer
  validation using the selected commit's declared Go minimum. Document normal
  maintainer tag commands and multiple-operator access. Integrate PR #6
  and align its pending execution record with the already-approved CI method.
- Validation: meaningful missing/conflicting/complete remote-tag and exact Go
  verification tests, full helper suite, actionlint, Ruff and whole-surface
  audit. Verify no workflow/helper command creates tags or needs a stored
  GitHub account token. Existing Rust/Go
  runtime validation remains applicable because preparation runtime/manifests
  are unchanged. Complete external and independent final review again.
- Artifacts: guide/skill/spec record the operational permission contract;
  SOW-0151 stays pending until real publication. This SOW and ledger record
  the reviewed correction's completion. No new AGENTS responsibility or
  consumer wiki change is needed.
- Open decisions: none for implementation. A permitted maintainer performs
  the tag push after Rust publication using existing GitHub access.

## Plan

1. Release helper and controlled tests.
2. Manual workflow and operator/artifact updates.
3. Local validation, complete-SOW review, closure and PR submission.

## Implementation And Review Plan

Implementation:

- One implementation owner writes sequentially; read-only reviewers inspect
  the final complete surface. Caches/output stay in repository `.local/`.

Reviewers:

- Independent final review is required. Run the complete external review batch
  after local validation under the user's standing conversation authorization.

Repository boundary block for every external-reviewer prompt:

```text
CRITICAL REPOSITORY BOUNDARY:
- Do not make changes outside this repository for any reason.
- Repository path: current repository root.
- You may inspect external references read-only when the task requires it.
- Write, edit, delete, move, reset, checkout, install, generate, cache, or format nothing outside this repository.
- The only write exception outside the repository is /tmp.
- Prefer .local/ inside this repository for scratch work, generated temporary files, cloned references, logs, and working notes.
```

Failure handling:

- Record and correct verified implementation/test failures in this SOW.
- Publication timeout can occur after an upload; verify registry state before
  deciding whether that crate succeeded. Never blindly upload it again.
- Wrong SHA, artifact checksum, yanked version or tag conflict stops the release.
- Local work performs no crate upload, public tag push or account-setting change.

## Execution Log

### 2026-10-06

- Completed design research and read-only release planning. User approved A.
- Created dedicated branch from canonical master with a clean starting tree.
- Recorded gate and scope before implementation.
- Added helper, workflow, controlled failure/recovery tests and operator guide.
- Initial audit identified a missing explicit sensitive-data validation gate;
  added that declaration before rerunning the audit. No source or account
  operation was blocked by the audit finding.
- Finished helper tests, workflow/Python checks, minimum-toolchain SDK suites,
  real published-archive verification, a downloaded Go consumer, and a locked
  Cargo dry-run. Public artifacts and account settings were not changed.
- Verified a fresh exact-version Rust 0.8.2 registry consumer at Rust 1.91;
  it built successfully without path overrides. Repeated actionlint passed
  and its exit-code receipt is retained in ignored scratch.
- PR #6 merged as `fba9d45d53044d54163da53c903c7eedbde6e34f`. Relative to
  the validated preparation revision, its Rust/Go runtime and manifests are
  unchanged; only Rust README migration text changed in those directories.
  Fetched canonical master without changing the frozen candidate. Integrate
  that master and align pending SOW-0151 with the approved CI method before
  final submission; actual publication remains pending.
- Completed external round `f319cd6b43164705b98c1931d1313aa3`. Two valid
  reports contain non-blocking notes; three reviewers timed out, and one
  returned no accepted verdict. Collection reports failure, so this is not a
  successful six-reviewer gate. Preserve all results; do not retry this
  already-rejected candidate before the permission correction.
- Independent final review returned FAIL for the older-source native tag
  permission contract. Presented decision 2 and stopped dependent work.
- Recorded the user's clarification that maintainers push tags using existing
  access. Removed the CI tag job, account-token/App setup and native tag-write
  helper. Retained only canonical tag reads and a manual verification command.
- Preserved a signed private checkpoint for master integration, then rebased
  onto PR #6's merged source. The final public commit will include code and the
  completed SOW lifecycle together; no private checkpoint was pushed.
- Updated the Rust-success summary and guide for the maintainer handoff. The
  final verification checks both tag targets and all eight Rust archive source
  hashes before building the exact Go consumer. Its compiler/module declaration
  is read from the selected commit, independently of the working branch.
- Aligned pending SOW-0151 with the selected CI method and existing-access tag
  push. It remains open because owner setup and actual publication have not run.
- Retained the already-reported PR #6 preparation review-record gap as an
  explicit SOW-0151 execution prerequisite. A source merge is not a review
  waiver; this automation SOW does not silently resolve another SOW's gate.
- Whole-surface metadata audit found that a late package restricted to another
  registry could pass the public-package-set check, then fail only after its
  predecessors were published. Added a crates.io destination gate before
  any package operation and a controlled failure test. Explicit crates.io
  allowlists still pass.
- Completed external round `cf1a85d2cda447cb81ea8b86c47edd8f` for the frozen
  maintainer-handoff candidate. Four accepted reports return PRODUCTION GRADE;
  two reviewer coverage gaps are recorded below. Independent final review
  returns PASS for the complete ten-file surface against merged master.
- Clarified the operator retry text after review: a rejected atomic push
  changes neither remote tag, but another maintainer may have created correct
  tags. Re-check and preserve them; when local tags already match, retry only
  the push command. No workflow/helper/test implementation changes followed
  the broad review before initial submission. Retained fresh lint receipts
  with commands, exit codes, timestamps and input hashes in ignored scratch.
- Fresh pre-close audit after review records and operator wording changes
  exits zero and reports SOW initialization complete and clean. Complete/move
  this implementation SOW with its code and ledger in the same public commit;
  require the independent closure-impact check before submission.
- Submitted PR #7 with the signed implementation/completed-SOW commit.
  The release-helper CI passed; ordinary PR execution correctly skipped
  source publication jobs. A subsequent import-style comment is corrected
  with one from-import for TestCase, main and mock. The suggested aliased
  module import conflicted with Ruff PLR0402; direct imports satisfy both
  checks. All four test-class bases and the script entry point retain the
  same objects. Thirty-five tests, Ruff check/format and whitespace checks
  pass again, with code-bound receipts retained in ignored scratch.
- Reviewed all newly reported program/download hints: Bandit_B404 at the
  subprocess imports in the helper and tests (alerts 3686/3689), Bandit_B603
  at both program runners (3687/3690), and Bandit_B310 at the HTTP reader
  (3688). These are explicitly dispositioned as intended operations with
  restricted callers, not SDK runtime defects. Executables and argument
  vectors are fixed Git/Cargo/Go or controlled test commands; no shell is
  used, versions/commits are validated and crate names are constrained.
  Tests operate on repository-owned fixtures. HTTP callers build URLs from
  fixed HTTPS GitHub/crates.io hosts, the fixed repository/allowed crates and
  validated versions; there is no caller-supplied URL or scheme CLI option.
  The initial Codacy analysis gate failed on exactly these five findings.
  Record five rule-specific inline dispositions after the call-site audit,
  following tests/code_scanning/export_codacy_issues.py; no broad rule or
  configuration suppression and no operational behavior change is needed.
  The final three-file correction passes all 35 tests, Ruff check/format
  and whitespace checks again. Fresh input hashes and logs are retained in
  ignored scratch. Local Bandit is absent from the documented development
  environment; the pushed CI run will validate the inline dispositions.
  The post-submission SOW audit exits zero and reports complete and clean.
- The optional GOPROXY=direct suggestion is rejected for the primary guide:
  the default resolver exercises normal Go consumer availability. Canonical
  tags and archive sources are checked separately, lookup has a bounded
  wait, and the guide documents same-source retries after propagation lag.
  No fallback feature or SDK repair is deferred by these comments.
- The proposed direct-dependency/TOML expansion is rejected for the present
  release surface: every active Rust installation example uses the checked
  package/version inline-table form, and neither active scan directory has
  a TOML file. Current examples all target 0.9.0. Additional syntax support
  is an optional expansion when consumer examples actually introduce it;
  no current stale-version finding or contract mismatch was demonstrated.
- Adjudicated six additional post-submission comments. Documented the
  existing major-version 0/1 guard before dispatch and aligned SOW-0151's
  plan with the guide's preserve-existing/push-missing tag recovery. These
  are operator wording corrections with no helper/workflow behavior change.
  Rejected mandatory Go Origin.Hash: the official Go module proxy protocol
  requires Version metadata, not Origin, and a normal proxy may omit that
  field (https://go.dev/ref/mod#goproxy-protocol). Canonical tag targets and
  Rust archive sources are checked independently, exact Go module/version
  consumption is checked, and any provided conflicting origin hash fails.
  The active install-form finding duplicates the disposition above.
  The 120-minute job limit intentionally bounds one attempt; verified
  published crates resume without re-upload, so increasing a worst-case
  theoretical budget is not required by the release contract. No slow live
  upload was performed in this PR. Retained the eight explicit ordered
  prepare/auth/publish groups: their crate/step/token references match, fresh
  authentication follows each dry-run, and a matrix does not guarantee
  dependency order. Factoring correct groups is optional refactoring, not
  an unimplemented release requirement.

## Validation

### Acceptance And Executable Evidence

- Thirty-five helper tests pass. Controlled registry/archive cases cover version
  identity, checksum/source mismatches, yanked versions, API/index lag in both
  directions, transient reads, bounded waits, partial publication and an
  accepted upload whose command failed. A simulated interrupted batch resumes
  without uploading its verified prefix again.
- Actual task-owned bare Git repositories prove the documented maintainer's
  paired annotated/atomic pushes, verification without tag mutation,
  incomplete-pair rejection while preserving a correct root sibling,
  rejection of dirty/unmerged source and conflicting tags, and atomic push
  rejection leaving neither new tag remotely. An operator CLI test works
  without CI identity; another rejects matching tags whose source differs
  from the Rust archives before Go consumption. No public tags are created.
- Input/manual-branch/environment tests prove canonical-master dispatch and
  exact master-only environment policy checks. Metadata tests cover public
  package set, internal pins, unpublished dependencies and ordering. Consumer
  tests cover exact registry versions, source mismatches and command failures;
  crate publishing permission is supplied through environment rather than argv.
- Rust-success summary tests prove that the selected SHA and maintainer handoff
  are reported after verification, and a failed Rust check emits no successful
  summary. Multiple synthetic operators pass the same manual workflow inputs.
- `actionlint .github/workflows/release.yml`, Ruff check/format check for the
  Python 3.11 target, and `git diff --check` pass. The helper uses only Python
  standard-library modules available on the workflow's Ubuntu 24.04 runner.
- Actual PR #6 metadata and active installation examples pass for version
  0.9.0: 23 workspace members and exactly eight public crates. SDK source and
  manifests at the later `6fe321b36567` revision equal the validated
  `14fe7550f0fa` revision; its intervening changes are release records/docs.
- Rust 1.91 workspace tests pass: 538 passed, zero failed, five ignored across
  41 suite groups. Go 1.26.2 tests pass with CGO disabled. These retain the
  current supported compiler minimums and exercise both language SDKs.
- The first Go suite run alongside Rust hit the existing live-writer stress
  test's 45-second timeout. Its isolated retry passed in 19.15 seconds, then
  the entire Go suite passed when run alone. Resource contention is a working
  theory, not a proven cause. No SDK source was changed; the workflow runs the
  language suites sequentially, matching the successful validation.
- Real registry reads verify all eight existing 0.8.2 archives against API and
  index checksums, non-yanked metadata, package identity and clean Cargo VCS
  source `8ad648a7b4d36bf2d75ee6ee98b42446ff64a276`. The helper also downloads
  and builds an exact Go 0.8.2 consumer, checking its historical tag source
  `51bf47f1f90562f7b7717ca41aaaac74f75f90f0`. These different historical
  sources are the documented staggered-release exception, not a new paired
  release made by this SOW.
- A fresh Rust consumer requiring the exact published SDK 0.8.2 builds at
  Rust 1.91 using downloaded registry packages and no path overrides. This
  complements the real Go consumer and archive-verification evidence.
- The actual revised operator `verify-go` command passes against the existing
  paired 0.8.1 release at `9d5e3e19cf53179aaec3af67ac409d844a44c15f`:
  both annotated canonical tags, all eight non-yanked Rust archives and their
  source/checksum metadata agree, then a fresh exact-version Go consumer
  builds with CGO disabled at the selected Go 1.26.2 minimum. This exercises
  the complete maintainer verification path without uploading or pushing.
- A real locked `cargo publish --dry-run` for common 0.9.0 passes at Rust 1.91
  without publishing credentials. An initial offline attempt could not read
  the registry; the online read-only retry passed and explicitly aborted the
  upload because it was a dry-run. Production already uses online registry
  reads. Later 0.9.0 crate dry-runs require their predecessors to be published
  and therefore belong to actual execution, not this local implementation.
- Evidence logs remain in ignored scratch. Actual owner setup, temporary CI
  permission exchange and new-version publication cannot be exercised before
  workflow merge and owner configuration. Controlled tests and real existing
  artifact/consumer checks establish the implementation without claiming that
  an actual release has occurred.

### Review And Same-Failure Gate

- Round 1 result: FAIL for the original automatic-tag candidate. Independent
  review verified a P1 in its native tag permission path: that revision's
  `.github/workflows/release.yml:263,288` used the built-in token with Contents
  write, and `tests/release/release.py:530,545` created annotated tags and
  pushed them atomically. When a selected merged
  source's workflow tree is no longer retained by a branch tip, GitHub can
  reject that historical tag push for missing Workflows permission. A later
  rerun cannot obtain that permission from the built-in token. Already
  published Rust crates must keep the original source, so switching source
  is not a recovery method.
- Public reproduced case and GitHub Support response:
  `https://github.com/orgs/community/discussions/151442`; related native
  annotated-tag case:
  `https://github.com/orgs/community/discussions/26164`. Official reference
  permission sets:
  `https://docs.github.com/en/rest/git/refs#create-a-reference`; workflow
  permission keys:
  `https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions`.
  Native annotation was already used by the reproduced failure. Separate
  REST reference creation does not preserve the approved atomic pair.
- Qualification: current historical canonical branches retain the earlier
  workflow tree, so the initial 0.9.0 tag attempt is not proven to fail now.
  Those incidental branches are not a maintained retry guarantee. The
  realistic blocker is the supported retry after workflow changes and removal
  of such historical branch tips. No public mutation was used to reproduce it.
- External round 1: glm and deepseek returned PRODUCTION GRADE with P3 notes;
  minimax exited without an accepted verdict, while mimo, kimi and qwen hit
  their 1,800-second limits with no final reports. These four are coverage
  gaps, not positive votes. No verified transient transport failure supports
  a minimax retry. Timeout retries are not useful before the already-required
  material correction; the corrected complete candidate needs a fresh batch.
- Adjudicated P3 notes: optional Go origin metadata, future install-snippet
  styles, upload-failure diagnostics, token-window optimization and larger
  future HTTP responses do not demonstrate a current in-scope failure. The
  remote tag targets are independently checked; mismatched Go origin hashes
  are rejected when provided; upload errors remain visible; command duration
  is bounded; current crate responses are small. Do not expand the approved
  work for hypothetical new forms or sizes. An actionlint evidence concern
  was resolved by rerunning the check and retaining its successful receipt.
  The pre-existing manual fallback scan omission is outside this workflow's
  operative path, which includes the Go README.
- The earlier Go stress timeout does not establish an SDK regression: its
  isolated and full sequential retries pass, and this PR changes no SDK source.
  Record the evidence and reject an unrelated timeout increase or speculative
  repair SOW here.
- P1 disposition: the user's clarified responsibility split removes the
  affected CI tag-write path entirely. No CI job requests Contents write, no
  additional GitHub account token is referenced, and the helper never creates
  or pushes a tag. Maintainers use their existing permission after Rust CI
  passes. The revised complete candidate passed the new broad review below.
- Whole-surface audit checks all eight workflow prepare/auth/publish groups,
  source and manual-run gates, environment checks, registry absence/failure
  handling, upload resumption, immutable tag conflicts, and final consumer
  checks. Mutations are confined to the intended selected checkout and task
  scratch; no SDK runtime, host journal, service or account setting is altered.
- The local SOW audit reports initialization complete and clean after the
  explicit sensitive-data gate and again after the review/guide records.
  Closure also requires a fresh audit of final status/directory and ledger
  consistency, with the clean verdict inspected before commit/submission.
- Maintainer-handoff candidate: 35 helper tests, actionlint, Ruff check/format,
  whitespace checks and the local SOW audit pass. The updated public metadata
  gate rejects a package destination that excludes crates.io before registry
  or crate operations. Rust/Go source and manifests match merged master;
  no new SDK suite run is needed for unchanged source already validated above.
- External round 2: minimax, mimo, kimi and deepseek each return PRODUCTION
  GRADE after complete-surface static inspection. No P0/P1/P2 finding is
  established. glm fails with HTTP 429 and fails with HTTP 429 again on its
  single permitted retry; qwen reaches its 1,800-second limit without an
  accepted report. Collection exits 1 because of those failures. These two
  missing opinions are coverage gaps, not positive votes; no waiver is
  inferred. The qwen retry allowance is unused. Available full-scope reports,
  executable evidence and independent final review support the adjudicated
  PASS for this implementation and PR submission.
- Independent final review: PASS for all ten changed paths against
  `fba9d45d53044d54163da53c903c7eedbde6e34f`, with equal staged/working
  contents and no Rust/Go source or manifest diff. The remaining P3 operator
  wording is corrected above. Closure records and this wording correction
  receive an independent impact check before the public commit/submission.
- False-positive placeholder finding: RELEASING.md invokes `check-inputs`
  before any tag creation. The helper validates the exact version/SHA format
  and rejects the literal placeholders. No additional guard is needed.
- Final lint evidence: fresh actionlint, Ruff check, Ruff format check and
  whitespace checks all return zero. The retained structured receipt records
  the command, timestamp, output path and reviewed input hashes; silent
  successful actionlint output is expected, not evidence of an absent run.
- Remaining P3 notes are explicitly rejected as unnecessary for the approved
  outcome: missing Go origin metadata is already documented as optional while
  exact module/version and canonical tags are mandatory; unexpected null/JSON
  shapes still fail with a nonzero exit; upload failures keep Cargo stderr and
  the explicit registry-check message; first-read failures stop safely and
  same-input reruns recover. Preserve package verification after the dry-run:
  the publish command has a 900-second limit below the temporary permission
  lifetime, with the warmed packaging target reused.
- Future install-snippet forms, archives larger than the current 16 MiB cap,
  cosmetic action-version comments and automatic scratch removal add scope
  without an established current failure. Existing active examples and real
  archives are verified; ignored scratch retains diagnostic evidence and CI
  runners are temporary. No such feature is deferred by this SOW.
- The unchanged manual fallback's origin/scan-root omissions are separate
  pre-existing notes, rejected from this CI-path scope. The delivered guide
  uses the explicit canonical remote and the operative scanner includes the
  Go README. The earlier Go stress timeout and PR #6 preparation review-record
  gap retain their recorded dispositions; the latter remains an explicit
  prerequisite in pending SOW-0151, not silently accepted here.

Sensitive data gate:

- Reviewed the changed durable artifacts. They contain public package names,
  versions, source hashes, permission variable names and clearly synthetic
  test values, with no raw credentials or personal/customer identifiers.
- Real permission values are not requested or stored. Their command/environment
  handling is covered by the helper tests; reviewer outputs stay in scratch.

### Artifact Maintenance Gate

- AGENTS.md: unchanged because this implements its existing release, parity,
  SOW and compiler policies; no new project-wide responsibility is needed.
- Runtime project skills: release-tagging skill updated for Rust CI,
  one-time configuration, immutable reruns and existing-access maintainer
  tag/verification handoff; manual fallback remains.
- Specs: product-scope now records the implemented paired release contract.
- End-user/operator docs: new root RELEASING.md covers setup, source selection,
  execution and partial-release recovery; README links it. SDK usage and wiki
  examples are unchanged, so no consumer wiki edits are required.
- End-user/operator skills: none are distributed by this project; the operator
  procedure is documented directly instead of introducing an unused skill.
- SOW lifecycle: this implementation is completed in done/ with code and the
  lifecycle change in one public commit. Account setup and actual publication
  are represented by PR #6's real pending SOW-0151. The current directory
  retains its committed .gitkeep; final closure is audited before submission.
- SOW-status.md: records completion together with the SOW move. Other pending
  work is unchanged.
- Lessons: recorded below. Follow-up mapping has a concrete execution SOW;
  no untracked SDK changes or additional release features are deferred here.

## Outcome

Delivered the reviewed maintainer-handoff implementation. CI validates both
languages and publishes Rust with Trusted Publishing; permitted maintainers
push paired tags with existing access, then verify matching Rust sources and
exact Go consumption. Local executable evidence and independent final review
pass. Four external reports are positive; two missing reports are disclosed
coverage gaps. SOW completion and code are included together in the separate
PR. No version publication, public tag or account-setting change is performed
by this implementation SOW; actual execution remains in pending SOW-0151.

## Lessons Extracted

- Registry upload command failure is not proof that publication failed. Read
  both registry views and verify the immutable archive before resuming.
- A successful audit exit code is insufficient; inspect the audit's verdict
  and record/repair missing declarations before closure.
- Keep a release's source SHA separate from the workflow's dispatch SHA. This
  permits release preparation to precede automation merge and makes reruns
  stable while master advances; the guide/spec/skill preserve that contract.
- Record transient test failures and successful isolated/full retries without
  inventing a root cause or expanding the PR into unrelated SDK changes.
- Local native Git tests establish atomic tag behavior but do not exercise
  GitHub's historical-workflow permission checks. A permission method must
  support the full recovery contract, not only the currently retained branches.
- Confirm the requested division of responsibility before automating an
  account operation. CI crate publication with a source receipt and normal
  maintainer tag pushes delivers the clarified requirement without another
  account credential or automation identity.

## Followup

Actual 0.9.0 Trusted Publishing setup, Rust CI publication, maintainer tag
pushes and final verification remain with pending SOW-0151, merged through
PR #6 and aligned here with the approved procedure. This SOW supplies the
workflow and operator procedure.

## Regression Log

No regression identified.

Append dated regression sections after completion only if shipped behavior
later contradicts this SOW's outcome.
