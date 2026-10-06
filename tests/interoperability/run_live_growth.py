#!/usr/bin/env python3
"""Deterministic regular/compact growth checks for both SDK writers/readers.

Writers pause for reader acknowledgements, so no polling or timing race can miss
the selected stage. Rust additionally pauses after DATA preparation and before
ENTRY publication. This supplements the stock/live feature matrix.
"""
from __future__ import annotations

import hashlib
import json
import os
from pathlib import Path
import queue
import shlex
import subprocess
import sys
import tempfile
import threading

from run_live_matrix import REPO_ROOT, build_env


def run(command: list[str], *, cwd: Path, env: dict[str, str]) -> str:
    print(f"+ {shlex.join(command)}", file=sys.stderr, flush=True)
    result = subprocess.run(command, cwd=cwd, env=env, text=True, capture_output=True, timeout=180)
    if result.returncode:
        print(f"failed in {cwd}, status {result.returncode}: {result.stderr}", file=sys.stderr)
        raise SystemExit(result.returncode)
    return result.stdout


def read_stages(process: subprocess.Popen, stages: queue.Queue) -> None:
    for line in process.stdout:
        stages.put(line)
    stages.put(None)


def main() -> None:
    env = build_env()
    destination = REPO_ROOT / ".local" / "interoperability" / "live-growth"
    destination.mkdir(parents=True, exist_ok=True)
    suffix = ".exe" if os.name == "nt" else ""
    go = destination / f"go-livegrowth{suffix}"
    rust = Path(env["CARGO_TARGET_DIR"]) / "debug" / "examples" / f"live_growth{suffix}"
    run(["go", "build", "-o", str(go), "./internal/testcmd/livegrowth"], cwd=REPO_ROOT / "go", env=env)
    run(["cargo", "build", "--manifest-path", "rust/Cargo.toml", "-p", "systemd-journal-sdk-core", "--example", "live_growth"], cwd=REPO_ROOT, env=env)
    tools = {"go": go, "rust": rust}
    expected = [hashlib.sha256(value).hexdigest() for value in
                (b"MESSAGE=seed", b"MESSAGE=" + b"x" * (9 * 1024 * 1024 - 8))]
    observations = 0
    for writer, binary in tools.items():
        for compact in (False, True):
            with tempfile.TemporaryDirectory(dir=destination) as temporary:
                path = Path(temporary) / "system.journal"
                command = [str(binary), "write", str(path)] + (["compact"] if compact else [])
                print(f"+ {shlex.join(command)}", file=sys.stderr, flush=True)
                with tempfile.TemporaryFile(mode="w+") as errors:
                    process = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                               stderr=errors, text=True, env=env)
                    stages: queue.Queue = queue.Queue()
                    reader = threading.Thread(target=read_stages, args=(process, stages), daemon=True)
                    reader.start()
                    seen = []
                    try:
                        while True:
                            line = stages.get(timeout=30)
                            if line is None:
                                break
                            stage = json.loads(line)
                            seen.append(stage["stage"])
                            with path.open("rb") as journal:
                                header = journal.read(160)
                                physical = os.fstat(journal.fileno()).st_size
                            declared = sum(int.from_bytes(header[i:i+8], "little") for i in (88, 96))
                            if stage["stage"] == "prepared":
                                assert declared > physical > 8 * 1024 * 1024, (declared, physical)
                            elif stage["stage"] == "committed":
                                assert physical >= declared > 8 * 1024 * 1024, (declared, physical)
                            for name, tool in tools.items():
                                hashes = json.loads(run([str(tool), "read", str(path)], cwd=REPO_ROOT, env=env))
                                assert hashes == expected[:stage["entries"]], (writer, compact, stage, name, hashes)
                                observations += 1
                            process.stdin.write("readers complete\n")
                            process.stdin.flush()
                        status = process.wait(timeout=30)
                        errors.seek(0)
                        assert status == 0, (writer, status, errors.read())
                        assert seen == (["seed", "prepared", "committed"] if writer == "rust" else ["seed", "committed"]), seen
                    finally:
                        # Terminate only this task-owned writer on a failed check.
                        if process.poll() is None:
                            process.kill()
                            process.wait()
                        process.stdin.close()
                        reader.join(timeout=5)
                        process.stdout.close()
                print(f"PASS {writer} writer, compact={compact}, stages={seen}")
    print(f"PASS {observations} reader observations across 2 writers, 2 readers and 2 layouts")


if __name__ == "__main__":
    main()
