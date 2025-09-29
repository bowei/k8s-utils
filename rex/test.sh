#!/bin/sh
set -e

if ! command -v jq >/dev/null 2>&1; then
    echo "jq is not installed. Please install jq to run this test."
    exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf -- "$TMPDIR"' EXIT

go run . --json sample > "$TMPDIR/sample.json"
go run . --json sample2 > "$TMPDIR/sample2.json"

if ! diff -u <(jq -S . "sample/expected.json") <(jq -S . "$TMPDIR/sample.json"); then
    echo "FAIL: sample output does not match expected output"
    exit 1
fi

if ! diff -u <(jq -S . "sample2/expected.json") <(jq -S . "$TMPDIR/sample2.json"); then
    echo "FAIL: sample2 output does not match expected output"
    exit 1
fi

echo "PASS"
