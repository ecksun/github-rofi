#!/bin/bash

forge="${1:?Missing argument forge}"
owner="${2:?Missing argument owner}"
repo="${3:?Missing argument repo}"
branch="${4:?Missing argument branch}"

get_folder() {
    owner="$1"
    repo="$2"

    echo "$HOME/code/$owner/$repo"
}

get_worktree() {
    owner="$1"
    repo="$2"
    branch="$3"

    repo_dir="$(get_folder "$owner" "$repo")"
    worktree="${branch//\//-}"

    echo "$repo_dir.$worktree"
}

repo_dir="$(get_folder "$owner" "$repo")"

initfile=$(mktemp --suffix "$forge-pr-init")
trap 'rm "$initfile"' EXIT

cat << EOF > "$initfile"
#!/bin/bash
( # run all git setup in a subshell to be able to exit on error
set -euo pipefail
EOF
chmod +x "$initfile"

if ! [ -d "$repo_dir" ]; then
    cat << EOF >> "$initfile"
# Can't find repository, cloning git@$forge/$repo into $repo_dir
echo + git clone "git@$forge:$owner/$repo" "$repo_dir"
git clone "git@$forge:$owner/$repo" "$repo_dir"
EOF
fi

worktree_dir="$(get_worktree "$owner" "$repo" "$branch")"

if ! git -C "$repo_dir" show-ref --verify --quiet "refs/remotes/origin/$branch"; then
    cat << EOF >> "$initfile"
# Remote branch doesn't exist, fetching..
echo + git -C "$repo_dir" fetch origin
git -C "$repo_dir" fetch origin
EOF
fi

if ! [ -d "$worktree_dir" ]; then
    cat << EOF >> "$initfile"
# No worktree exists, creating it..
if git -C "$repo_dir" rev-parse --quiet --verify "$branch" > /dev/null; then
  echo + git -C "$repo_dir" worktree add --guess-remote "$worktree_dir" "$branch"
  git -C "$repo_dir" worktree add --guess-remote "$worktree_dir" "$branch"
else
  echo + git -C "$repo_dir" worktree add --track -b "$branch" "$worktree_dir" "origin/$branch"
  git -C "$repo_dir" worktree add --track -b "$branch" "$worktree_dir" "origin/$branch"
fi
EOF
fi

cat << EOF >> "$initfile"
echo + git -C "$repo_dir" fetch origin "$branch"
git -C "$repo_dir" fetch origin "$branch"
)

# load files normally loaded without --rcfile
. /etc/bash.bashrc
. ~/.bashrc
if [ -d "$worktree_dir" ]; then
    cd "$worktree_dir"
else
    cd "$repo_dir"
fi
EOF

kitty bash --rcfile "$initfile"
