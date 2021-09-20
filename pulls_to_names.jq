.data | to_entries | map(.value.nodes) | add # Merge searches
    | group_by(.url) | map(.[0]) # deduplicate by URL
    | sort_by(.createdAt) | reverse
    | .[] | ((.repository.nameWithOwner + "/pull/" + (.number | tostring)) + ": " + .title)
