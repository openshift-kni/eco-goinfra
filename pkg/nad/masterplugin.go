package nad

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	"k8s.io/utils/strings/slices"
)

var (
	// allowedMacVlanMode represents all allowed modes for macvlan plugin type.
	allowedMacVlanMode      = []string{"bridge", "passthru", "private", "vepa"}
	invalidIpamParameterMsg = "invalid ipam parameter"
)

// MasterMacVlanPlugin provides struct for NetworkAttachmentDefinition Master plugin with macvlan configuration.
type MasterMacVlanPlugin struct {
	masterPlugin *MasterPlugin
	errorMsg     string
}

// NewMasterMacVlanPlugin creates new instance of MasterMacVlanPlugin.
func NewMasterMacVlanPlugin(name string) *MasterMacVlanPlugin {
	glog.V(100).Infof(
		"Initializing new MasterVlanPlugin structure with the following param: %s", name)

	builder := &MasterMacVlanPlugin{
		masterPlugin: &MasterPlugin{
			CniVersion: "0.3.1",
			Name:       name,
			Type:       "macvlan",
		},
	}

	if builder.masterPlugin.Name == "" {
		glog.V(100).Infof("error MasterMacVlanPlugin can not be empty")

		builder.errorMsg = "MasterMacVlanPlugin name is empty"

		return builder
	}

	return builder
}

// WithMode defines macvlan type to MasterMacVlanPlugin. Default is bridge.
func (plugin *MasterMacVlanPlugin) WithMode(mode string) *MasterMacVlanPlugin {
	glog.V(100).Infof("Adding macvlan mode %s to MasterMacVlanPlugin", mode)

	if !slices.Contains(allowedMacVlanMode, mode) {
		glog.V(100).Infof("error to add mode %s, allowed modes are %v", mode, allowedMacVlanMode)

		plugin.errorMsg = "invalid mode parameter"
	}

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterMacVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterMacVlanPlugin")

		return plugin
	}

	plugin.masterPlugin.Mode = mode

	return plugin
}

// WithMasterInterface defines master interface to MasterMacVlanPlugin. Default is cn0.
func (plugin *MasterMacVlanPlugin) WithMasterInterface(master string) *MasterMacVlanPlugin {
	glog.V(100).Infof("Adding master interface %s to MasterMacVlanPlugin", master)

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterMacVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterMacVlanPlugin")
	}

	if master == "" {
		glog.V(100).Infof("error to add master interface, the name of interface can not be empty")

		plugin.errorMsg = "invalid master parameter"

		return plugin
	}

	plugin.masterPlugin.Master = master

	return plugin
}

// WithIPAM defines IPAM configuration to MasterMacVlanPlugin. Default is empty.
func (plugin *MasterMacVlanPlugin) WithIPAM(ipam *IPAM) *MasterMacVlanPlugin {
	glog.V(100).Infof("Adding ipam configuration %v to MasterMacVlanPlugin", ipam)

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterMacVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterMacVlanPlugin")
	}

	if ipam == nil {
		glog.V(100).Infof("error to add empty ipam to MasterMacVlanPlugin")

		plugin.errorMsg = invalidIpamParameterMsg

		return plugin
	}

	plugin.masterPlugin.Ipam = ipam

	return plugin
}

// WithLinkInContainer defines MasterMacVlan plugin using linkInContainer feature.
func (plugin *MasterMacVlanPlugin) WithLinkInContainer() *MasterMacVlanPlugin {
	glog.V(100).Infof("Adding linkInContainer configuration to MasterMacVlanPlugin")

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterMacVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterMacVlanPlugin")
	}

	plugin.masterPlugin.LinkInContainer = true

	return plugin
}

// GetMasterPluginConfig returns master plugin if error is not occur.
func (plugin *MasterMacVlanPlugin) GetMasterPluginConfig() (*MasterPlugin, error) {
	if plugin.errorMsg != "" {
		return nil, fmt.Errorf("error to build MaterPlugin config due to :%s", plugin.errorMsg)
	}

	return plugin.masterPlugin, nil
}

// MasterBridgePlugin provides struct for MasterPlugin set to bridge in NetworkAttachmentDefinition.
type MasterBridgePlugin struct {
	masterPlugin *MasterPlugin
	errorMsg     string
}

// NewMasterBridgePlugin creates new instance of MasterBridgePlugin.
func NewMasterBridgePlugin(name, bridgeName string) *MasterBridgePlugin {
	glog.V(100).Infof(
		"Initializing new MasterBridgePlugin structure %s, with bridge %s", name, bridgeName)

	builder := &MasterBridgePlugin{
		masterPlugin: &MasterPlugin{
			CniVersion: "0.3.1",
			Name:       name,
			Type:       "bridge",
			Bridge:     bridgeName,
		},
	}

	if builder.masterPlugin.Name == "" {
		glog.V(100).Infof("error MasterBridgePlugin can not be empty")

		builder.errorMsg = "MasterBridgePlugin name is empty"

		return builder
	}

	return builder
}

// GetMasterPluginConfig returns master plugin if error does not occur.
func (plugin *MasterBridgePlugin) GetMasterPluginConfig() (*MasterPlugin, error) {
	if plugin.errorMsg != "" {
		return nil, fmt.Errorf("error to build MaterPlugin config due to :%s", plugin.errorMsg)
	}

	return plugin.masterPlugin, nil
}

// WithIPAM defines IPAM configuration to MasterBridgePlugin. Default is empty.
func (plugin *MasterBridgePlugin) WithIPAM(ipam *IPAM) *MasterBridgePlugin {
	glog.V(100).Infof("Adding ipam configuration %v to MasterBridgePlugin", ipam)

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterBridgePlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterBridgePlugin")

		return plugin
	}

	if ipam == nil {
		glog.V(100).Infof("error adding empty ipam to MasterBridgePlugin")

		plugin.errorMsg = invalidIpamParameterMsg

		return plugin
	}

	plugin.masterPlugin.Ipam = ipam

	return plugin
}

// MasterVlanPlugin provides struct for MasterPlugin set to vlan in NetworkAttachmentDefinition.
type MasterVlanPlugin struct {
	masterPlugin *MasterPlugin
	errorMsg     string
}

// NewMasterVlanPlugin creates new instance of MasterVlanPlugin.
func NewMasterVlanPlugin(name string, vlanID uint16) *MasterVlanPlugin {
	glog.V(100).Infof(
		"Initializing new MasterVlanPlugin structure %s, with vlanId %s", name, vlanID)

	builder := &MasterVlanPlugin{
		masterPlugin: &MasterPlugin{
			CniVersion: "0.3.1",
			Name:       name,
			Type:       "vlan",
			VlanID:     vlanID,
		},
	}

	if vlanID > 4094 {
		glog.V(100).Infof("error vlan id can not be greater than 4094")

		builder.errorMsg = "MasterVlanPlugin vlanID is greater than 4094"

		return builder
	}

	if builder.masterPlugin.Name == "" {
		glog.V(100).Infof("error MasterVlanPlugin name can not be empty")

		builder.errorMsg = "MasterVlanPlugin name is empty"

		return builder
	}

	return builder
}

// WithIPAM defines IPAM configuration to MasterVlanPlugin. Default is empty.
func (plugin *MasterVlanPlugin) WithIPAM(ipam *IPAM) *MasterVlanPlugin {
	glog.V(100).Infof("Adding IPAM configuration to MasterVlanPlugin: %v", ipam)

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterVlanPlugin")

		return plugin
	}

	if ipam == nil {
		glog.V(100).Infof("error adding empty ipam to MasterVlanPlugin")

		plugin.errorMsg = invalidIpamParameterMsg

		return plugin
	}

	plugin.masterPlugin.Ipam = ipam

	return plugin
}

// WithMasterInterface defines master interface to MasterVlanPlugin. Default is cn0.
func (plugin *MasterVlanPlugin) WithMasterInterface(masterInterfaceName string) *MasterVlanPlugin {
	glog.V(100).Infof("Adding masterInterfaceName interface %s to MasterVlanPlugin", masterInterfaceName)

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterVlanPlugin")

		return plugin
	}

	if masterInterfaceName == "" {
		glog.V(100).Infof("error to add masterInterfaceName interface, the name of interface can not be empty")

		plugin.errorMsg = "invalid masterInterfaceName parameter"

		return plugin
	}

	plugin.masterPlugin.Master = masterInterfaceName

	return plugin
}

// WithLinkInContainer defines MasterVlanPlugin using linkInContainer feature.
func (plugin *MasterVlanPlugin) WithLinkInContainer() *MasterVlanPlugin {
	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterVlanPlugin")

		return plugin
	}

	plugin.masterPlugin.LinkInContainer = true

	return plugin
}

// GetMasterPluginConfig returns master plugin if error does not occur.
func (plugin *MasterVlanPlugin) GetMasterPluginConfig() (*MasterPlugin, error) {
	if plugin.errorMsg != "" {
		return nil, fmt.Errorf("error to build MaterPlugin config due to :%s", plugin.errorMsg)
	}

	return plugin.masterPlugin, nil
}

// MasterIPVlanPlugin provides struct for MasterPlugin set to IP vlan in NetworkAttachmentDefinition.
type MasterIPVlanPlugin struct {
	masterPlugin *MasterPlugin
	errorMsg     string
}

// NewMasterIPVlanPlugin creates new instance of MasterIP VlanPlugin.
func NewMasterIPVlanPlugin(name string) *MasterIPVlanPlugin {
	glog.V(100).Infof(
		"Initializing new MasterIPVlanPlugin structure %s", name)

	builder := &MasterIPVlanPlugin{
		masterPlugin: &MasterPlugin{
			CniVersion: "0.3.1",
			Name:       name,
			Type:       "ipvlan",
		},
	}

	if builder.masterPlugin.Name == "" {
		glog.V(100).Infof("error MasterIPVlanPlugin can not be empty")

		builder.errorMsg = "MasterIPVlanPlugin name is empty"

		return builder
	}

	return builder
}

// WithIPAM defines IPAM configuration to MasterIPVlanPlugin. Default is empty.
func (plugin *MasterIPVlanPlugin) WithIPAM(ipam *IPAM) *MasterIPVlanPlugin {
	glog.V(100).Infof("Adding IPAM configuration %v to MasterIPVlanPlugin", ipam)

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterIPVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterIPVlanPlugin")

		return plugin
	}

	if ipam == nil {
		glog.V(100).Infof("error adding empty ipam to MasterIPVlanPlugin")

		plugin.errorMsg = invalidIpamParameterMsg

		return plugin
	}

	plugin.masterPlugin.Ipam = ipam

	return plugin
}

// WithMasterInterface defines master interface to MasterVlanPlugin. Default is cn0.
func (plugin *MasterIPVlanPlugin) WithMasterInterface(masterInterfaceName string) *MasterIPVlanPlugin {
	glog.V(100).Infof("Adding masterInterfaceName %s to MasterIPVlanPlugin", masterInterfaceName)

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterIPVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterIPVlanPlugin")

		return plugin
	}

	if masterInterfaceName == "" {
		glog.V(100).Infof("error to add master interface, the name of interface can not be empty")

		plugin.errorMsg = "invalid masterInterfaceName parameter"

		return plugin
	}

	plugin.masterPlugin.Master = masterInterfaceName

	return plugin
}

// WithLinkInContainer defines MasterIPVlanPlugin using linkInContainer feature.
func (plugin *MasterIPVlanPlugin) WithLinkInContainer() *MasterIPVlanPlugin {
	glog.V(100).Infof("Adding linkInContainer feature to MasterIPVlanPlugin")

	if plugin.masterPlugin == nil {
		glog.V(100).Infof(msg.UndefinedCrdObjectErrString("MasterIPVlanPlugin"))
		plugin.errorMsg = msg.UndefinedCrdObjectErrString("MasterIPVlanPlugin")

		return plugin
	}

	plugin.masterPlugin.LinkInContainer = true

	return plugin
}

// GetMasterPluginConfig returns master plugin if error does not occur.
func (plugin *MasterIPVlanPlugin) GetMasterPluginConfig() (*MasterPlugin, error) {
	if plugin.errorMsg != "" {
		return nil, fmt.Errorf("error to build MaterPlugin config due to :%s", plugin.errorMsg)
	}

	return plugin.masterPlugin, nil
}

// MasterBondPlugin provides struct for MasterPlugin to create a NAD for bonded interfaces.
type MasterBondPlugin struct {
	masterPlugin *MasterPlugin
	errorMsg     string
}

// NewMasterBondPlugin creates new instance of MasterBondPlugin.
func NewMasterBondPlugin(name, mode string) *MasterBondPlugin {
	glog.V(100).Infof("Initializing new NewMasterBondPlugin structure %s and %s", name, mode)

	validModes := map[string]bool{
		"balance-rr":    true,
		"active-backup": true,
		"balance-xor":   true,
	}

	builder := &MasterBondPlugin{
		masterPlugin: &MasterPlugin{
			CniVersion: "0.3.1",
			Name:       name,
			Type:       "bond",
			Mode:       mode,
		},
	}

	// Check if the provided mode is valid
	if !validModes[mode] {
		glog.V(100).Infof("error: invalid mode type %s used for MasterBondPlugin bond interface", mode)

		builder.errorMsg = "Bond mode type is not valid"

		return builder
	}

	if builder.masterPlugin.Name == "" {
		glog.V(100).Infof("error NewMasterBondPlugin name can not be empty")

		builder.errorMsg = "NewMasterBondPlugin name is empty"

		return builder
	}

	return builder
}

// WithLinksInContainer defines linksInContainer configuration to MasterBondPlugin.
func (plugin *MasterBondPlugin) WithLinksInContainer(linksInContainer bool) *MasterBondPlugin {
	glog.V(100).Infof("Adding linksInContainer %v to MasterBoundPlugin", linksInContainer)
	plugin.masterPlugin.LinksInContainer = linksInContainer

	return plugin
}

// WithFailOverMac defines FailOverMac configuration to MasterBondPlugin.
func (plugin *MasterBondPlugin) WithFailOverMac(failOverMac int) *MasterBondPlugin {
	glog.V(100).Infof("Adding failOverMac %d to MasterBoundPlugin", failOverMac)

	if failOverMac < 0 || failOverMac > 2 {
		glog.V(100).Infof("error adding incorrect value %d for FailOverMac in MasterBondPlugin", failOverMac)

		plugin.errorMsg = "error adding incorrect FailOverMac to MasterBondPlugin"

		return plugin
	}

	plugin.masterPlugin.FailOverMac = failOverMac

	return plugin
}

// WithMiimon defines Miimon configuration to MasterBondPlugin.
func (plugin *MasterBondPlugin) WithMiimon(miimon int) *MasterBondPlugin {
	glog.V(100).Infof("Adding miimon %d to MasterBoundPlugin", miimon)

	if miimon < 0 {
		glog.V(100).Infof("error adding incorrect miimon value %d to MasterBondPlugin", miimon)

		plugin.errorMsg = "error adding incorrect miimon value to MasterBondPlugin"

		return plugin
	}

	plugin.masterPlugin.Miimon = fmt.Sprintf("%d", miimon)

	return plugin
}

// WithLinks defines Links configuration to MasterBondPlugin.
func (plugin *MasterBondPlugin) WithLinks(links []Link) *MasterBondPlugin {
	glog.V(100).Infof("Adding links %v to MasterBoundPlugin", links)

	if links == nil {
		glog.V(100).Infof("links value %v cannot be nil or empty in MasterBondPlugin", links)

		plugin.errorMsg = "error adding empty links to MasterBondPlugin"

		return plugin
	}

	plugin.masterPlugin.Links = links

	return plugin
}

// WithCapabilities defines Capabilities configuration to MasterBondPlugin.
func (plugin *MasterBondPlugin) WithCapabilities(capabilities *Capability) *MasterBondPlugin {
	glog.V(100).Infof("Adding capabilities value %d to MasterBoundPlugin", capabilities)

	if capabilities == nil {
		glog.V(100).Infof("error adding nil value %d capabilities to MasterBondPlugin", capabilities)

		plugin.errorMsg = "error adding empty capabilities to MasterBondPlugin"

		return plugin
	}

	plugin.masterPlugin.Capabilities = capabilities

	return plugin
}

// WithIPAM defines IPAM configuration to MasterBondPlugin.
func (plugin *MasterBondPlugin) WithIPAM(ipam *IPAM) *MasterBondPlugin {
	glog.V(100).Infof("Adding ipam %v to MasterBoundPlugin", ipam)

	if ipam == nil {
		glog.V(100).Infof("error adding ipam value %d to MasterBondPlugin", ipam)

		plugin.errorMsg = "error adding empty ipam to MasterBondPlugin"

		return plugin
	}

	plugin.masterPlugin.Ipam = ipam

	return plugin
}

// WithVLANInContainer defines a VLAN ID configuration to MasterBondPlugin.
func (plugin *MasterBondPlugin) WithVLANInContainer(vlan uint16) *MasterBondPlugin {
	glog.V(100).Infof("Adding vlan %d to MasterBoundPlugin", vlan)

	if vlan > 4095 {
		glog.V(100).Infof("error adding incorrect vlan id %d to MasterBondPlugin", vlan)

		plugin.errorMsg = "error adding incorrect vlan to MasterBondPlugin"

		return plugin
	}

	plugin.masterPlugin.VlanID = vlan

	return plugin
}

// GetMasterPluginConfig returns master plugin if error is not occur.
func (plugin *MasterBondPlugin) GetMasterPluginConfig() (*MasterPlugin, error) {
	if plugin.errorMsg != "" {
		return nil, fmt.Errorf("error to build MasterPlugin config due to :%s", plugin.errorMsg)
	}

	return plugin.masterPlugin, nil
}
