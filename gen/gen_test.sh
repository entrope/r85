#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

BIN="$TMPDIR/r85gen"
go build -C "$SCRIPT_DIR" -o "$BIN" .

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1 -- $2"; FAIL=$((FAIL + 1)); }

# Test 1: Default mode - one 15-char ID
out=$("$BIN")
lines=$(echo "$out" | wc -l | tr -d ' ')
len=${#out}
if [ "$lines" -eq 1 ] && [ "$len" -eq 15 ]; then
	pass "default_single_id"
else
	fail "default_single_id" "lines=$lines len=$len"
fi

# Test 2: -n 5 produces 5 lines, each 15 chars
out=$("$BIN" -n 5)
lines=$(echo "$out" | wc -l | tr -d ' ')
bad=0
while IFS= read -r line; do
	[ ${#line} -eq 15 ] || bad=$((bad + 1))
done <<< "$out"
if [ "$lines" -eq 5 ] && [ "$bad" -eq 0 ]; then
	pass "n_flag_count_and_lengths"
else
	fail "n_flag_count_and_lengths" "lines=$lines bad_lengths=$bad"
fi

# Test 3: UUID v4 - one 20-char ID
out=$("$BIN" -uuid v4)
len=${#out}
if [ "$len" -eq 20 ]; then
	pass "v4_length"
else
	fail "v4_length" "len=$len expected=20"
fi

# Test 4: UUID v7 - one 20-char ID
out=$("$BIN" -uuid v7)
len=${#out}
if [ "$len" -eq 20 ]; then
	pass "v7_length"
else
	fail "v7_length" "len=$len expected=20"
fi

# Test 5: Sequential mode produces correct count
out=$("$BIN" -n 10 -seq)
lines=$(echo "$out" | wc -l | tr -d ' ')
if [ "$lines" -eq 10 ]; then
	pass "seq_count"
else
	fail "seq_count" "lines=$lines expected=10"
fi

# Test 6: Sequential v4
out=$("$BIN" -n 5 -uuid v4 -seq)
lines=$(echo "$out" | wc -l | tr -d ' ')
bad=0
while IFS= read -r line; do
	[ ${#line} -eq 20 ] || bad=$((bad + 1))
done <<< "$out"
if [ "$lines" -eq 5 ] && [ "$bad" -eq 0 ]; then
	pass "seq_v4"
else
	fail "seq_v4" "lines=$lines bad_lengths=$bad"
fi

# Test 7: Sequential v7
out=$("$BIN" -n 5 -uuid v7 -seq)
lines=$(echo "$out" | wc -l | tr -d ' ')
bad=0
while IFS= read -r line; do
	[ ${#line} -eq 20 ] || bad=$((bad + 1))
done <<< "$out"
if [ "$lines" -eq 5 ] && [ "$bad" -eq 0 ]; then
	pass "seq_v7"
else
	fail "seq_v7" "lines=$lines bad_lengths=$bad"
fi

# Test 8: Invalid -uuid flag causes error
if "$BIN" -uuid v3 2>/dev/null; then
	fail "invalid_uuid_flag" "should have exited non-zero"
else
	pass "invalid_uuid_flag"
fi

# Test 9: -n 0 causes error
if "$BIN" -n 0 2>/dev/null; then
	fail "n_zero" "should have exited non-zero"
else
	pass "n_zero"
fi

# Test 10: All output characters are valid r85 (ASCII 40-126)
# Use tr to delete all valid r85 characters; if anything remains, the line is bad.
# The r85 alphabet spans ASCII 40 '(' through ASCII 126 '~'.
out=$("$BIN" -n 20)
bad=0
while IFS= read -r line; do
	remainder=$(printf '%s' "$line" | LC_ALL=C tr -d '(-~')
	if [ -n "$remainder" ]; then
		bad=$((bad + 1))
	fi
done <<< "$out"
if [ "$bad" -eq 0 ]; then
	pass "valid_r85_chars"
else
	fail "valid_r85_chars" "$bad lines had invalid chars"
fi

# Test 11: 100 random IDs are all unique
out=$("$BIN" -n 100)
unique=$(echo "$out" | sort -u | wc -l | tr -d ' ')
if [ "$unique" -eq 100 ]; then
	pass "random_uniqueness"
else
	fail "random_uniqueness" "unique=$unique expected=100"
fi

# Test 12: 100 sequential IDs are all unique
out=$("$BIN" -n 100 -seq)
unique=$(echo "$out" | sort -u | wc -l | tr -d ' ')
if [ "$unique" -eq 100 ]; then
	pass "seq_uniqueness"
else
	fail "seq_uniqueness" "unique=$unique expected=100"
fi

# Test 13: Extra positional args rejected
if "$BIN" extra_arg 2>/dev/null; then
	fail "extra_args" "should have exited non-zero"
else
	pass "extra_args"
fi

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] || exit 1
