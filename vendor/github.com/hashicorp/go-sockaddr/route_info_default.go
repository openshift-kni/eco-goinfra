<<<<<<< HEAD
//go:build android || nacl || plan9 || js
// +build android nacl plan9 js

package sockaddr

// getDefaultIfName is the default interface function for unsupported platforms.
func getDefaultIfName() (string, error) {
	return "", ErrNoInterface
}

func NewRouteInfo() (routeInfo, error) {
	return routeInfo{}, ErrNoRoute
}

// GetDefaultInterfaceName returns the interface name attached to the default
// route on the default interface.
func (ri routeInfo) GetDefaultInterfaceName() (string, error) {
	return "", ErrNoInterface
=======
// +build android nacl plan9

package sockaddr

import "errors"

// getDefaultIfName is the default interface function for unsupported platforms.
func getDefaultIfName() (string, error) {
	return "", errors.New("No default interface found (unsupported platform)")
>>>>>>> f03ab420 (bump vendors)
}
