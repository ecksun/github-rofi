#!/bin/bash

token=$(<"$HOME/.config/github-rofi/token")
username=$(<"$HOME/.config/github-rofi/username")

data=$(jq --arg query "$(cat orgpulls.graphql)" \
    --null-input \
    '{ query: $query }')

time curl \
    --silent \
    -u "$username:$token" \
    -H "Accept: application/vnd.github.v3+json" \
    --data "$data" \
    "https://api.github.com/graphql" | jq .
