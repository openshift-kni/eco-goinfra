# eco-goinfra

## Overview
The [eco-goinfra](https://github.com/openshift-kni/eco-goinfra) project contains a collection of generic [packages](./pkg) that can be used across various test projects.

### Project requirements
* golang v1.19.x

## Usage
In order to re-use code from this project you need to import relevant package/packages in to your project code.

```go
import "github.com/openshift-kni/eco-goinfra/pkg/NAME_OF_A_NEEDED_PACKAGE"
```

### Clients package:
[clients](./pkg/clients) package contains several api clients combined in to the single struct.
The New function of client package returns ready connection to cluster api.
If the path to kubeconfig is not specified to the new function then the KUBECONFIG environment variable is used.
In case of failure client.New("") returns nil.
```go
import "github.com/openshift-kni/eco-goinfra/pkg/clients"

func main() {
    apiClients := clients.New("")
	
    if apiClients == nil {
        panic("Failed to load api client")
        }
    }
)
```
[Client usage example](./usage/client/client.go)

### Cluster Objects
Every cluster object namespace, configmap, daemonset, deployment and other has its own package under [packages](./pkg) directory.
The structure of any object has common interface:
```go
func NewBuilder() or New[ObjectName]Builder() // Initiates object struct. This function require minimum set of parameters that allow to create object on a cluster.
func Pull() or Pull[ObjectName]() // Pulls existing object to struct.
func Create()  // Creates new object on cluster if it doesn't exist.
func Delete() // Removes object from cluster if it exists.
func Update() // Updates object based on new object's definition.
func Exist() // Returns bool if object exist.
func With***() // Set of mutiation functions that can mutate any part of the object. 
```
Please refer to [namespace](./usage/namespace/namespace.go) example for more info.


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