import datetime
import importlib.util
import json
import os
import pathlib
import subprocess
import unittest
import sys

sys.dont_write_bytecode = True

directory = pathlib.Path(__file__).parent
spec = importlib.util.spec_from_file_location("gate", directory / "check-go-vulnerabilities.py")
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)


class SecurityChecks(unittest.TestCase):
    def test_release_tags_and_version(self):
        for target, tag, expected in [("main", "v1.3.5", "version=v1.3.5"), ("dev", "v1.3.5", "version=devv1.3.5"), ("dev", "dev1.3.5", "version=dev1.3.5")]:
            result = self.release(target, tag)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn(expected, result.stdout)
            self.assertIn("ghcr.io/fixture/project:", result.stdout)

    def release(self, target, tag):
        env = dict(os.environ, RELEASE_TARGET=target, RELEASE_TAG=tag, RELEASE_REPOSITORY="Fixture/Project")
        return subprocess.run(["bash", str(directory / "resolve-release-tags.sh")], env=env, text=True, capture_output=True, timeout=5)

    def test_release_rejects_shell_syntax_and_other_targets(self):
        for tag in ["$(id)", "`id`", "v1\nversion=injected", "v1;id", "a" * 125, "-v1"]:
            self.assertNotEqual(self.release("main", tag).returncode, 0)
        self.assertNotEqual(self.release("feature", "v1.3.5").returncode, 0)

    def fixture(self):
        identifier = "GO-2026-4883"
        exceptions = json.loads((directory / "go-vulnerability-reviews.json").read_text())
        messages = [{"config": {}}, {"osv": {"id": identifier, "database_specific": {"review_status": "UNREVIEWED"}, "affected": [{"package": {"name": "github.com/docker/docker"}}]}}, {"finding": {"osv": identifier, "trace": [{"module": "github.com/docker/docker", "version": "v28.5.2+incompatible", "package": "github.com/docker/docker/client", "function": "ContainerList"}]}}]
        return messages, exceptions

    def test_known_sdk_mapping_review(self):
        messages, exceptions = self.fixture()
        failures, reviewed, _ = gate.check(messages, exceptions, datetime.date(2026, 10, 5))
        self.assertFalse(failures)
        self.assertEqual(reviewed, {"GO-2026-4883"})

    def test_review_expiry_version_change_daemon_and_database_refinement_fail(self):
        for case in ["expiry", "version", "daemon", "api_server", "reviewed", "imports", "new_advisory"]:
            messages, exceptions = self.fixture()
            today = datetime.date(2026, 10, 5)
            first = messages[2]["finding"]["trace"][0]
            if case == "expiry": today = datetime.date(2026, 12, 1)
            if case == "version": first["version"] = "v29.0.0"
            if case == "daemon": first["package"] = "github.com/docker/docker/daemon"
            if case == "api_server": first["package"] = "github.com/docker/docker/api/server/router"
            if case == "reviewed": messages[1]["osv"]["database_specific"]["review_status"] = "REVIEWED"
            if case == "imports": messages[1]["osv"]["affected"][0]["ecosystem_specific"] = {"imports": [{"path": first["package"]}]}
            if case == "new_advisory": messages[2]["finding"]["osv"] = "GO-2099-9999"
            self.assertTrue(gate.check(messages, exceptions, today)[0], case)

    def test_empty_or_malformed_scan_fails(self):
        for value in ["", "{}", '{"config":{}} garbage']:
            with self.assertRaises(ValueError): gate.decode_stream(value)


if __name__ == "__main__":
    unittest.main()
