#!/usr/bin/env python3
"""Verify manually vendored asset hashes and optionally query OSV advisories."""

import hashlib
import json
import pathlib
import sys
import urllib.request

root = pathlib.Path(__file__).resolve().parents[2]
packages = json.loads((root / "web/static/vendor/xterm/manifest.json").read_text())
for package in packages:
    for artifact, expected in package["sha256"].items():
        actual = hashlib.sha256((root / "web/static/vendor/xterm" / artifact).read_bytes()).hexdigest()
        if actual != expected:
            raise ValueError(f"Vendor asset changed without inventory update: {artifact}")
    if "--audit" in sys.argv:
        query = {"package": {"ecosystem": "npm", "name": package["name"]}, "version": package["version"]}
        request = urllib.request.Request("https://api.osv.dev/v1/query", data=json.dumps(query).encode(), headers={"Content-Type": "application/json"})
        with urllib.request.urlopen(request, timeout=30) as response:
            findings = json.load(response).get("vulns", [])
        if findings:
            raise ValueError(f"Vendor vulnerabilities in {package['name']}: " + ", ".join(v["id"] for v in findings))
    print(f"Verified {package['name']}@{package['version']}")
