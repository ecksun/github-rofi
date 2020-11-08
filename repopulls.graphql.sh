#!/bin/bash

repos=("ecksun/marantz" "ecksun/Contiki-IDS")

echo 'query {'

for repo in "${repos[@]}"; do
    name="${repo##*/}"
    owner="${repo%%/*}"
    alias="${repo/-/_}"
    alias="${alias/\//_}"
cat <<EOF
  $alias: repository(owner: "$owner", name: "$name") {
    pullRequests(first: 100, states: [OPEN]) {
      nodes {
        number
        url
        state
        title
      }
    }
  }
EOF
done

# {
#   repository(owner: "ecksun", name: "marantz") {
#     pullRequests(first: 100, states: [OPEN]) {
#       nodes {
#         number
#         url
#         state
#         title
#       }
#     }
#   }

echo '}'
