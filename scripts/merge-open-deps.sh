#!/usr/bin/env bash
# merge-open-deps.sh — sweep and merge every open dependabot PR on this
# repo. Safe to run on a cron or at the start of every Claude session.
#
# Usage:
#   bash scripts/merge-open-deps.sh
#
# Behaviour:
#   - For each open PR authored by `app/dependabot`:
#       * if the PR is MERGEABLE and not CONFLICTING, invoke
#         scripts/merge-pr.sh <n>
#       * if it is CONFLICTING, post `@dependabot rebase`, wait up to
#         3 minutes for a rebase, then try again once
#       * otherwise, skip and print the reason
#   - Exits 0 when every dependabot PR has been handled (merged, skipped,
#     or reported).

set -euo pipefail

REPO="llmstatus/llmstatus"
HERE=$(cd "$(dirname "$0")" && pwd)

want_rebase() {
    local n=$1
    echo "[deps] #$n: CONFLICTING — requesting @dependabot rebase"
    gh pr comment "$n" --repo "$REPO" --body "@dependabot rebase" >/dev/null
    for i in $(seq 1 18); do
        sleep 10
        state=$(gh pr view "$n" --repo "$REPO" --json mergeable --jq '.mergeable')
        if [ "$state" = "MERGEABLE" ]; then
            echo "[deps] #$n: rebased after ${i}x10s"
            return 0
        fi
    done
    echo "[deps] #$n: still not mergeable after 3min — skipping"
    return 1
}

prs=$(gh pr list --repo "$REPO" --state open --author app/dependabot \
    --json number,mergeable,mergeStateStatus,title --jq '.[]')

if [ -z "$prs" ]; then
    echo "[deps] no open dependabot PRs."
    exit 0
fi

echo "$prs" | python3 -c "
import json, sys
for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    d = json.loads(line)
    print(f\"{d['number']}\t{d['mergeable']}\t{d['title']}\")
" | while IFS=$'\t' read -r n state title; do
    echo ""
    echo "[deps] #$n ($state): $title"
    case "$state" in
        MERGEABLE)
            bash "$HERE/merge-pr.sh" "$n" "auto-merge: dependency bump" || true
            ;;
        CONFLICTING)
            if want_rebase "$n"; then
                bash "$HERE/merge-pr.sh" "$n" "auto-merge: dependency bump (rebased)" || true
            fi
            ;;
        *)
            echo "[deps] #$n: state=$state — skipping"
            ;;
    esac
done

echo ""
echo "[deps] sweep complete."
