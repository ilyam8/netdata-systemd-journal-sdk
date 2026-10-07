# SOW-0154 - Project Release Skill

## Status

Status: completed

Sub-state: Full-process skill implemented, locally validated and independently reviewed; prepared for PR #8 with review gaps disclosed.

## Requirements

### Purpose

Give the team a discoverable project skill for the complete existing Rust/Go release process.

### User Request

Update PR #8 with instructions explaining releases; the user selected creating a `project-release` skill as the better delivery format.

### Assistant Understanding

Facts:

- `RELEASING.md` already documents the approved manual CI dispatch, Rust publication, maintainer tag handoff, verification and recovery.
- `project-release-tagging` covers tag/package details but there is no full-process `project-release` entry point.
- The user requires language parity: accepted Go features also belong in Rust and vice versa.
- PR #8 records the completed paired `0.9.0` publication. This request adds instructions to that PR.

Inferences:

- Preserve the approved process and its existing detailed guide; add routing and release acceptance instructions rather than redesigning automation.

Unknowns:

- No unresolved product decision affects this release-guidance change. A future version choice or publication request is outside this SOW.

### Acceptance Criteria

- A valid `.agents/skills/project-release/SKILL.md` routes preparation, publication, tagging, verification, recovery and close-out.
- The instructions preserve language parity, immutable source selection, CI/maintainer responsibilities and same-input recovery.
- The skill is registered in `AGENTS.md` and linked from the operator guide and focused tagging skill.
- PR #8 includes the skill and a link explaining how the team uses it.
- Local validation, independent review and the complete-and-clean SOW audit pass; the completed SOW and work are committed together.

## Analysis

Sources checked:

- `AGENTS.md`, `.agents/skills/project-agent-orchestration/SKILL.md` and `.agents/sow/SOW.template.md`.
- `RELEASING.md`, `.agents/skills/project-release-tagging/SKILL.md` and `.agents/sow/specs/product-scope.md`.
- `.github/workflows/release.yml`, `tests/release/release.py` and `tests/release/test_release.py`.
- Completed SOW-0150/0151/0152, pending SOW-0153 and SOW-0066, current/pending inventories and both status ledgers.
- System-wide skill-creator and external-reviewers guidance, read as authoring/review inputs only.

State before activation (2026-10-07):

- No implementation SOW was current. The worktree was clean on PR #8's branch at `b63b306269d36fd5d88e2847e2df031fd7db0467`.
- The eight Rust `0.9.0` crates and both signed annotated tags were verified from `fba9d45d53044d54163da53c903c7eedbde6e34f`; this work does not publish or retag them.
- SOW-0153's consumer prose changes and SOW-0066's future stable release remain separate, unactivated work.

Risks:

- Incorrect routing could send a maintainer to tag before Rust CI succeeds or use a newer master commit as release source.
- Duplicating executable commands could drift from the tested helper and operator guide.
- A skill must preserve the user's actual authorization; creating instructions does not authorize a new publication.

## Pre-Implementation Gate

Status: ready

Problem / root-cause model:

- The detailed operator guide and focused tagging skill exist, but the requested full-process project skill is missing from the runtime skill catalog.

Evidence reviewed:

- `RELEASING.md` and `product-scope.md` agree with the workflow's separate workflow/source checkouts, eight-crate sequence and human tag handoff.
- The release helper and its recovery tests enforce exact source matching, annotated paired tags, same-input resumption and consumer verification.
- Current/pending SOW inventories show no competing release-instructions implementation.

Affected contracts and surfaces:

- Runtime skill discovery, maintainer release guidance, SOW tracking and PR #8's description.
- The changed-artifact audit must distinguish the guide's canonical public Git SSH remote from an email address, while preserving email detection elsewhere.
- SDK behavior, workflow code, version metadata and existing published artifacts are unchanged.

Existing patterns to reuse:

- Repository-local `project-*` skills with YAML name/description, relative links and explicit registration in `AGENTS.md`.
- The existing operator guide for commands/setup and focused tagging skill for package/tag details.
- Existing SOW template, audit and explicit-path commit practice.

Risk and blast radius:

- Instructions can influence future externally visible releases; compare every procedural invariant to the tested implementation and evaluate representative interrupted-release scenarios.
- There is no SDK runtime, compiler-minimum, journal-format or performance change.

Sensitive data handling plan:

- Record only public repository/package identifiers, relative paths, sanitized review summaries and public release receipts. Keep raw reports and validation logs under ignored `.local/` or the review runner's private temporary directory.

Implementation plan:

1. Add the full-process skill using existing approved mechanics, including parity and publication authorization boundaries.
2. Register and link it; validate links, skill structure, representative decisions and source agreement.
   Repair any audit false match exposed by the changed guide with a scoped rule and direct classifier checks.
3. Review the whole SOW, close it with both ledgers, commit explicit paths and push to PR #8; update the PR description with the team entry point.

Validation plan:

- Run the skill-creator validator and `git diff --check`; resolve every new relative Markdown target.
- Independently evaluate source advancement, partial Rust publication, existing matching/conflicting tags and an instructions-only request.
- Compare the skill with the actual release workflow/helper and prior successful `0.9.0` release receipts; do not trigger another publication to validate prose.
- Require the audit's explicit complete-and-clean verdict, not only its exit code.

Artifact impact plan:

- AGENTS.md: register the full-process release skill.
- Runtime project skills: add `project-release`; route to it from `project-release-tagging`.
- Local audit: exclude only the canonical public SSH remote from email matching and verify other email matches remain detectable.
- Specs: no behavior change; the current Paired CI Releases contract already describes this process.
- End-user/operator docs: add a skill entry-point link to `RELEASING.md` and PR #8.
- End-user/operator skills: no separate exported skill exists; the requested team entry point is the repository's runtime skill.
- SOW lifecycle: complete this SOW and move it to `done/` with the instructions commit; keep unrelated pending work parked.
- SOW-status.md: update the canonical ledger and root convenience index at activation and completion.

Open-source reference evidence:

- No external source checkout is needed; the repository's approved guide, tested workflow and successful publication receipts are the authority for this change.

Open decisions:

- Resolved by the user's request: create the full-process `project-release` skill and include it in PR #8. No automation or release-policy redesign is proposed.

## Implications And Decisions

1. User selection: create `project-release` rather than only adding release prose to PR #8. Implement as minimal-complete long-term maintenance of the already approved process.
2. Reuse `RELEASING.md` for executable commands and setup, and preserve `project-release-tagging` for focused details. This avoids deleting an existing entry point or maintaining competing command recipes.
3. Language parity is a release-preparation gate. A discovered mismatch requires the user's scope/version decision before publication.

## Plan

1. Author and register the release entry point without changing release mechanics.
2. Validate the entire instructional surface and review it independently.
3. Close tracking and deliver the change through the existing PR #8.

## Implementation And Review Plan

Implementation:

- One author edits the skill and routing links sequentially, preserving all existing release artifacts and unrelated work.
- Capture local validation and review evidence before committing.

Reviewers:

- Standing user authorization for external review applies to the current conversation. Use the system-wide external-reviewers skill at the complete, locally validated SOW boundary.
- A read-only independent forward-test evaluates realistic use of the new release instructions; no publication, tag mutation or package commands are permitted.

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

- Verify every finding against the actual procedure before editing. Stop for unresolved design/authorization conflicts; repair verified in-scope instructional defects, revalidate and disclose any reviewer coverage gap.

## Execution Log

### 2026-10-07

- Completed preliminary analysis, recorded the user's selection and activated this SOW before authoring the new skill.
- The first changed-artifact audit exited 2: the guide's two existing canonical Git SSH commands matched the email rule. Both hits identify the public repository remote, not personal data. Extend scope only to excluding that exact SSH address from email matching, preserving all other rules and checking mixed-line detection.
- Added the skill, catalog registration and operator/focused-skill links. The structure validator, all six relative links/anchors and whitespace checks pass.
- Executed seven cases through the actual audit classifier; the canonical fetch/push address is excluded while ordinary, bare GitHub, mixed-line, different-repository and suffixed-address cases remain reported. `bash -n` and the changed-artifact full audit pass with the explicit complete-and-clean verdict.
- Captured PR #8's existing head: all 27 checks succeeded at `b63b306269d36fd5d88e2847e2df031fd7db0467` before submission of this change.
- Independent forward-testing found the release scenarios correct and a historical-state wording issue, corrected by labeling the pre-activation snapshot explicitly.
- A boundary probe showed the initial audit exception also matched after dot/hyphen email-local-part prefixes. Stopped the first external batch before any report returned, tightened the left boundary against all original email-local-part characters and repeated local validation before a new whole-SOW review. The first batch's early provider failure was HTTP 429; five unfinished workers were intentionally stopped, not accepted as review coverage.
- Completed external batch `8b684dccd14246a49bb36d64744e726e`: one positive report, four execution timeouts and one repeated HTTP 429 failure. No verdict is claimed for unavailable reviewers.
- Handled the report's two actionable P3 notes: made the operator invocation hint tool-neutral and constrained the audit exception to whole whitespace-delimited canonical remote tokens. Expanded actual-classifier validation to 43 cases, including every ASCII punctuation prefix, adjacent email and conservative quoted-token behavior.
- Closed this SOW with the skill, routing, audit correction and both ledgers together; the existing PR #8 is the delivery vehicle.
- The closing changed-artifact scan flagged a validation sentence using the word pass before a colon as a password assignment. Rephrased that evidence sentence without changing the classifier, then reran the full audit before commit.

## Validation

Acceptance criteria evidence:

- The full-process skill, catalog registration and both routing links are implemented. Skill coverage agrees with the approved guide, workflow/helper and current release spec.
- The guide and prepared PR #8 description expose the entry point and explain the team handoff. No new version, tag or package publication was performed for this SOW.

Tests or equivalent validation:

- Skill-creator `quick_validate.py` reports `Skill is valid!`; all six relative links and the spec heading resolve.
- `git diff --check` and `bash -n .agents/sow/audit.sh` pass.
- All 43 final real-classifier scenarios pass. Canonical command tokens are excluded, all ASCII punctuation/alphanumeric prefix probes remain detectable, adjacent email is retained, and quoted canonical tokens remain conservatively reported.
- `SOW_AUDIT_SENSITIVE_CHANGED=1 bash .agents/sow/audit.sh` exits 0 and explicitly reports `SOW initialization complete and clean` after the scoped false-match repair.

Real-use evidence:

- Existing release run `37607475693` and completed SOW-0151 provide actual execution evidence for the unchanged procedure.
- Independent forward-testing passed five realistic requests: missing Rust parity, master advancement after publication, an accepted prefix with ambiguous upload timeout, matching/conflicting existing tags, and an instructions-only PR request. Every new local link/anchor resolves.
- The changed audit procedure was executed directly against representative public-command and synthetic-email inputs; no live release action was needed to validate instructions.

Reviewer findings:

- Independent forward-test: PASS. Corrected its two notes about the historical analysis snapshot and the initial audit left boundary; the reviewer verified both corrections.
- External report: PRODUCTION GRADE, with three non-blocking P3 notes. The full-process instructions were consistent with the actual guide, workflow, helper, tests and spec.
- `cli_specific_invocation_hint`: accepted; operator/PR guidance now says to ask for the skill without prescribing one tool's syntax.
- `audit_exemption_attext_prefix`: accepted; the exception now requires whitespace or line start before the canonical token, preserving detection with any attached non-whitespace prefix. Actual-classifier checks cover all ASCII punctuation prefixes.
- `audit_exemption_formatting_brittle`: rejected as an expansion unnecessary for this change. The affected guide's actual shell commands use unquoted, whitespace-delimited remotes; the final audit is clean. Broader Markdown-delimiter exemptions are not needed for those commands and could widen matching. Quoted-token reporting is explicitly checked as conservative behavior.
- Unrelated legacy fallback `origin`/non-atomic push guidance: rejected from this SOW, as already dispositioned in SOW-0152. The new paired workflow routes to the guide's explicit canonical atomic procedure.
- Unrelated opt-in full-history scanning of ignored scratch: rejected from this SOW. It predates the change and does not affect the required changed-artifact audit; a broader audit cleanup is not needed to deliver the requested skill.
- Coverage: one returned external report; four workers timed out at approximately 601 seconds with no reports; one provider failed with HTTP 429 at approximately 64 seconds after its single retry. Retained the returned report and disclosed all gaps. The first, superseded batch has no accepted reports. Raw session reports remain private; no verdict is invented for a failed or stopped worker.

Same-failure scan:

- Checked the operator guide, focused tagging skill, runtime catalog and release spec for inconsistent handoffs, source selection and incomplete verification. The existing process is consistent; the new skill adds the missing complete-process entry point.
- The audit repair excludes only the exact whitespace-delimited canonical SSH repository address from the email-check copy; it does not exempt entire lines or unrelated addresses, and every other classifier still sees the original line.

Sensitive data gate:

- Durable changes contain public identifiers and relative references only. The changed-artifact scan passes; raw evidence stays under ignored `.local/`.

Artifact maintenance gate:

- AGENTS.md: registered `.agents/skills/project-release/SKILL.md`.
- Runtime project skills: added the full-process entry point and routing in `project-release-tagging`.
- Specs: current release behavior already documented; no new contract is introduced.
- End-user/operator docs: `RELEASING.md` links the new skill; PR #8's prepared description links it and explains the four-step release process.
- End-user/operator skills: no separate exported skill exists.
- SOW lifecycle: completed in `done/`; the status/move and implementation are prepared for one commit together. SOW-0153 and SOW-0066 remain pending and unactivated.
- SOW-status.md: both ledgers record completion and no active implementation SOW.
- Local audit: exact public Git SSH remote no longer produces an email false match; actual classifier checks preserve other detection.

Specs update:

- No update is needed: `product-scope.md` already describes all unchanged release mechanics.

Project skills update:

- New full-process skill and focused-skill routing are the deliverable.

End-user/operator docs update:

- Added the guide link and prepared the PR #8 team handoff. Executable setup/tag/verification/recovery commands remain in the existing guide and helper.

End-user/operator skills update:

- The repository-local skill is the requested team entry point; no exported skill copies need synchronization.

Lessons:

- Keep full-process routing separate from executable command detail to reduce instructional drift.
- A public SSH address needs a whole-token exception, not a substring exclusion; validate adjacent addresses and punctuation prefixes through the actual classifier.

Follow-up mapping:

- Implemented the requested skill, discoverability, routing and necessary changed-artifact audit repair.
- Rejected broader delimiter handling, legacy fallback cleanup and opt-in full-history scratch scanning changes with scope/evidence above; these are not deferred requirements.
- SOW-0153 consumer prose and SOW-0066 stable-release gates remain separately tracked and unactivated. No new deferred scope is introduced.

## Outcome

The repository has a complete `project-release` entry point for the team's existing paired CI/maintainer process. Language parity, authorization, immutable source selection, verified Rust publication, annotated paired tags, exact Go consumption and recovery are explicit. The skill and links are prepared for PR #8 with completed tracking and truthful review coverage.

## Lessons Extracted

The requested entry point should reuse the existing guide and tested helper instead of creating another release implementation.

Keep exceptions for public command addresses bounded to whole tokens so a false-match repair does not suppress neighboring data. Both invocation wording and validation must support the repository's multiple-tool setup.

## Followup

Existing unrelated SOW-0153 and SOW-0066 remain pending. This SOW adds no deferred implementation.

## Regression Log

No regression is identified; this is a new instructional entry point.
