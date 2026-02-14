#!/bin/sh
# Smoke tests for the gen command-line utility.
# Run from the repository root: sh gen/test.sh

set -e

cd "$(dirname "$0")/.."
go build -o gen/gen ./gen
trap 'rm -f gen/gen' EXIT

pass=0
fail=0

check() {
    desc="$1"; shift
    if "$@"; then
        pass=$((pass + 1))
    else
        echo "FAIL: $desc"
        fail=$((fail + 1))
    fi
}

check_fail() {
    desc="$1"; shift
    if "$@" 2>/dev/null; then
        echo "FAIL (expected error): $desc"
        fail=$((fail + 1))
    else
        pass=$((pass + 1))
    fi
}

# --- Output length checks ---

check "96-bit ID is 15 chars" \
    test "$(gen/gen | wc -c)" -eq 16  # 15 chars + newline

check "UUID v4 is 20 chars" \
    test "$(gen/gen -v4 | wc -c)" -eq 21  # 20 chars + newline

check "UUID v7 is 20 chars" \
    test "$(gen/gen -v7 | wc -c)" -eq 21

# --- Count checks ---

check "default count is 1 line" \
    test "$(gen/gen | wc -l)" -eq 1

check "-n 5 produces 5 lines" \
    test "$(gen/gen -n 5 | wc -l)" -eq 5

check "-n 5 -v4 produces 5 lines" \
    test "$(gen/gen -v4 -n 5 | wc -l)" -eq 5

check "-n 5 -v7 produces 5 lines" \
    test "$(gen/gen -v7 -n 5 | wc -l)" -eq 5

# --- Sequential: stable suffix for 96-bit (little-endian increment) ---

seq_out=$(gen/gen -n 5 -seq)
first_suffix=$(echo "$seq_out" | head -1 | cut -c6-15)
suffix_mismatches=$(echo "$seq_out" | cut -c6-15 | grep -cFv "$first_suffix" || true)
check "sequential 96-bit IDs share a suffix" \
    test "$suffix_mismatches" -eq 0

# --- Sequential: lines are distinct ---

check "-n 5 -seq produces 5 distinct lines" \
    test "$(gen/gen -n 5 -seq | sort -u | wc -l)" -eq 5

check "-n 5 produces 5 distinct lines" \
    test "$(gen/gen -n 5 | sort -u | wc -l)" -eq 5

# --- Sequential UUID modes produce distinct valid-length output ---

check "-v4 -n 5 -seq distinct" \
    test "$(gen/gen -v4 -n 5 -seq | sort -u | wc -l)" -eq 5

check "-v7 -n 5 -seq distinct" \
    test "$(gen/gen -v7 -n 5 -seq | sort -u | wc -l)" -eq 5

# --- Error cases ---

check_fail "-v4 -v7 rejects mutually exclusive flags" \
    gen/gen -v4 -v7

check_fail "-n 0 rejects zero count" \
    gen/gen -n 0

# --- Summary ---

echo
echo "$pass passed, $fail failed"
test "$fail" -eq 0
