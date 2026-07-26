#!/usr/bin/env python3
"""Unit tests for run_writer_core_benchmarks.py."""

from __future__ import annotations

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import run_writer_core_benchmarks as writer_bench  # noqa: E402


class WriterCoreRunnerTests(unittest.TestCase):
    def test_netflow_workload_metadata_is_exact(self) -> None:
        self.assertEqual(
            writer_bench.workload_metadata("netflow-v5-repeating-256"),
            {
                "application_fields_per_row": 29,
                "entry_items_per_row": 30,
                "application_logical_bytes_per_row": 450,
                "total_logical_data_bytes_per_row": 491,
            },
        )

    def test_non_default_workload_supports_unique_rust_go_selections(self) -> None:
        writer_bench.validate_workload_languages(
            writer_bench.DEFAULT_WORKLOAD,
            ["systemd", "rust", "go"],
        )
        writer_bench.validate_workload_languages(
            "netflow-v5-repeating-256",
            ["rust"],
        )
        writer_bench.validate_workload_languages(
            "netflow-v5-repeating-256",
            ["go"],
        )
        writer_bench.validate_workload_languages(
            "netflow-v5-repeating-256",
            ["rust", "go"],
        )
        for languages in (["systemd"], ["rust", "rust"], []):
            with self.subTest(languages=languages):
                with self.assertRaises(ValueError):
                    writer_bench.validate_workload_languages(
                        "netflow-v5-repeating-256",
                        languages,
                    )

    def test_rust_command_records_selected_workload(self) -> None:
        command = writer_bench.bench_command(
            ["writer_core_bench"],
            language="rust",
            output=Path("/tmp/output.journal"),
            rows=256,
            journal_format="compact",
            final_state="online",
            max_size_bytes=128 * 1024 * 1024,
            workload="netflow-v5-repeating-256",
            api_mode="structured-field",
            rust_trusted_unique_payloads=True,
            live_publish_every_entries=0,
            rust_mmap_strategy="windowed",
        )

        workload_index = command.index("--workload")
        self.assertEqual(command[workload_index + 1], "netflow-v5-repeating-256")
        self.assertIn("--trusted-unique-payloads", command)

    def test_go_command_records_selected_workload(self) -> None:
        command = writer_bench.bench_command(
            ["go-writer-core-bench"],
            language="go",
            output=Path("/tmp/output.journal"),
            rows=256,
            journal_format="compact",
            final_state="online",
            max_size_bytes=128 * 1024 * 1024,
            workload="netflow-v5-repeating-256",
            api_mode="structured-field",
            rust_trusted_unique_payloads=False,
            live_publish_every_entries=0,
            rust_mmap_strategy="windowed",
        )

        workload_index = command.index("--workload")
        self.assertEqual(command[workload_index + 1], "netflow-v5-repeating-256")
        self.assertNotIn("--trusted-unique-payloads", command)

    def test_driver_shape_mismatch_is_a_failure(self) -> None:
        driver = {
            "workload": "netflow-v5-repeating-256",
            **writer_bench.workload_metadata("netflow-v5-repeating-256"),
        }
        self.assertEqual(
            writer_bench.driver_workload_errors(
                driver,
                "netflow-v5-repeating-256",
            ),
            [],
        )
        driver["entry_items_per_row"] = 29
        self.assertEqual(
            writer_bench.driver_workload_errors(
                driver,
                "netflow-v5-repeating-256",
            ),
            ["driver entry_items_per_row mismatch: got 29, want 30"],
        )

    def test_default_rust_driver_requires_complete_metadata(self) -> None:
        driver = {
            "workload": writer_bench.DEFAULT_WORKLOAD,
            **writer_bench.workload_metadata(writer_bench.DEFAULT_WORKLOAD),
        }
        driver.pop("total_logical_data_bytes_per_row")
        self.assertEqual(
            writer_bench.driver_workload_errors(
                driver,
                writer_bench.DEFAULT_WORKLOAD,
            ),
            [
                "driver total_logical_data_bytes_per_row mismatch: "
                "got None, want 817"
            ],
        )


if __name__ == "__main__":
    unittest.main()
