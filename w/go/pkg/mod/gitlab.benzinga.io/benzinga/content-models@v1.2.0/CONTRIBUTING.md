# Contributing

## Submitting a Merge Request (MR)
- Create a branch based on `origin/master`. The branch must be named using the Jira issue ID as prefix, followed by a short description of the issue. For e.g., (`CE-xxx_fixIssue`)
- [dep](https://github.com/golang/dep) is used to vendor dependencies. If the changes require updating or installing dependencies, run `dep ensure -add dependency`. The `vendor` directory is
tracked and must be committed in version control
- Ensure your code follows the [Code style](#code-style)
- Ensure your code passes tests (`make test`)
- Push your branch to the upstream repository (`git push origin CE-xxx_fixIssue`) and create a merge request by following the link returned by `git push` or
by manually creating it via https://gitlab.benzinga.io/benzinga/content-models/merge_requests/new
- The title of the MR must include the Jira ID, followed by a description of the issue. For e.g., "CE-42 Add contribution guide"

## Code Style
- [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) is used to fix imports and format code in the same style as [gofmt](https://golang.org/cmd/gofmt/)

## Testing
- [testify](https://github.com/stretchr/testify/) is used for assertions
