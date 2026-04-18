#!/usr/bin/env bash
# check-worktrees.sh — enforce the discipline described in CLAUDE.md §5.
#
# For every git worktree on this host:
#   - lists its branch and HEAD
#   - checks whether the branch has an open PR on origin
#   - warns if the worktree is older than 24h and has no open PR
#
# Exit 0 always — this is an advisory, not a gate.

set -u

# Colour helpers: emit ANSI only when stdout is a TTY.
if [ -t 1 ]; then
    RED=$'\e[31m'
    YEL=$'\e[33m'
    GRN=$'\e[32m'
    DIM=$'\e[2m'
    RST=$'\e[0m'
else
    RED=""; YEL=""; GRN=""; DIM=""; RST=""
fi

have_gh() { command -v gh >/dev/null 2>&1; }

# Resolve the default branch of origin (falls back to "main").
default_branch() {
    git symbolic-ref --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|origin/||' || echo main
}

DEFAULT_BRANCH=$(default_branch)
NOW=$(date +%s)
DAY_SECS=$((24 * 60 * 60))

warnings=0
entries=0

while read -r line; do
    # `git worktree list --porcelain` emits blocks separated by blank lines.
    # We read in non-porcelain mode for simplicity: "<path> <head> [branch]".
    [ -z "$line" ] && continue
    entries=$((entries + 1))

    path=$(echo "$line" | awk '{print $1}')
    head=$(echo "$line" | awk '{print $2}')
    branch=$(echo "$line" | awk '{print $3}' | tr -d '[]')

    if [ -z "$branch" ] || [ "$branch" = "$DEFAULT_BRANCH" ]; then
        echo "  ${GRN}✓${RST} ${path} ${DIM}(${head} on ${branch:-detached})${RST}"
        continue
    fi

    # Age (modification time of the worktree dir).
    age_secs=0
    if [ -d "$path" ]; then
        mtime=$(stat -c %Y "$path" 2>/dev/null || stat -f %m "$path" 2>/dev/null || echo 0)
        age_secs=$((NOW - mtime))
    fi
    age_h=$((age_secs / 3600))

    pr_state="unknown"
    if have_gh; then
        # Look up any PR from this branch.
        pr_state=$(gh pr list --head "$branch" --state open --json number --limit 1 2>/dev/null \
            | python3 -c "import json, sys; d=json.load(sys.stdin); print('open' if d else 'none')" 2>/dev/null \
            || echo "unknown")
    fi

    if [ "$pr_state" = "open" ]; then
        echo "  ${GRN}✓${RST} ${path} ${DIM}(branch ${branch}, PR open, age ${age_h}h)${RST}"
    elif [ "$age_secs" -gt "$DAY_SECS" ]; then
        echo "  ${RED}!${RST} ${path} ${YEL}(branch ${branch}, NO open PR, age ${age_h}h — close or PR)${RST}"
        warnings=$((warnings + 1))
    else
        echo "  ${YEL}~${RST} ${path} ${DIM}(branch ${branch}, no PR yet, age ${age_h}h)${RST}"
    fi
done < <(git worktree list)

if [ "$entries" -eq 0 ]; then
    echo "  ${DIM}(no worktrees found — are you inside a git repo?)${RST}"
fi

if [ "$warnings" -gt 0 ]; then
    echo ""
    echo "${YEL}${warnings} worktree(s) need attention. Per CLAUDE.md §5: each worktree must cycle (develop → PR → merge → delete) within one day.${RST}"
fi

exit 0
