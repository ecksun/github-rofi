#!/bin/bash

pull_cache="$HOME/.cache/github-rofi/gitlab-pulls.json"

./update_cache.sh gitlab

rows() {
    jq --raw-output '.[] | [.reference, .branch, .title] | @tsv' < "$pull_cache" | column -t --separator $'\t'
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
    ./update_cache.sh gitlab --force
    exit 0
fi

[[ $hit =~ ([^/]+)/([^!]+)!([0-9]+)[[:space:]]+([^[:space:]]+) ]]

export owner="${BASH_REMATCH[1]}"
export repo="${BASH_REMATCH[2]}"
export pr="${BASH_REMATCH[3]}"
export branch="${BASH_REMATCH[4]}"
