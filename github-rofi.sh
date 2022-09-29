#!/bin/bash

pull_cache="$HOME/.cache/github-rofi/github-pulls.json"

cd "$(dirname "$(readlink "${BASH_SOURCE[0]}")")" || exit 1
./update_cache.sh github

rows() {
    jq --from-file pulls_to_branches.jq --raw-output < "$pull_cache" | \
    column -t --separator $'\t'
}

hit=$(
    (
        rows && \
        echo "refresh"
    ) | rofi -width 70 -dmenu -theme Arc-Dark -i)
if [ -z "$hit" ]; then
    exit 0
fi
if [ "$hit" == "refresh" ]; then
    ./update_cache.sh github --force
    exit 0
fi

[[ $hit =~ ([^/]+)/([^[:space:]]+)[[:space:]]+([0-9]+)[[:space:]]+([^[:space:]]+) ]]

export owner="${BASH_REMATCH[1]}"
export repo="${BASH_REMATCH[2]}"
export pr="${BASH_REMATCH[3]}"
export branch="${BASH_REMATCH[4]}"
