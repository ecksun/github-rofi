.data.requests.nodes + .data.created.nodes + .data.mentions.nodes, .data.assigned.nodes
    | sort_by(.createdAt) | reverse
    | .[] | ((.repository.nameWithOwner + "/pull/" + (.number | tostring)) + ": " + .title)
