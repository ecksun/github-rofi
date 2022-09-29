#!/bin/bash

forge="$1"

pull_cache="$HOME/.cache/github-rofi/$forge-pulls.json"
cache_meta="$HOME/.cache/github-rofi/$forge-metadata"

cd "$(dirname "$(readlink "${BASH_SOURCE[0]}")")" || exit 1

update_cache() {
    notify-send --hint="string:synchronous:$forge-rofi" "$forge-rofi" "Updating cache for $forge.."
    mkdir -p "$(dirname "$pull_cache")"
    "./$forge-fetch.sh" > "$pull_cache.tmp"
    mv "$pull_cache.tmp" "$pull_cache"
    touch "$cache_meta"
    sed -i '/^pull_time/d' "$cache_meta"
    echo "pull_time $(date --rfc-3339=s)" >> "$cache_meta"
    notify-send --hint="string:synchronous:$forge-rofi" --expire-time=3000 "$forge-rofi" "Cache updated for $forge!"
}

if [ "$2" == "--force" ]; then
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

