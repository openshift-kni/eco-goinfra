<<<<<<< HEAD
//go:build darwin || dragonfly || freebsd || netbsd || openbsd
=======
>>>>>>> f03ab420 (bump vendors)
// +build darwin dragonfly freebsd netbsd openbsd

package sockaddr

import "os/exec"

<<<<<<< HEAD
var cmds = map[string][]string{
	"route": {"/sbin/route", "-n", "get", "default"},
}

=======
var cmds map[string][]string = map[string][]string{
	"route": {"/sbin/route", "-n", "get", "default"},
}

type routeInfo struct {
	cmds map[string][]string
}

>>>>>>> f03ab420 (bump vendors)
// NewRouteInfo returns a BSD-specific implementation of the RouteInfo
// interface.
func NewRouteInfo() (routeInfo, error) {
	return routeInfo{
		cmds: cmds,
	}, nil
}

// GetDefaultInterfaceName returns the interface name attached to the default
// route on the default interface.
func (ri routeInfo) GetDefaultInterfaceName() (string, error) {
	out, err := exec.Command(cmds["route"][0], cmds["route"][1:]...).Output()
	if err != nil {
		return "", err
	}

	var ifName string
	if ifName, err = parseDefaultIfNameFromRoute(string(out)); err != nil {
		return "", err
	}
	return ifName, nil
}
