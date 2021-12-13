#!/bin/bash

pull_cache="$HOME/.cache/github-rofi/pulls.json"
cache_meta="$HOME/.cache/github-rofi/metadata"

cd "$(dirname "$(readlink "${BASH_SOURCE[0]}")")" || exit 1

update_cache() {
    notify-send --hint=string:synchronous:github-rofi github-rofi "Updating cache.."
    mkdir -p "$(dirname "$pull_cache")"
    ./graphql.sh > "$pull_cache.tmp"
    mv "$pull_cache.tmp" "$pull_cache"
    touch "$cache_meta"
    sed -i '/^pull_time/d' "$cache_meta"
    echo "pull_time $(date --rfc-3339=s)" >> "$cache_meta"
    notify-send --hint=string:synchronous:github-rofi --expire-time=3000 github-rofi "Cache updated!"
}

if [ "$1" == "--force" ]; then
    update_cache
    exit 0
fi

if [ ! -e "$pull_cache" ]; then
    echo >&2 "No pull cache found, populating it"
    update_cache
fi

if [ -e "$cache_meta" ]; then
    pull_time=$(awk '/pull_time/ { $1=""; print $0 }' "$cache_meta")
fi

outdated_cache="$(date --date "now - 180 min" --rfc-3339=s)"
if [ -z "$pull_time" ] || [[ $outdated_cache > $pull_time ]]; then
    echo >&2 "Cache is outdated, updating now"
    update_cache
    pull_time="$(date --rfc-3339=s)"
fi

old_cache="$(date --date "now - 10 min" --rfc-3339=s)"
if [[ $old_cache > $pull_time ]]; then
    echo >&2 "Cache is somewhat old, updating in background"
    update_cache &
    disown
fi

