package kmm

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	moduleV1Beta1 "github.com/rh-ecosystem-edge/kernel-module-management/api/v1beta1"
	v1 "k8s.io/api/core/v1"
)

// ModuleLoaderContainerBuilder provides struct for the module object containing the ModuleLoaderContainerSpec
// definitions.
type ModuleLoaderContainerBuilder struct {
	// ModuleLoaderContainerBuilder definition. Used to create a Module object.
	definition *moduleV1Beta1.ModuleLoaderContainerSpec
	// errorMsg is processed before the Module object is created.
	errorMsg string
}

// ModuleLoaderContainerAdditionalOptions additional options for ModuleLoaderContainer object.
type ModuleLoaderContainerAdditionalOptions func(
	builder *ModuleLoaderContainerBuilder) (*ModuleLoaderContainerBuilder, error)

// NewModLoaderContainerBuilder creates a new instance of ModuleLoaderContainerBuilder.
func NewModLoaderContainerBuilder(modName string) *ModuleLoaderContainerBuilder {
	glog.V(100).Infof(
		"Initializing new ModuleLoaderContainerBuilder structure with following params: %s", modName)

	builder := &ModuleLoaderContainerBuilder{
		definition: &moduleV1Beta1.ModuleLoaderContainerSpec{
			Modprobe: moduleV1Beta1.ModprobeSpec{
				ModuleName: modName,
			},
		},
	}

	if modName == "" {
		glog.V(100).Infof("The modName of the NewModLoaderContainerBuilder is empty")

		builder.errorMsg = "'modName' cannot be empty"
	}

	return builder
}

// WithModprobeSpec adds the specified Modprobe to the ModuleLoaderContainerBuilder.
func (builder *ModuleLoaderContainerBuilder) WithModprobeSpec(
	dirName, fwPath string, parameters, args, rawargs []string) *ModuleLoaderContainerBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating new ModuleLoaderContainerBuilder structure with following modprob params. "+
			"DirName: %s, FirmwarePath: %s, Parameters: %v", dirName, fwPath, parameters)

	builder.definition.Modprobe.DirName = dirName
	builder.definition.Modprobe.FirmwarePath = fwPath
	builder.definition.Modprobe.Parameters = parameters
	builder.definition.Modprobe.Args.Load = args
	builder.definition.Modprobe.RawArgs.Load = rawargs

	return builder
}

// WithKernelMapping adds the specified KernelMapping to the ModuleLoaderContainerBuilder.
func (builder *ModuleLoaderContainerBuilder) WithKernelMapping(
	mapping *moduleV1Beta1.KernelMapping) *ModuleLoaderContainerBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating new ModuleLoaderContainerBuilder structure with following KernelMapping %v", mapping)

	if mapping == nil {
		glog.V(100).Infof("The mapping is undefined")

		builder.errorMsg = "'mapping' can not be empty nil"

		return builder
	}

	builder.definition.KernelMappings = append(builder.definition.KernelMappings, *mapping)

	return builder
}

// WithImagePullPolicy adds the specified ImagePullPolicy to the ModuleLoaderContainerBuilder.
func (builder *ModuleLoaderContainerBuilder) WithImagePullPolicy(policy string) *ModuleLoaderContainerBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating new ModuleLoaderContainerBuilder structure with following policy %v", policy)

	if policy == "" {
		builder.errorMsg = "'policy' can not be empty"

		return builder
	}

	builder.definition.ImagePullPolicy = v1.PullPolicy(policy)

	return builder
}

// WithOptions creates ModuleLoaderContainer with generic mutation options.
func (builder *ModuleLoaderContainerBuilder) WithOptions(
	options ...ModuleLoaderContainerAdditionalOptions) *ModuleLoaderContainerBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting ModuleLoaderContainer additional options")

	for _, option := range options {
		if option != nil {
			builder, err := option(builder)

			if err != nil {
				glog.V(100).Infof("Error occurred in mutation function")

				builder.errorMsg = err.Error()

				return builder
			}
		}
	}

	return builder
}

// BuildModuleLoaderContainerCfg returns ModuleLoaderContainerSpec struct.
func (builder *ModuleLoaderContainerBuilder) BuildModuleLoaderContainerCfg() (
	*moduleV1Beta1.ModuleLoaderContainerSpec, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Returning the ModuleLoaderContainerBuilder structure %v", builder.definition)

	return builder.definition, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ModuleLoaderContainerBuilder) validate() (bool, error) {
	if builder == nil {
		glog.V(100).Infof("The builder is uninitialized")

		return false, fmt.Errorf("error: received nil builder")
	}

	if builder.definition == nil {
		glog.V(100).Infof("The moduleloadercontainer is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("ModuleLoaderContainer")
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The builder has error message: %s", builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}

// DevicePluginContainerBuilder provides struct for the module object containing the DevicePluginContainerSpec
// definitions.
type DevicePluginContainerBuilder struct {
	// DevicePluginContainerBuilder definition. Used to create a Module object.
	definition *moduleV1Beta1.DevicePluginContainerSpec
	// object is created.
	errorMsg string
}

// NewDevicePluginContainerBuilder creates DevicePluginContainerSpec based on given arguments and mutation functs.
func NewDevicePluginContainerBuilder(image string) *DevicePluginContainerBuilder {
	glog.V(100).Infof(
		"Initializing new DevPluginContainerBuilder structure with the following params: %s", image)

	builder := DevicePluginContainerBuilder{
		definition: &moduleV1Beta1.DevicePluginContainerSpec{
			Image: image,
		},
	}

	if image == "" {
		glog.V(100).Infof("The image of NewDevicePluginContainerBuilder is empty")

		builder.errorMsg = "invalid parameter 'image' cannot be empty"
	}

	return &builder
}

// WithEnv adds specific env to DevicePlugin Container.
func (builder *DevicePluginContainerBuilder) WithEnv(name, value string) *DevicePluginContainerBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating new DevPluginContainerBuilder structure with following Env. Name: %s, Value: %s", name, value)

	if name == "" {
		glog.V(100).Infof("The name of WithEnv is empty")

		builder.errorMsg = "'name' can not be empty for DevicePlugin Env"
	}

	if value == "" {
		glog.V(100).Infof("The value of WithEnv is empty")

		builder.errorMsg = "'value' can not be empty for DevicePlugin Env"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.definition.Env = append(builder.definition.Env, v1.EnvVar{Name: name, Value: value})

	return builder
}

// WithVolumeMount adds VolumeMount to DevicePlugin Container.
func (builder *DevicePluginContainerBuilder) WithVolumeMount(mountPath, name string) *DevicePluginContainerBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating new DevPluginContainerBuilder structure with mountPath Env. Name: %s, MountPath: %s",
		name, mountPath)

	if name == "" {
		glog.V(100).Infof("The name of WithVolumeMount is empty")

		builder.errorMsg = "'name' can not be empty for DevicePlugin mountPath"
	}

	if mountPath == "" {
		glog.V(100).Infof("The mountPath of WithVolumeMount is empty")

		builder.errorMsg = "'mountPath' can not be empty for DevicePlugin mountPath"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.definition.VolumeMounts = append(
		builder.definition.VolumeMounts, v1.VolumeMount{Name: name, MountPath: mountPath})

	return builder
}

// GetDevicePluginContainerConfig returns DevicePluginContainerSpec with needed configuration.
func (builder *DevicePluginContainerBuilder) GetDevicePluginContainerConfig() (
	*moduleV1Beta1.DevicePluginContainerSpec, error) {
	if valid, err := builder.validate(); !valid {
		return nil, fmt.Errorf("error building DevicePluginContainerSpec config due to :%w", err)
	}

	return builder.definition, nil
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *DevicePluginContainerBuilder) validate() (bool, error) {
	if builder == nil {
		glog.V(100).Infof("The builder is uninitialized")

		return false, fmt.Errorf("error: received nil builder")
	}

	if builder.definition == nil {
		glog.V(100).Infof("The deviceplugincontainer is undefined")

		builder.errorMsg = msg.UndefinedCrdObjectErrString("DevicePluginContainer")
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The builder has error message: %s", builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
