#!/usr/bin/env python3
"""Fail on reachable Go vulnerabilities; narrowly review known SDK false positives."""

import datetime
import json
import pathlib
import sys


def decode_stream(text):
    decoder = json.JSONDecoder()
    messages = []
    while text.strip():
        text = text.lstrip()
        message, end = decoder.raw_decode(text)
        messages.append(message)
        text = text[end:]
    if not any("config" in message for message in messages):
        raise ValueError("missing govulncheck configuration; scan is incomplete")
    return messages


def check(messages, exceptions, today):
    advisories = {m["osv"]["id"]: m["osv"] for m in messages if "osv" in m}
    failures, reviewed, module_only = set(), set(), set()
    for message in messages:
        if "finding" not in message:
            continue
        finding = message["finding"]
        first = finding["trace"][0]
        identifier = finding["osv"]
        if not first.get("function"):
            module_only.add(identifier)
            continue
        exception = exceptions.get(identifier)
        advisory = advisories.get(identifier, {})
        affected = [a for a in advisory.get("affected", []) if a["package"]["name"] == first["module"]]
        package = first.get("package", "")
        # These advisories currently map every SDK symbol to daemon-only bugs.
        # Fail if the database narrows the advisory, the dependency changes, the
        # review expires, or any daemon/plugin code becomes reachable.
        accepted = (
            exception is not None
            and today <= datetime.date.fromisoformat(exception["review_by"])
            and first["module"] == exception["module"]
            and first["version"] == exception["version"]
            and advisory.get("database_specific", {}).get("review_status") == "UNREVIEWED"
            and bool(affected)
            and all(not a.get("ecosystem_specific", {}).get("imports") for a in affected)
            and (package in exception["client_packages"] or any(package == prefix or package.startswith(prefix + "/") for prefix in exception["client_package_prefixes"]))
        )
        (reviewed if accepted else failures).add(identifier)
    return failures, reviewed, module_only - failures - reviewed


if __name__ == "__main__":
    messages = decode_stream(pathlib.Path(sys.argv[1]).read_text())
    exceptions = json.loads(pathlib.Path(__file__).with_name("go-vulnerability-reviews.json").read_text())
    failures, reviewed, module_only = check(messages, exceptions, datetime.datetime.now(datetime.timezone.utc).date())
    print("Reviewed SDK-only mappings:", ", ".join(sorted(reviewed)) or "none")
    print("Module/package-only findings (not reachable):", ", ".join(sorted(module_only)) or "none")
    if failures:
        print("Unreviewed reachable vulnerabilities:", ", ".join(sorted(failures)), file=sys.stderr)
        sys.exit(1)
