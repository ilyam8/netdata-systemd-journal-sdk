"""Exercise release failures using controlled archives and repository-owned Git."""

import hashlib
import io
import json
import os
import subprocess
import tarfile
import tempfile
import unittest
import urllib.error
from pathlib import Path
from unittest import mock

import release

VERSION = "0.9.0"
SHA = "a" * 40
OTHER_SHA = "b" * 40


def archive(name, *, commit=SHA, dirty=False, version=VERSION):
    buffer = io.BytesIO()
    files = {
        ".cargo_vcs_info.json": json.dumps(
            {"git": {"sha1": commit, "dirty": dirty}}
        ).encode(),
        "Cargo.toml": f'[package]\nname = "{name}"\nversion = "{version}"\n'.encode(),
    }
    with tarfile.open(fileobj=buffer, mode="w:gz") as stream:
        for filename, data in files.items():
            info = tarfile.TarInfo(f"{name}-{VERSION}/{filename}")
            info.size = len(data)
            stream.addfile(info, io.BytesIO(data))
    return buffer.getvalue()


class MemoryRegistry(release.Registry):
    def __init__(self):
        self.uploads = {}

    def add(self, name, *, commit=SHA):
        self.uploads[name] = archive(name, commit=commit)

    def probe(self, name, version):
        if name not in self.uploads:
            return None, None
        checksum = hashlib.sha256(self.uploads[name]).hexdigest()
        return (
            {"num": version, "yanked": False, "checksum": checksum},
            {"name": name, "vers": version, "yanked": False, "cksum": checksum},
        )

    def verify(self, name, version, commit, api, index):
        release.verify_archive(
            self.uploads[name], index["cksum"], name, version, commit
        )


def test_directory(prefix):
    parent = release.ROOT / ".local/release-tests"
    parent.mkdir(parents=True, exist_ok=True)
    # Keep task-owned fixtures for inspection; never clean another session's files.
    return Path(tempfile.mkdtemp(prefix=prefix, dir=parent))


class InputAndManifestTests(unittest.TestCase):
    def test_accepts_stable_version_and_immutable_source(self):
        release.validate_inputs(VERSION, SHA)

    def test_rejects_shell_fragments_short_sha_and_ambiguous_versions(self):
        for version, commit in [
            ("v0.9.0", SHA),
            ("0.9.0-rc.1", SHA),
            ("00.9.0", SHA),
            ("0.9.0\n1.0.0", SHA),
            ("$(exit 0)", SHA),
            (VERSION, "master"),
            (VERSION, SHA[:12]),
            ("2.0.0", SHA),
        ]:
            with (
                self.subTest(version=version, commit=commit),
                self.assertRaises(release.ReleaseError),
            ):
                release.validate_inputs(version, commit)

    def test_publication_requires_canonical_master_manual_dispatch(self):
        valid = {
            "GITHUB_ACTIONS": "true",
            "GITHUB_REPOSITORY": release.REPOSITORY,
            "GITHUB_REF": "refs/heads/master",
            "GITHUB_EVENT_NAME": "workflow_dispatch",
        }
        with mock.patch.dict(os.environ, valid, clear=True):
            release.require_ci()
            for key, value in [
                ("GITHUB_REPOSITORY", "example/fork"),
                ("GITHUB_REF", "refs/heads/release"),
                ("GITHUB_EVENT_NAME", "pull_request"),
            ]:
                with (
                    mock.patch.dict(os.environ, {key: value}),
                    self.assertRaises(release.ReleaseError),
                ):
                    release.require_ci()

    def test_manual_release_is_independent_of_dispatching_operator(self):
        valid = {
            "GITHUB_ACTIONS": "true",
            "GITHUB_REPOSITORY": release.REPOSITORY,
            "GITHUB_REF": "refs/heads/master",
            "GITHUB_EVENT_NAME": "workflow_dispatch",
        }
        for actor in ("maintainer-a", "maintainer-b"):
            with (
                self.subTest(actor=actor),
                mock.patch.dict(
                    os.environ, {**valid, "GITHUB_ACTOR": actor}, clear=True
                ),
                mock.patch.object(release, "Registry") as registry,
            ):
                self.assertEqual(
                    release.main(["inputs", "--version", VERSION, "--commit", SHA]),
                    0,
                )
                registry.assert_not_called()

    def test_environment_requires_exact_master_branch_policy(self):
        environment = {
            "deployment_branch_policy": {
                "protected_branches": False,
                "custom_branch_policies": True,
            }
        }
        rules = {"branch_policies": [{"name": "master", "type": "branch"}]}
        release.validate_environment(environment, rules)
        for bad in [
            {"branch_policies": []},
            {"branch_policies": [{"name": "*", "type": "branch"}]},
            {"branch_policies": [{"name": "master", "type": "tag"}]},
            {"branch_policies": rules["branch_policies"] * 2},
        ]:
            with self.subTest(bad=bad), self.assertRaises(release.ReleaseError):
                release.validate_environment(environment, bad)
        with self.assertRaises(release.ReleaseError):
            release.validate_environment({"deployment_branch_policy": None}, rules)

    def test_rejects_missing_setup_without_creating_environment(self):
        with mock.patch.object(release, "http_get", return_value=None) as request:
            with self.assertRaisesRegex(
                release.ReleaseError, "Create the release environment"
            ):
                release.check_environment()
            self.assertEqual(request.call_count, 1)

    def test_metadata_rejects_unpublished_dependencies_and_wrong_pins(self):
        packages = [
            {
                "id": name,
                "name": name,
                "version": VERSION,
                "publish": None,
                "dependencies": [],
            }
            for name in release.PACKAGES
        ]
        metadata = {"workspace_members": list(release.PACKAGES), "packages": packages}
        release.validate_metadata(metadata, VERSION)
        packages[-1]["dependencies"] = [
            {
                "name": release.PACKAGES[0],
                "path": "a/path",
                "req": "^0.8.2",
                "kind": None,
            }
        ]
        with self.assertRaisesRegex(release.ReleaseError, "version pin"):
            release.validate_metadata(metadata, VERSION)
        packages[-1]["dependencies"][0].update(name="private-tool", req=f"^{VERSION}")
        with self.assertRaisesRegex(release.ReleaseError, "unpublished"):
            release.validate_metadata(metadata, VERSION)

    def test_metadata_detects_publication_order_drift(self):
        packages = [
            {
                "id": name,
                "name": name,
                "version": VERSION,
                "publish": None,
                "dependencies": [],
            }
            for name in release.PACKAGES
        ]
        packages[0]["dependencies"] = [
            {
                "name": release.PACKAGES[-1],
                "path": "sdk",
                "req": f"^{VERSION}",
                "kind": None,
            }
        ]
        with self.assertRaisesRegex(release.ReleaseError, "order"):
            release.validate_metadata(
                {"workspace_members": list(release.PACKAGES), "packages": packages},
                VERSION,
            )

    def test_wrong_registry_is_rejected_before_package_publication(self):
        packages = [
            {
                "id": name,
                "name": name,
                "version": VERSION,
                "publish": None,
                "dependencies": [],
            }
            for name in release.PACKAGES
        ]
        metadata = {"workspace_members": list(release.PACKAGES), "packages": packages}
        packages[-1]["publish"] = ["private-registry"]
        registry = mock.Mock()
        with (
            mock.patch.object(release, "validate_checkout"),
            mock.patch.object(release, "manifest"),
            mock.patch.object(
                release, "run", return_value=json.dumps(metadata)
            ) as command,
            self.assertRaisesRegex(release.ReleaseError, "crates.io publication"),
        ):
            release.preflight(release.ROOT, VERSION, SHA, registry)
        self.assertEqual(command.call_count, 1)
        self.assertIn("metadata", command.call_args.args[0])
        registry.existing.assert_not_called()
        packages[-1]["publish"] = ["crates-io"]
        release.validate_metadata(metadata, VERSION)

    def test_install_scan_includes_lower_level_rust_crates(self):
        source = test_directory("docs-")
        for name in ("README.md", "go/README.md", "rust/README.md"):
            path = source / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text("")
        (source / "rust/README.md").write_text(
            'core = { package = "systemd-journal-sdk-core", version = "0.8.2" }'
        )
        with self.assertRaisesRegex(release.ReleaseError, "Stale install"):
            release.validate_install_examples(source, VERSION)


class RegistryTests(unittest.TestCase):
    def test_checks_archive_checksum_source_identity_and_dirty_flag(self):
        name = release.PACKAGES[0]
        valid = archive(name)
        release.verify_archive(
            valid, hashlib.sha256(valid).hexdigest(), name, VERSION, SHA
        )
        for data, checksum, source in [
            (valid, "0" * 64, SHA),
            (valid, hashlib.sha256(valid).hexdigest(), OTHER_SHA),
            (archive(name, dirty=True), None, SHA),
            (archive(name, version="0.8.2"), None, SHA),
        ]:
            with (
                self.subTest(source=source, checksum=checksum),
                self.assertRaises(release.ReleaseError),
            ):
                release.verify_archive(
                    data,
                    checksum or hashlib.sha256(data).hexdigest(),
                    name,
                    VERSION,
                    source,
                )

    def test_rejects_yanked_or_conflicting_registry_metadata(self):
        registry = release.Registry()
        api = {"num": VERSION, "yanked": True, "checksum": "0" * 64}
        index = {
            "name": release.PACKAGES[0],
            "vers": VERSION,
            "yanked": False,
            "cksum": "0" * 64,
        }
        with self.assertRaisesRegex(release.ReleaseError, "yanked"):
            registry.verify(release.PACKAGES[0], VERSION, SHA, api, index)
        api.update(yanked=False, checksum="1" * 64)
        with self.assertRaises(release.RegistryUnavailable):
            registry.verify(release.PACKAGES[0], VERSION, SHA, api, index)

    def test_api_visible_index_pending_is_existing_never_a_second_upload(self):
        registry = release.Registry()
        api, index = {"num": VERSION}, {"vers": VERSION}
        with (
            mock.patch.object(
                registry, "probe", side_effect=[(api, None), (api, None), (api, index)]
            ),
            mock.patch.object(registry, "verify") as verify,
            mock.patch.object(release.time, "sleep"),
        ):
            self.assertTrue(registry.existing(release.PACKAGES[0], VERSION, SHA))
            verify.assert_called_once()

    def test_index_visible_api_pending_is_existing(self):
        registry = release.Registry()
        api, index = {"num": VERSION}, {"vers": VERSION}
        with (
            mock.patch.object(
                registry, "probe", side_effect=[(None, index), (api, index)]
            ),
            mock.patch.object(registry, "verify"),
            mock.patch.object(release.time, "sleep"),
        ):
            self.assertTrue(registry.existing(release.PACKAGES[0], VERSION, SHA))

    def test_http_failure_is_not_absence(self):
        error = urllib.error.HTTPError(
            "https://example.invalid", 503, "unavailable", {}, None
        )
        with (
            mock.patch.object(release.urllib.request, "urlopen", side_effect=error),
            self.assertRaises(release.RegistryUnavailable),
        ):
            release.http_get("https://example.invalid", missing_ok=True)

    def test_wait_retries_transport_but_stops_at_deadline(self):
        registry = release.Registry()
        with (
            mock.patch.object(
                registry,
                "probe",
                side_effect=release.RegistryUnavailable("unavailable"),
            ),
            mock.patch.object(
                release.time, "monotonic", side_effect=[0, release.WAIT_SECONDS]
            ),
            self.assertRaisesRegex(release.ReleaseError, "not fully available"),
        ):
            registry.wait(release.PACKAGES[0], VERSION, SHA)


class PublicationTests(unittest.TestCase):
    def test_partial_batch_resumes_without_reuploading_verified_prefix(self):
        registry = MemoryRegistry()
        events = []
        first_attempt = True

        def cargo(args, **_kwargs):
            name = args[args.index("--package") + 1]
            if "--dry-run" in args:
                events.append(("dry-run", name))
                if name == release.PACKAGES[2] and first_attempt:
                    raise release.ReleaseError("controlled package build failure")
            else:
                events.append(("upload", name))
                registry.add(name)
            return ""

        source = release.ROOT
        with (
            mock.patch.object(release, "validate_checkout"),
            mock.patch.object(release, "output"),
            mock.patch.object(release, "run", side_effect=cargo),
        ):
            for name in release.PACKAGES[:2]:
                self.assertTrue(release.prepare(source, VERSION, SHA, name, registry))
                release.publish(source, VERSION, SHA, name, registry)
            with self.assertRaisesRegex(release.ReleaseError, "build failure"):
                release.prepare(source, VERSION, SHA, release.PACKAGES[2], registry)
            self.assertNotIn(release.PACKAGES[2], registry.uploads)
            first_attempt = False
            for name in release.PACKAGES:
                if release.prepare(source, VERSION, SHA, name, registry):
                    release.publish(source, VERSION, SHA, name, registry)
            self.assertEqual(
                [name for kind, name in events if kind == "upload"],
                list(release.PACKAGES),
            )
            self.assertEqual(set(registry.uploads), set(release.PACKAGES))

    def test_upload_command_timeout_after_acceptance_is_verified(self):
        registry = MemoryRegistry()

        def uploaded_then_failed(*_args, **_kwargs):
            registry.add(release.PACKAGES[0])
            raise release.ReleaseError("controlled post-upload timeout")

        with (
            mock.patch.object(release, "validate_checkout"),
            mock.patch.object(release, "run", side_effect=uploaded_then_failed),
        ):
            release.publish(release.ROOT, VERSION, SHA, release.PACKAGES[0], registry)
        self.assertIn(release.PACKAGES[0], registry.uploads)

    def test_wrong_source_existing_crate_blocks_resumption(self):
        registry = MemoryRegistry()
        registry.add(release.PACKAGES[0], commit=OTHER_SHA)
        with (
            mock.patch.object(release, "validate_checkout"),
            mock.patch.object(release, "run") as command,
            self.assertRaisesRegex(release.ReleaseError, "different or dirty"),
        ):
            release.prepare(release.ROOT, VERSION, SHA, release.PACKAGES[0], registry)
        command.assert_not_called()

    def test_failed_upload_with_no_accepted_version_stops(self):
        registry = MemoryRegistry()
        with (
            mock.patch.object(release, "validate_checkout"),
            mock.patch.object(
                release, "run", side_effect=release.ReleaseError("upload rejected")
            ),
            mock.patch.object(
                release.time, "monotonic", side_effect=[0, release.WAIT_SECONDS]
            ),
            self.assertRaisesRegex(release.ReleaseError, "not fully available"),
        ):
            release.publish(release.ROOT, VERSION, SHA, release.PACKAGES[0], registry)

    def test_command_reports_original_exit_status(self):
        with self.assertRaisesRegex(release.ReleaseError, "exited 17"):
            release.run(
                [os.sys.executable, "-c", "import sys; sys.exit(17)"], cwd=release.ROOT
            )

    def test_rust_consumer_resolves_exact_registry_version_without_path_override(self):
        registry = MemoryRegistry()
        for name in release.PACKAGES:
            registry.add(name)
        root = test_directory("rust-consumer-")
        with (
            mock.patch.object(release, "ROOT", root),
            mock.patch.object(release, "run") as command,
        ):
            release.verify_rust(root, VERSION, SHA, registry)
        arguments = command.call_args.args[0]
        manifest = Path(arguments[arguments.index("--manifest-path") + 1]).read_text()
        self.assertIn('version = "=0.9.0"', manifest)
        self.assertNotIn("path =", manifest)

    def test_go_consumer_downloads_and_builds_exact_registry_version(self):
        source = release.ROOT
        root = test_directory("go-consumer-")
        lookup = json.dumps(
            {
                "Path": release.GO_MODULE,
                "Version": f"v{VERSION}",
                "Origin": {"Hash": SHA},
            }
        )
        with (
            mock.patch.object(release, "ROOT", root),
            mock.patch.object(
                release,
                "run",
                side_effect=[
                    f"module {release.GO_MODULE}\n\ngo 1.25.0\n",
                    lookup,
                    "",
                    "",
                    lookup,
                    "",
                ],
            ) as command,
        ):
            release.verify_go(source, VERSION, SHA)
        calls = command.call_args_list
        self.assertEqual(calls[0].args[0], ["git", "show", f"{SHA}:go/go.mod"])
        self.assertIn(f"{release.GO_MODULE}@v{VERSION}", calls[2].args[0])
        directory = calls[-1].kwargs["cwd"]
        self.assertIn(
            f"require {release.GO_MODULE} v{VERSION}",
            (directory / "go.mod").read_text(),
        )
        self.assertIn("go 1.25.0\n", (directory / "go.mod").read_text())
        self.assertIn("build", calls[-1].args[0])

    def test_go_source_conflict_stops_without_wait_or_consumer_build(self):
        lookup = json.dumps(
            {
                "Path": release.GO_MODULE,
                "Version": f"v{VERSION}",
                "Origin": {"Hash": OTHER_SHA},
            }
        )
        with (
            mock.patch.object(
                release,
                "run",
                side_effect=[
                    f"module {release.GO_MODULE}\n\ngo 1.26.2\n",
                    lookup,
                ],
            ) as command,
            mock.patch.object(release.time, "sleep") as sleep,
            self.assertRaisesRegex(release.ReleaseError, "different source"),
        ):
            release.verify_go(release.ROOT, VERSION, SHA)
        self.assertEqual(command.call_count, 2)
        sleep.assert_not_called()

    def test_rust_success_summary_hands_tags_to_a_maintainer(self):
        summary = test_directory("summary-") / "summary.md"
        valid = {
            "GITHUB_ACTIONS": "true",
            "GITHUB_REPOSITORY": release.REPOSITORY,
            "GITHUB_REF": "refs/heads/master",
            "GITHUB_EVENT_NAME": "workflow_dispatch",
            "GITHUB_STEP_SUMMARY": str(summary),
        }
        with (
            mock.patch.dict(os.environ, valid, clear=True),
            mock.patch.object(release, "verify_rust"),
        ):
            self.assertEqual(
                release.main(["verify-rust", "--version", VERSION, "--commit", SHA]),
                0,
            )
        text = summary.read_text()
        self.assertIn("Rust publication verified", text)
        self.assertIn(SHA, text)
        self.assertIn("maintainer pushes", text)
        self.assertIn(f"`go/v{VERSION}`", text)
        self.assertNotIn("Go lookup verified", text)

    def test_failed_rust_verification_does_not_write_success_summary(self):
        summary = test_directory("summary-failure-") / "summary.md"
        valid = {
            "GITHUB_ACTIONS": "true",
            "GITHUB_REPOSITORY": release.REPOSITORY,
            "GITHUB_REF": "refs/heads/master",
            "GITHUB_EVENT_NAME": "workflow_dispatch",
            "GITHUB_STEP_SUMMARY": str(summary),
        }
        with (
            mock.patch.dict(os.environ, valid, clear=True),
            mock.patch.object(
                release, "verify_rust", side_effect=release.ReleaseError("not verified")
            ),
            mock.patch("sys.stderr", new_callable=io.StringIO),
        ):
            self.assertEqual(
                release.main(["verify-rust", "--version", VERSION, "--commit", SHA]),
                1,
            )
        self.assertFalse(summary.exists())


class GitIntegrationTests(unittest.TestCase):
    def setUp(self):
        self.directory = test_directory("git-")
        self.remote = self.directory / "remote.git"
        self.source = self.directory / "source"
        self.git(
            self.directory,
            "init",
            "--bare",
            "--initial-branch=master",
            str(self.remote),
        )
        self.git(self.directory, "init", "--initial-branch=master", str(self.source))
        (self.source / "README.md").write_text("controlled release fixture\n")
        self.git(self.source, "add", "README.md")
        self.git(self.source, "commit", "-m", "Create controlled release fixture")
        self.commit = self.git(self.source, "rev-parse", "HEAD").strip()
        self.git(self.source, "remote", "add", "origin", str(self.remote))
        self.git(self.source, "push", "origin", "master")
        self.remote_patch = mock.patch.object(release, "REMOTE", str(self.remote))
        self.remote_patch.start()
        self.addCleanup(self.remote_patch.stop)

    def git(self, directory, *args):
        command = [
            "git",
            "-c",
            "user.name=release-test",
            "-c",
            "user.email=release-test@example.invalid",
            "-c",
            "commit.gpgsign=false",
            "-c",
            "tag.gpgsign=false",
            "-c",
            "core.hooksPath=/dev/null",
            *args,
        ]
        result = subprocess.run(
            command, cwd=directory, check=True, text=True, capture_output=True
        )
        return result.stdout

    def tags(self):
        return self.git(
            self.directory, "--git-dir", str(self.remote), "show-ref", "--tags"
        )

    def create_local_pair(self):
        for tag in (f"v{VERSION}", f"go/v{VERSION}"):
            self.git(self.source, "tag", "-a", tag, self.commit, "-m", tag)

    def test_verifies_maintainer_created_pair_without_mutation(self):
        self.create_local_pair()
        self.git(
            self.source,
            "push",
            "--atomic",
            "origin",
            f"refs/tags/v{VERSION}",
            f"refs/tags/go/v{VERSION}",
        )
        first = self.tags()
        release.verify_tag_pair(self.source, VERSION, self.commit)
        release.verify_tag_pair(self.source, VERSION, self.commit)
        self.assertEqual(self.tags(), first)
        for tag in (f"v{VERSION}", f"go/v{VERSION}"):
            self.assertEqual(
                self.git(
                    self.directory,
                    "--git-dir",
                    str(self.remote),
                    "cat-file",
                    "-t",
                    f"refs/tags/{tag}",
                ).strip(),
                "tag",
            )
            self.assertEqual(
                self.git(
                    self.directory,
                    "--git-dir",
                    str(self.remote),
                    "rev-parse",
                    f"refs/tags/{tag}^{{}}",
                ).strip(),
                self.commit,
            )

    def test_operator_verification_needs_no_ci_identity_and_matches_rust_source(self):
        self.create_local_pair()
        self.git(
            self.source,
            "push",
            "--atomic",
            "origin",
            f"refs/tags/v{VERSION}",
            f"refs/tags/go/v{VERSION}",
        )
        before = self.tags()
        registry = MemoryRegistry()
        for name in release.PACKAGES:
            registry.add(name, commit=self.commit)
        with (
            mock.patch.dict(os.environ, {}, clear=True),
            mock.patch.object(release, "Registry", return_value=registry),
            mock.patch.object(release, "verify_go") as consumer,
        ):
            self.assertEqual(
                release.main(
                    [
                        "verify-go",
                        "--source",
                        str(self.source),
                        "--version",
                        VERSION,
                        "--commit",
                        self.commit,
                    ]
                ),
                0,
            )
        consumer.assert_called_once_with(self.source, VERSION, self.commit)
        self.assertEqual(self.tags(), before)

    def test_operator_verification_rejects_tags_differing_from_published_rust(self):
        self.create_local_pair()
        self.git(
            self.source,
            "push",
            "--atomic",
            "origin",
            f"refs/tags/v{VERSION}",
            f"refs/tags/go/v{VERSION}",
        )
        before = self.tags()
        registry = MemoryRegistry()
        for name in release.PACKAGES:
            registry.add(name, commit=OTHER_SHA)
        with (
            mock.patch.dict(os.environ, {}, clear=True),
            mock.patch.object(release, "Registry", return_value=registry),
            mock.patch.object(release, "verify_go") as consumer,
            mock.patch("sys.stderr", new_callable=io.StringIO),
        ):
            self.assertEqual(
                release.main(
                    [
                        "verify-go",
                        "--source",
                        str(self.source),
                        "--version",
                        VERSION,
                        "--commit",
                        self.commit,
                    ]
                ),
                1,
            )
        consumer.assert_not_called()
        self.assertEqual(self.tags(), before)

    def test_incomplete_pair_requires_handoff_and_preserves_existing_root(self):
        self.git(
            self.directory,
            "--git-dir",
            str(self.remote),
            "tag",
            "-a",
            f"v{VERSION}",
            self.commit,
            "-m",
            "existing root tag",
        )
        before = self.tags()
        with self.assertRaisesRegex(release.ReleaseError, "pair is incomplete"):
            release.verify_tag_pair(self.source, VERSION, self.commit)
        self.assertEqual(self.tags(), before)

    def test_rejects_lightweight_tag_without_creating_sibling(self):
        self.git(
            self.directory,
            "--git-dir",
            str(self.remote),
            "tag",
            f"v{VERSION}",
            self.commit,
        )
        before = self.tags()
        with self.assertRaisesRegex(release.ReleaseError, "not an annotated tag"):
            release.verify_tag_pair(self.source, VERSION, self.commit)
        self.assertEqual(self.tags(), before)

    def test_rejects_dirty_source_before_any_tag_creation(self):
        (self.source / "README.md").write_text("uncommitted change\n")
        with self.assertRaisesRegex(release.ReleaseError, "uncommitted"):
            release.validate_checkout(self.source, self.commit)
        self.assertEqual(
            self.git(self.directory, "--git-dir", str(self.remote), "tag", "--list"), ""
        )

    def test_rejects_annotated_tag_at_another_commit(self):
        original = self.commit
        (self.source / "README.md").write_text("next controlled source\n")
        self.git(self.source, "add", "README.md")
        self.git(self.source, "commit", "-m", "Change controlled source")
        selected = self.git(self.source, "rev-parse", "HEAD").strip()
        self.git(self.source, "push", "origin", "master")
        self.git(
            self.directory,
            "--git-dir",
            str(self.remote),
            "tag",
            "-a",
            f"v{VERSION}",
            original,
            "-m",
            "previous source",
        )
        before = self.tags()
        with self.assertRaisesRegex(release.ReleaseError, "not an annotated tag"):
            release.verify_tag_pair(self.source, VERSION, selected)
        self.assertEqual(self.tags(), before)

    def test_rejects_source_not_merged_to_master(self):
        (self.source / "README.md").write_text("unmerged controlled source\n")
        self.git(self.source, "add", "README.md")
        self.git(self.source, "commit", "-m", "Create unmerged controlled source")
        unmerged = self.git(self.source, "rev-parse", "HEAD").strip()
        with self.assertRaisesRegex(release.ReleaseError, "--is-ancestor"):
            release.validate_checkout(self.source, unmerged)
        self.assertEqual(
            self.git(self.directory, "--git-dir", str(self.remote), "tag", "--list"), ""
        )

    def test_atomic_push_rejection_leaves_neither_tag_and_can_resume(self):
        hook = self.remote / "hooks/update"
        hook.write_text(
            '#!/bin/sh\ncase "$1" in refs/tags/go/*) exit 1;; esac\nexit 0\n'
        )
        hook.chmod(0o755)
        self.create_local_pair()
        with self.assertRaises(subprocess.CalledProcessError):
            self.git(
                self.source,
                "push",
                "--atomic",
                "origin",
                f"refs/tags/v{VERSION}",
                f"refs/tags/go/v{VERSION}",
            )
        self.assertEqual(
            self.git(self.directory, "--git-dir", str(self.remote), "tag", "--list"), ""
        )
        hook.write_text("#!/bin/sh\nexit 0\n")
        self.git(
            self.source,
            "push",
            "--atomic",
            "origin",
            f"refs/tags/v{VERSION}",
            f"refs/tags/go/v{VERSION}",
        )
        release.verify_tag_pair(self.source, VERSION, self.commit)
        self.assertIn(f"refs/tags/go/v{VERSION}", self.tags())


if __name__ == "__main__":
    unittest.main()
