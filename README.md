# eco-goinfra

## Overview
The [eco-goinfra](https://github.com/openshift-kni/eco-goinfra) is a subset of generic packages stored under public pkg 
directory that can be re-used(imported) across multiple test projects. The project is based on golang.

### Project requirements
* golang v1.19.x

# eco-goinfra - How to contribute

The project uses a development method - forking workflow
### The following is a step-by-step example of forking workflow:
1) A developer [forks](https://docs.gitlab.com/ee/user/project/repository/forking_workflow.html#creating-a-fork)
   the [eco-goinfra](https://github.com/openshift-kni/eco-goinfra) project
2) A new local feature branch is created
3) A developer makes changes on the new branch.
4) New commits are created for the changes.
5) The branch gets pushed to developer's own server-side copy.
6) Changes are tested.
7) A developer opens a pull request(`PR`) from the new branch to
   the [eco-goinfra](https://github.com/openshift-kni/eco-goinfra).
8) The pull request gets approved from at least 2 reviewers for merge and is merged into
   the [eco-goinfra](https://github.com/openshift-kni/eco-goinfra) .

### Code conventions
#### Lint
Push requested are tested in a pipeline with golangci-lint. It is advised to add [Golangci-lint integration](https://golangci-lint.run/usage/integrations/) to your development editor. It's recommended to run `make lint` before uploading a PR.

#### Functions format
If the function's arguments fit in a single line - use the following format:
```go
func Function(argInt1, argInt2 int, argString1, argString2 string) {
    ...
}
```

If the function's arguments don't fit in a single line - use the following format:
```go
func Function(
    argInt1 int,
    argInt2 int,
    argInt3 int,
    argInt4 int,
    argString1 string,
    argString2 string,
    argString3 string,
    argString4 string) {
    ...
}
```
One more acceptable format example:
```go
func Function(
    argInt1, argInt2 int, argString1, argString2 string, argSlice1, argSlice2 []string) {
	
}
```