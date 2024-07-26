# eco-goinfra
[![Test Incoming Changes](https://github.com/openshift-kni/eco-goinfra/actions/workflows/makefile.yml/badge.svg)](https://github.com/openshift-kni/eco-goinfra/actions/workflows/makefile.yml)
[![Unit Test Coverage](https://raw.githubusercontent.com/openshift-kni/eco-goinfra/badges/.badges/main/coverage.svg)](https://github.com/openshift-kni/eco-goinfra/actions/workflows/makefile.yml)
[![license](https://img.shields.io/github/license/openshift-kni/eco-goinfra?color=blue&labelColor=gray&logo=apache&logoColor=lightgray&style=flat)](https://github.com/openshift-kni/eco-goinfra/blob/master/LICENSE)

## Overview
The [eco-goinfra](https://github.com/openshift-kni/eco-goinfra) project contains a collection of generic [packages](./pkg) that can be used across various test projects.

### Project requirements
* golang v1.22.x

## Usage
In order to re-use code from this project you need to import relevant package/packages in to your project code.

```go
import "github.com/openshift-kni/eco-goinfra/pkg/NAME_OF_A_NEEDED_PACKAGE"
```
In addition, you need to add link to the project github.com/openshift-kni/eco-goinfra to your local go.mod file

```go
require(
    github.com/openshift-kni/eco-goinfra latest
)
```



### Clients package:
The [clients](./pkg/clients) package contains several api clients combined into a single struct.
The New function of the clients package returns a ready connection to the cluster api.
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
func NewBuilder() or New[ObjectName]Builder() // Initiates object struct. This function require minimum set of parameters that are required to create the object on a cluster.
func Pull() or Pull[ObjectName]() // Pulls existing object to struct.
func Create()  // Creates new object on cluster if it does not exist.
func Delete() // Removes object from cluster if it exists.
func Update() // Updates object based on new object's definition.
func Exist() // Returns bool if object exist.
func With***() // Set of mutation functions that can mutate any part of the object. 
```
Please refer to [namespace](./usage/namespace/namespace.go) example for more info.

### Validator Method
In order to ensure safe access to objects and members, each builder struct should include a `validate` method. This method should be invoked inside packages before accessing potentially uninitialized code to mitigate unintended errors. Example:
```go
func (builder *Builder) WithMethod(someString string) *Builder {
    if valid, _ := builder.validate(); !valid {
        return builder
    }
    
    glog.V(100).Infof(
        "Updating builder %s in namespace %s with the string: %s",
        builder.Definition.Name, builder.Definition.Namespace, someString
    )
    
    builder.Definition.StringHolder = someString
    
    return builder
}
```
Typically, validate methods will check that pointers are not nil and that errorMsg has not been set. Here is an example of how the secret package validate method ensures that Builder.apiClient has properly been initialized before being called:
```go
func main() {
	apiClient := clients.New("bad api client")

	_, err := secret.NewBuilder(
        apiClient, "mysecret", "mynamespace", v1SecretTypeDockerConfigJson).Create()
	if err != nil {
		log.Fatal(err)
	}
}
```
Instead of causing a panic, the method will return a proper error message:
```
2023/06/16 11:55:58 Loading kube client config from path "bad api client"
2023/06/16 11:55:58 Secret builder cannot have nil apiClient
exit status 1
```
Please refer to the [secret pkg](./pkg/secret/secret.go)'s use of the validate method for more information.

### BMC Package
The BMC package can be used to access the BMC's Redfish API, run BMC's CLI commands, or get the systems' serial console. Only the host must be provided in `New()` while Redfish and SSH credentials, along with other options, can be configured using separate methods.

```go
bmc := bmc.New("1.2.3.4").
    WithRedfishUser("redfishuser1", "redfishpass1").
    WithSSHUser("sshuser1", "sshpass1").
    WithSSHPort(1234)
```

You can check an example program for the BMC package in [usage](usage/bmc/bmc.go).

#### BMC's Redfish API
The access to BMC's Redfish API is done by methods that encapsulate the underlaying HTTP calls made by the external gofish library. The redfish system index is defaulted to 0, but it can be changed with `SetSystemIndex()`:
```
const systemIndex = 3
err = bmc.SetSystemIndex(systemIndex)
if err != nil {
    ...
}

manufacturer, err := bmc.SystemManufacturer()
if err != nil {
    ...
}

fmt.Printf("System %d's manufacturer: %v", systemIndex, manufacturer)

```

#### BMC's CLI
The method `RunCLICommand` has been implemented to run CLI commands.
```
func (bmc *BMC) RunCLICommand(cmd string, combineOutput bool, timeout time.Duration) stdout string, stderr string, err error)
```
This method is not interactive: it blocks the caller until the command ends, copying its output into stdout and stderr strings.

#### Serial Console
The method `OpenSerialConsole` can be used to get the systems's serial console, which is tunneled in the an underlaying SSH session.
```
func (bmc *BMC) OpenSerialConsole(openConsoleCliCmd string) (io.Reader, io.WriteCloser, error)
```
The user gets a (piped) reader and writer interfaces in order to read the output or write custom input (like CLI commands) in a interactive fashion.
A use case for this is a test case that needs to wait for some pattern to appear in the system's serial console after rebooting the system.

The `openConsoleCliCmd` is the command that will be sent to the BMC's (SSH'd) CLI to open the serial console. In case the user doesn't know the command,
it can be left empty. In that case, there's a best effort mechanism that will try to guess the CLI command based on the system's manufacturer, which will
be internally retrieved using the Redfish API.

It's important to close the serial console using the method `bmc.CloseSerialConsole()`, which closes the underlying SSH session. Otherwise, BMC's can reach
the maximum number of concurrent SSH sessions making other (SSH'd CLI) commands to fail. See an example program [here](usage/bmc/bmc.go).

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
#### Note: Every new package requires a coverage of <ins>ALL</ins> its public functions with unit tests. Unit tests are located in the same package as the resource, in a file with the name *resource*_test.go. Examples can be found in [configmap_test.go](./pkg/configmap/configmap_test.go) and [metallb_test.go](./pkg/metallb/metallb_test.go).

### Code conventions
#### Lint
Push requested are tested in a pipeline with golangci-lint. It is advised to add [Golangci-lint integration](https://golangci-lint.run/usage/integrations/) to your development editor. It is recommended to run `make lint` before uploading a PR.

#### Functions format
If the function's arguments fit in a single line - use the following format:
```go
func Function(argInt1, argInt2 int, argString1, argString2 string) {
    ...
}
```

If the function's arguments do not fit in a single line - use the following format:
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

### Sync Tool and Operator Types

Previously, packages relied on importing the types from individual operators and adding them as dependencies. This has led to eco-goinfra having many dependencies, making it harder to reuse and increasing the chance of version conflicts between indirect dependencies.

By using just the runtime client and removing dependencies on individual operators, the sync tool reduces the number of eco-goinfra dependencies and makes packages more modular. It will periodically copy over types from operator repos to this repo by cloning the desired path in the operator repo, applying changes specified in the config file, and then copying the resulting types into this repo.

#### Running

To run the sync tool, use the `lib-sync` makefile target.

```
make lib-sync
```

It may be faster to sync only one config file instead using the `--config-file` flag.

```
go run ./internal/sync --config-file ./internal/sync/configs/<config-file.yaml>
```

If the sync fails while adding a new set of operator types, remove the synced directory from `schemes/<pkg-to-sync>` and rerun the sync.

#### Configuration

Config files for the sync tool live in the [internal/sync/configs](./internal/sync/configs/) directory. A good example of all the features available is in the [nvidia-config.yaml](./internal/sync/configs/nvidia-config.yaml) file.

Each config is a YAML document with a top level list of objects that follow this form:

```yaml
- name: operator-repo # 1
  sync: true # 2
  repo_link: "https://github.com/operator/repo" # 3
  branch: main # 4
  remote_api_directory: pkg/apis/v1 # 5
  local_api_directory: schemes/operator/operatortypes # 6
  replace_imports: # 7
    - old: '"github.com/operator/repo/pkg/apis/v1/config"' # 8
      new: '"github.com/openshift-kni/eco-goinfra/pkg/schemes/operator/operatortypes/config"'
    - old: '"github.com/operator/repo/pkg/apis/v1/utils/exec"'
      new: exec "github.com/openshift-kni/eco-goinfra/pkg/schemes/operator/operatortypes/executils" # 9
  excludes:
    - "*_test.go" # 10
```

1. Name, which does not have to be unique, identifies the repos in logs and controls where it is cloned during the sync. It should be named using only alphanumeric characters and hyphens, although the name itself does not affect how the repo gets synced.
2. Sync is whether the operator repo will be synced both periodically and when `make lib-sync` is used manually.
3. Repo link is the url of the operator repo itself. It does not need to end in `.git`.
4. Branch is the branch of the operator repo to sync with.
5. Remote API directory is the path in the operator repo to sync with, relative to the operator repo root.
6. Local API directory is the path in eco-goinfra where the remote API directory should be synced to. Relative to the eco-goinfra root, it should start with `schemes/`.
7. The operator may import code from other parts of its repo or even other repos. These imports must be updated to new paths when synced to eco-goinfra. Replace imports is an optional field to allow updating these paths during the sync.
8. Import replacement is done through find and replace, so we include the double quotes to make sure it is an import being matched.
9. If the package name in eco-goinfra is different than the operator repo, the import should be renamed so the code still works.
10. Excludes is an optional list of file patterns to exclude. Since tests and mocks may add their own dependencies, excluding them can reduce how many other dependencies need to be synced.

Like in the [nvidia-config.yaml](./internal/sync/configs/nvidia-config.yaml) example, it is often the case that one repo will import a few others. All of the imported repos should be specified in the sync config to avoid adding new dependencies.

#### Using the Synced Types

Instead of directly adding the scheme synced from the operator repo to the [clients](./pkg/clients/) package, the client scheme should be updated in `NewBuilder()`, `Pull()`, and any `List***()` functions. The [metallb](./pkg/metallb/metallb.go) package provides a good example of this.

```go
// in NewBuilder
err := apiClient.AttachScheme(mlbtypes.AddToScheme)
if err != nil {
	glog.V(100).Infof("Failed to add metallb scheme to client schemes")

	return nil
}
```

[metallb_test.go](./pkg/metallb/metallb_test.go) provides an example of how to use the schemes in unit tests. There should be a variable for the test schemes.

```go
var mlbTestSchemes = []clients.SchemeAttacher{
	mlbtypes.AddToScheme,
}
```

And these schemes should be provided when creating the test client.

```go
// in buildMetalLbTestClientWithDummyObject
clients.GetTestClients(clients.TestClientParams{
	K8sMockObjects:  buildDummyMetalLb(),
	SchemeAttachers: mlbTestSchemes,
})
```
