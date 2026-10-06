#!/usr/bin/env python3
"""Publish Rust in CI and verify the maintainer's matching Go release."""

from __future__ import annotations

import argparse
import hashlib
import io
import json
import os
import re
import shlex
import subprocess  # nosec B404
import sys
import tarfile
import tempfile
import time
import tomllib
import urllib.error
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
REPOSITORY = "netdata/systemd-journal-sdk"
REMOTE = f"https://github.com/{REPOSITORY}.git"
GO_MODULE = f"github.com/{REPOSITORY}/go"
PACKAGES = (
    "systemd-journal-sdk-common",
    "systemd-journal-sdk-registry",
    "systemd-journal-sdk-core",
    "systemd-journal-sdk-host",
    "systemd-journal-sdk-log-writer",
    "systemd-journal-sdk-index",
    "systemd-journal-sdk-engine",
    "systemd-journal-sdk",
)
VERSION_RE = re.compile(r"(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)")
SHA_RE = re.compile(r"[0-9a-f]{40}")
WAIT_SECONDS = 180
HTTP_LIMIT = 16 * 1024 * 1024


class ReleaseError(RuntimeError):
    """An incomplete or conflicting release that must stop this run."""


class RegistryUnavailable(ReleaseError):
    """A temporary transport failure, never evidence that a version is absent."""


def run(args, *, cwd, capture=False, env=None, timeout=900):
    command = shlex.join(str(arg) for arg in args)
    print(f"{cwd.name} > {command}", file=sys.stderr, flush=True)
    try:
        # Callers supply fixed tool vectors and validated release arguments.
        result = subprocess.run(  # nosec B603
            args,
            cwd=cwd,
            env=env,
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE if capture else None,
            text=True,
            timeout=timeout,
            check=False,
        )
    except subprocess.TimeoutExpired as error:
        raise ReleaseError(f"Command timed out after {timeout}s: {command}") from error
    if result.returncode:
        raise ReleaseError(
            f"Command exited {result.returncode} in {cwd.name}: {command}"
        )
    return result.stdout if capture else ""


def validate_inputs(version, commit):
    if not VERSION_RE.fullmatch(version):
        raise ReleaseError(
            "Version must be X.Y.Z with no v prefix or prerelease suffix"
        )
    if not SHA_RE.fullmatch(commit):
        raise ReleaseError(
            "Source commit must be the full 40-character lowercase Git SHA"
        )
    if int(version.split(".")[0]) > 1:
        raise ReleaseError(
            "A v2+ Go release needs a reviewed module-path migration first"
        )


def require_ci():
    expected = {
        "GITHUB_ACTIONS": "true",
        "GITHUB_REPOSITORY": REPOSITORY,
        "GITHUB_REF": "refs/heads/master",
        "GITHUB_EVENT_NAME": "workflow_dispatch",
    }
    if any(os.environ.get(key) != value for key, value in expected.items()):
        raise ReleaseError(
            "Publication runs only through manual Release on canonical master"
        )


def source_directory(value):
    source = Path(value).resolve(strict=True)
    if not source.is_dir() or not source.is_relative_to(ROOT) or len(str(source)) < 10:
        raise ReleaseError("Release source must be a directory inside this checkout")
    return source


def output(key, value):
    if path := os.environ.get("GITHUB_OUTPUT"):
        with Path(path).open("a", encoding="utf-8") as stream:
            stream.write(f"{key}={value}\n")


def manifest(source, version):
    workspace = tomllib.loads((source / "rust/Cargo.toml").read_text())
    package = workspace["workspace"]["package"]
    if package["version"] != version:
        raise ReleaseError(
            f"Rust workspace version is {package['version']}, requested {version}"
        )
    rust_version = package["rust-version"]
    if not re.fullmatch(r"[0-9]+\.[0-9]+(?:\.[0-9]+)?", rust_version):
        raise ReleaseError("Workspace must declare a numeric Rust compiler minimum")
    go_text = (source / "go/go.mod").read_text()
    if not re.search(rf"^module\s+{re.escape(GO_MODULE)}\s*$", go_text, re.MULTILINE):
        raise ReleaseError("Unexpected Go module path")
    output("rust-version", rust_version)
    return workspace


def validate_checkout(source, commit):
    head = run(["git", "rev-parse", "HEAD"], cwd=source, capture=True).strip()
    if head != commit:
        raise ReleaseError(f"Checkout SHA {head} differs from selected source {commit}")
    if run(["git", "status", "--porcelain"], cwd=source, capture=True).strip():
        raise ReleaseError("Release source has uncommitted or untracked files")
    remote = run(
        ["git", "remote", "get-url", "origin"], cwd=source, capture=True
    ).strip()
    if remote not in (
        REMOTE,
        REMOTE.removesuffix(".git"),
        f"git@github.com:{REPOSITORY}.git",
    ):
        raise ReleaseError(
            "Release source origin must identify the canonical repository"
        )
    run(["git", "merge-base", "--is-ancestor", commit, "origin/master"], cwd=source)


def validate_metadata(metadata, version):
    members = set(metadata["workspace_members"])
    packages = {p["name"]: p for p in metadata["packages"] if p["id"] in members}
    publishable = {name for name, p in packages.items() if p["publish"] != []}
    if publishable != set(PACKAGES):
        raise ReleaseError(f"Publishable package set differs: {sorted(publishable)}")
    for name in PACKAGES:
        package = packages[name]
        if package["version"] != version:
            raise ReleaseError(f"{name} version differs from {version}")
        if package["publish"] is not None and "crates-io" not in package["publish"]:
            raise ReleaseError(f"{name} does not allow crates.io publication")
        for dependency in package["dependencies"]:
            if dependency.get("path") is None:
                continue
            if dependency["name"] not in PACKAGES:
                if dependency.get("kind") != "dev":
                    raise ReleaseError(f"{name} depends on an unpublished path package")
                continue
            if dependency["req"] not in (f"^{version}", f"={version}"):
                raise ReleaseError(f"{name} has a mismatched internal version pin")
            if dependency.get("kind") != "dev" and PACKAGES.index(
                dependency["name"]
            ) >= PACKAGES.index(name):
                raise ReleaseError(
                    f"Publication order no longer fits dependency {dependency['name']}"
                )


def validate_install_examples(source, version):
    files = [source / path for path in ("README.md", "go/README.md", "rust/README.md")]
    for directory in ("docs", ".agents/sow/specs"):
        files.extend((source / directory).rglob("*.md"))
    rust_pattern = re.compile(
        r'package\s*=\s*"systemd-journal-sdk(?:-[a-z-]+)?"\s*,\s*version\s*=\s*"([0-9]+\.[0-9]+\.[0-9]+)"'
    )
    go_pattern = re.compile(re.escape(GO_MODULE) + r"@v([0-9]+\.[0-9]+\.[0-9]+)")
    for path in files:
        text = path.read_text()
        for found in (*rust_pattern.findall(text), *go_pattern.findall(text)):
            if found != version:
                raise ReleaseError(
                    f"Stale install version {found} in {path.relative_to(source)}"
                )


def http_get(url, *, missing_ok=False, headers=None):
    request_headers = {"User-Agent": "systemd-journal-sdk release workflow"}
    request_headers.update(headers or {})
    try:
        request = urllib.request.Request(url, headers=request_headers)
        # Callers construct fixed GitHub/crates.io HTTPS endpoints.
        with urllib.request.urlopen(request, timeout=30) as response:  # nosec B310
            data = response.read(HTTP_LIMIT + 1)
    except urllib.error.HTTPError as error:
        if error.code == 404 and missing_ok:
            return None
        if error.code == 429 or error.code >= 500:
            raise RegistryUnavailable(f"HTTP {error.code}: {url}") from error
        raise ReleaseError(f"HTTP {error.code}: {url}") from error
    except (urllib.error.URLError, TimeoutError) as error:
        raise RegistryUnavailable(f"Transport failed: {url}") from error
    if len(data) > HTTP_LIMIT:
        raise ReleaseError(f"Registry response exceeded the size limit: {url}")
    return data


def validate_environment(environment, policies):
    if environment.get("deployment_branch_policy") != {
        "protected_branches": False,
        "custom_branch_policies": True,
    }:
        raise ReleaseError(
            "Configure release environment for selected branches: master only"
        )
    rules = policies.get("branch_policies", [])
    if (
        len(rules) != 1
        or rules[0].get("name") != "master"
        or rules[0].get("type") != "branch"
    ):
        raise ReleaseError(
            "Release environment must allow only the master branch, no tags"
        )


def check_environment():
    url = f"https://api.github.com/repos/{REPOSITORY}/environments/release"
    headers = {
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    if token := os.environ.get("GH_TOKEN"):
        headers["Authorization"] = f"Bearer {token}"
    data = http_get(url, missing_ok=True, headers=headers)
    if data is None:
        raise ReleaseError(
            "Create the release environment and Trusted Publishing bindings; see RELEASING.md"
        )
    policies = http_get(
        url + "/deployment-branch-policies?per_page=100", headers=headers
    )
    validate_environment(json.loads(data), json.loads(policies))


def index_path(name):
    return f"{name[:2]}/{name[2:4]}/{name}"


def verify_archive(data, checksum, name, version, commit):
    if (
        not re.fullmatch(r"[0-9a-f]{64}", checksum)
        or hashlib.sha256(data).hexdigest() != checksum
    ):
        raise ReleaseError(f"{name} archive checksum differs from the registry")
    prefix = f"{name}-{version}/"
    try:
        with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as archive:

            def read_member(filename):
                member = archive.getmember(prefix + filename)
                if not member.isfile() or member.size > 1024 * 1024:
                    raise ReleaseError(f"Invalid {filename} in {name}")
                with archive.extractfile(member) as stream:
                    return stream.read()

            vcs = json.loads(read_member(".cargo_vcs_info.json"))
            package = tomllib.loads(read_member("Cargo.toml").decode())["package"]
    except (tarfile.TarError, KeyError, ValueError) as error:
        raise ReleaseError(
            f"Missing or invalid published metadata for {name}"
        ) from error
    if (
        vcs.get("git", {}).get("sha1") != commit
        or vcs.get("git", {}).get("dirty", False) is not False
    ):
        raise ReleaseError(
            f"{name} was published from a different or dirty source; never skip it"
        )
    if package.get("name") != name or package.get("version") != version:
        raise ReleaseError(f"{name} published package identity differs")


class Registry:
    def probe(self, name, version):
        api_data = http_get(
            f"https://crates.io/api/v1/crates/{name}/{version}", missing_ok=True
        )
        index_data = http_get(
            f"https://index.crates.io/{index_path(name)}", missing_ok=True
        )
        api = json.loads(api_data)["version"] if api_data is not None else None
        index = None
        if index_data is not None:
            entries = (json.loads(line) for line in index_data.splitlines())
            index = next((entry for entry in entries if entry["vers"] == version), None)
        return api, index

    def verify(self, name, version, commit, api, index):
        if (
            api.get("num") != version
            or index.get("vers") != version
            or index.get("name") != name
        ):
            raise ReleaseError(f"Unexpected registry identity for {name}")
        if api.get("yanked") is not False or index.get("yanked") is not False:
            raise ReleaseError(f"{name} {version} is yanked; stop this release")
        if api.get("checksum") != index.get("cksum"):
            raise RegistryUnavailable(
                f"Registry API/index checksums have not converged for {name}"
            )
        data = http_get(
            f"https://static.crates.io/crates/{name}/{name}-{version}.crate"
        )
        verify_archive(data, index["cksum"], name, version, commit)

    def existing(self, name, version, commit):
        api, index = self.probe(name, version)
        if api is None and index is None:
            return False
        self.wait(name, version, commit)
        return True

    def wait(self, name, version, commit):
        deadline = time.monotonic() + WAIT_SECONDS
        while True:
            try:
                api, index = self.probe(name, version)
                if api is not None and index is not None:
                    self.verify(name, version, commit, api, index)
                    print(f"Verified {name} {version} from {commit}", flush=True)
                    return
            except RegistryUnavailable as error:
                print(f"Waiting: {error}", file=sys.stderr, flush=True)
            if time.monotonic() >= deadline:
                raise ReleaseError(
                    f"{name} {version} is not fully available after {WAIT_SECONDS}s; rerun the same source"
                )
            time.sleep(5)


def remote_tags(source, version, commit):
    refs = [
        f"refs/tags/{tag}{suffix}"
        for tag in (f"v{version}", f"go/v{version}")
        for suffix in ("", "^{}")
    ]
    result = run(
        ["git", "ls-remote", "--tags", REMOTE, *refs],
        cwd=source,
        capture=True,
        timeout=60,
    )
    found = dict(line.split()[::-1] for line in result.splitlines())
    for tag in (f"v{version}", f"go/v{version}"):
        ref = f"refs/tags/{tag}"
        if ref in found and found.get(ref + "^{}") != commit:
            raise ReleaseError(
                f"Existing {tag} is not an annotated tag at {commit}; never move it"
            )
    return found


def preflight(source, version, commit, registry):
    validate_checkout(source, commit)
    manifest(source, version)
    metadata = json.loads(
        run(
            [
                "cargo",
                "metadata",
                "--manifest-path",
                str(source / "rust/Cargo.toml"),
                "--no-deps",
                "--format-version",
                "1",
                "--locked",
            ],
            cwd=source,
            capture=True,
        )
    )
    validate_metadata(metadata, version)
    validate_install_examples(source, version)
    remote_tags(source, version, commit)
    for name in PACKAGES:
        registry.existing(name, version, commit)


def prepare(source, version, commit, name, registry):
    validate_checkout(source, commit)
    for predecessor in PACKAGES[: PACKAGES.index(name)]:
        registry.wait(predecessor, version, commit)
    if registry.existing(name, version, commit):
        print(f"Already published and verified: {name} {version}")
        output("publish", "false")
        return False
    run(
        [
            "cargo",
            "publish",
            "--manifest-path",
            str(source / "rust/Cargo.toml"),
            "--package",
            name,
            "--registry",
            "crates-io",
            "--locked",
            "--dry-run",
        ],
        cwd=source,
    )
    output("publish", "true")
    return True


def publish(source, version, commit, name, registry):
    validate_checkout(source, commit)
    if registry.existing(name, version, commit):
        return
    try:
        run(
            [
                "cargo",
                "publish",
                "--manifest-path",
                str(source / "rust/Cargo.toml"),
                "--package",
                name,
                "--registry",
                "crates-io",
                "--locked",
            ],
            cwd=source,
        )
    except ReleaseError:
        # Cargo can time out after the registry accepted the upload.
        print(
            f"Upload command failed; checking whether {name} reached the registry",
            file=sys.stderr,
        )
        registry.wait(name, version, commit)
        return
    registry.wait(name, version, commit)


def verify_rust(source, version, commit, registry):
    for name in PACKAGES:
        registry.wait(name, version, commit)
    directory = consumer_directory("rust")
    (directory / "src").mkdir(exist_ok=True)
    (directory / "Cargo.toml").write_text(
        '[package]\nname = "release-consumer"\nversion = "0.0.0"\nedition = "2024"\n'
        "[workspace]\n[dependencies]\n"
        f'journal = {{ package = "systemd-journal-sdk", version = "={version}" }}\n'
    )
    (directory / "src/main.rs").write_text("fn main() {}\n")
    run(
        ["cargo", "check", "--manifest-path", str(directory / "Cargo.toml")], cwd=source
    )


def consumer_directory(language):
    parent = ROOT / ".local/release/consumers"
    if not parent.resolve().is_relative_to(ROOT.resolve(strict=True)):
        raise ReleaseError("Consumer scratch directory escapes this checkout")
    parent.mkdir(parents=True, exist_ok=True)
    if not parent.resolve(strict=True).is_relative_to(ROOT.resolve(strict=True)):
        raise ReleaseError("Created consumer scratch directory escapes this checkout")
    return Path(tempfile.mkdtemp(prefix=f"{language}-", dir=parent))


def verify_tag_pair(source, version, commit):
    found = remote_tags(source, version, commit)
    if any(
        f"refs/tags/{tag}^{{}}" not in found
        for tag in (f"v{version}", f"go/v{version}")
    ):
        raise ReleaseError(
            "Release tag pair is incomplete; a maintainer must push both "
            "annotated tags at the selected source before Go verification"
        )


def verify_go(source, version, commit):
    go_text = run(["git", "show", f"{commit}:go/go.mod"], cwd=source, capture=True)
    if not re.search(rf"^module\s+{re.escape(GO_MODULE)}\s*$", go_text, re.MULTILINE):
        raise ReleaseError("Selected source has an unexpected Go module path")
    go_version = re.search(
        r"^go\s+([0-9]+\.[0-9]+(?:\.[0-9]+)?)\s*$", go_text, re.MULTILINE
    )
    if go_version is None:
        raise ReleaseError("Go source must declare its compiler minimum")
    deadline = time.monotonic() + WAIT_SECONDS
    while True:
        try:
            value = json.loads(
                run(
                    ["go", "list", "-m", "-json", f"{GO_MODULE}@v{version}"],
                    cwd=source,
                    capture=True,
                    timeout=60,
                )
            )
        except ReleaseError:
            if time.monotonic() >= deadline:
                raise
            time.sleep(5)
            continue
        validate_go_lookup(value, version, commit)
        break
    directory = consumer_directory("go")
    (directory / "go.mod").write_text(
        f"module release-consumer\n\ngo {go_version[1]}\n\nrequire {GO_MODULE} v{version}\n"
    )
    (directory / "main.go").write_text(
        f'package main\n\nimport _ "{GO_MODULE}/journal"\n\nfunc main() {{}}\n'
    )
    run(["go", "mod", "download", f"{GO_MODULE}@v{version}"], cwd=directory)
    run(["go", "mod", "tidy"], cwd=directory)
    resolved = json.loads(
        run(["go", "list", "-m", "-json", GO_MODULE], cwd=directory, capture=True)
    )
    validate_go_lookup(resolved, version, commit)
    run(["go", "build", "-o", str(directory / "consumer"), "."], cwd=directory)


def validate_go_lookup(value, version, commit):
    if value.get("Path") != GO_MODULE or value.get("Version") != f"v{version}":
        raise ReleaseError("Go resolved an unexpected module/version")
    if value.get("Origin", {}).get("Hash", commit) != commit:
        raise ReleaseError("Go resolved a different source commit")


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "command",
        choices=(
            "inputs",
            "check-inputs",
            "manifest",
            "preflight",
            "prepare",
            "publish",
            "verify-rust",
            "verify-go",
        ),
    )
    parser.add_argument("--version", required=True)
    parser.add_argument("--commit", required=True)
    parser.add_argument("--source", default=str(ROOT))
    parser.add_argument("--package", choices=PACKAGES)
    args = parser.parse_args(argv)
    try:
        validate_inputs(args.version, args.commit)
        if args.command == "check-inputs":
            return 0
        if args.command == "inputs":
            require_ci()
            output("commit", args.commit)
            output("version", args.version)
            return 0
        source = source_directory(args.source)
        if args.command == "manifest":
            manifest(source, args.version)
            return 0
        if args.command == "verify-go":
            verify_tag_pair(source, args.version, args.commit)
            registry = Registry()
            for name in PACKAGES:
                registry.wait(name, args.version, args.commit)
            verify_go(source, args.version, args.commit)
            print(
                f"Verified Rust source, paired tags and Go {args.version} from {args.commit}"
            )
            return 0
        require_ci()
        registry = Registry()
        if args.command == "preflight":
            check_environment()
            preflight(source, args.version, args.commit, registry)
        elif args.command in ("prepare", "publish"):
            if not args.package:
                raise ReleaseError("Package name is required")
            operation = prepare if args.command == "prepare" else publish
            operation(source, args.version, args.commit, args.package, registry)
        elif args.command == "verify-rust":
            verify_rust(source, args.version, args.commit, registry)
            if summary := os.environ.get("GITHUB_STEP_SUMMARY"):
                with Path(summary).open("a", encoding="utf-8") as stream:
                    stream.write(
                        f"## Rust publication verified: {args.version}\n\n"
                        f"Source: `{args.commit}`\n\n"
                    )
                    stream.write(
                        "All eight Rust crates and an exact-version registry consumer passed.\n\n"
                        f"Next: a permitted maintainer pushes annotated `v{args.version}` "
                        f"and `go/v{args.version}` at this source, then verifies Go "
                        "consumption with the procedure in RELEASING.md.\n"
                    )
    except (ReleaseError, OSError, ValueError, KeyError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
