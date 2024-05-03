<<<<<<< HEAD
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

=======
>>>>>>> f03ab420 (bump vendors)
package api

// Sys is used to perform system-related operations on Vault.
type Sys struct {
	c *Client
}

// Sys is used to return the client for sys-related API calls.
func (c *Client) Sys() *Sys {
	return &Sys{c: c}
}
