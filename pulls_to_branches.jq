.data | to_entries | map(.value.nodes) | add # Merge searches
    | group_by(.url) | map(.[0]) # deduplicate by URL
    | sort_by(.createdAt) | reverse
    | .[] | [.repository.nameWithOwner, (.number | tostring), .headRef.name[0:50], .title] | @tsv
