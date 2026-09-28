#!/usr/bin/env bash

set -euo pipefail

fixture_dir=$(mktemp -d "${ETE_DIR:-${TMPDIR:-/tmp}}/scnnr-e2e.XXXXXX")

trap 'rm -rf "$fixture_dir"' EXIT

mkdir -p "$fixture_dir/nested" "$fixture_dir/ignored"

printf 'needle needle\ntoken\n' > "$fixture_dir/alpha.txt"
printf 'needle\n' > "$fixture_dir/nested/beta.txt"
printf 'needle\n' > "$fixture_dir/ignored/hidden.txt"
printf 'needle\n' > "$fixture_dir/skip.log"
printf 'abc' > "$fixture_dir/abc.bin"

dd if=/dev/zero of="$fixture_dir/large.bin" bs=1000000 count=1 2>/dev/null

assert_output() {
    local expected=$1
    shift
    local actual
    
    # The original scanner appends matches from concurrent file readers.
    actual=$(go run main.go "$@" | LC_ALL=C sort)
    expected=$(printf '%s' "$expected" | LC_ALL=C sort)
    
    if [[ "$actual" != "$expected" ]]; then
        printf 'Unexpected output for %s\nExpected:\n%s\nActual:\n%s\n' "$*" "$expected" "$actual" >&2
        exit 1
    fi
}

assert_output "$fixture_dir/alpha.txt"$'\n'"$fixture_dir/nested/beta.txt" \
    -d "$fixture_dir" -e .txt -k needle -xd ignored

assert_output "$fixture_dir/alpha.txt:1:1"$'\n'"$fixture_dir/nested/beta.txt:1:1" \
    -d "$fixture_dir" -e .txt -k needle -c -xd ignored

assert_output "$fixture_dir/alpha.txt:1:needle"$'\n'"$fixture_dir/alpha.txt:2:token"$'\n'"$fixture_dir/nested/beta.txt:1:needle" \
    -d "$fixture_dir" -e .txt -k needle,token -l -xd ignored

assert_output "$fixture_dir/alpha.txt"$'\n'"$fixture_dir/nested/beta.txt" \
    -d "$fixture_dir" -k 'n.edle' -r -xd ignored -xe .log

assert_output "$fixture_dir${fixture_dir}alpha.txt" \
    -m fnf -p "$fixture_dir" -f alpha

assert_output "$fixture_dir${fixture_dir}large.bin" \
    -m fsf -d "$fixture_dir" -s 1MB

assert_output "$fixture_dir/abc.bin" \
    -m fff -d "$fixture_dir" -k ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad

assert_output "" -d "$fixture_dir" -k absent

for failure in missing-directory invalid-regex missing-keywords; do
    case "$failure" in
        missing-directory) args=(-d "$fixture_dir/missing") ;;
        invalid-regex) args=(-d "$fixture_dir" -k '[' -r) ;;
        missing-keywords) args=(-c) ;;
    esac
    
    if go run main.go "${args[@]}" >/dev/null 2>&1; then
        printf 'Expected failure for %s\n' "$failure" >&2
        exit 1
    fi
done

printf 'CLI fixtures passed\n'
