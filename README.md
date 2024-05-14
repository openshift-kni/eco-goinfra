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
The BMC package can be used to access the BMC's Redfish API, run BMC's CLI commands or getting the systems' serial console. Credentials for both Redfish and SSH user, the SSH port and timeouts need to be passed as parameters of the New() method. E.g.

```
redfishUser := bmc.User{Name: "redfishuser1", Password: "redfishpass1"}
sshUser := bmc.User{Name: "sshuser1", Password: "sshpass1"}
timeOuts := bmc.TimeOuts{Redfish: 10*time.Second, SSH: 10*time.Second}

bmc, err := bmc.New("1.2.3.4", redfishUser, sshUser, 22, timeOuts)
```

You can check an example program for the BMC package [here](usage/bmc/bmc.go).

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
