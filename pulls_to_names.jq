.data.viewer.organizations.nodes[].repositories.nodes[]
    | {name: .nameWithOwner, pull: .pullRequests.nodes[]}
    | {name: (.name + "/pull/" + (.pull.number | tostring)), pull: .pull }
    | (.name + ": " + .pull.title)
