#!/bin/bash

config_dir="$HOME/.config/gitforge-rofi"

custom='[]'
if [ -f "$config_dir/gitlab-custom-query" ]; then
    custom=$(curl -s \
        --header "Authorization: Bearer $(< ~/.config/gitforge-rofi/token)" \
        "$(<"$config_dir/gitlab-custom-query")"
    )
fi

created=$(curl -s \
    --header "Authorization: Bearer $(< ~/.config/gitforge-rofi/token)" \
    'https://gitlab.com/api/v4/merge_requests?state=opened&scope=created_by_me'
)

assigned=$(curl -s \
    --header "Authorization: Bearer $(< ~/.config/gitforge-rofi/token)" \
    'https://gitlab.com/api/v4/merge_requests?state=opened&scope=assigned_to_me'
)

jq --slurp '(.[0] + .[1]  + .[2]) | map({ title: .title, url:  .web_url, created: .created_at, reference: .references.full, branch: .source_branch })' <(echo "$created") <(echo "$assigned") <(echo "$custom")
