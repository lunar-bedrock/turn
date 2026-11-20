// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package client

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermission(t *testing.T) {
	t.Run("Getter and setter", func(t *testing.T) {
		perm := &permission{}

		assert.Equal(t, permStateIdle, perm.state())
		perm.setState(permStatePermitted)
		assert.Equal(t, permStatePermitted, perm.state())
	})
}

func TestPermissionMap(t *testing.T) {
	t.Run("Basic operations", func(t *testing.T) {
		pm := newPermissionMap()
		assert.NotNil(t, pm)
		assert.NotNil(t, pm.permMap)

		perm1 := &permission{st: permStateIdle}
		perm2 := &permission{st: permStatePermitted}
		perm3 := &permission{st: permStateIdle}
		udpAddr1, _ := net.ResolveUDPAddr("udp", "1.2.3.4:5000")
		udpAddr2, _ := net.ResolveUDPAddr("udp", "5.6.7.8:8888")
		tcpAddr, _ := net.ResolveTCPAddr("tcp", "7.8.9.10:5000")
		assert.True(t, pm.insert(udpAddr1, perm1))
		assert.Equal(t, 1, len(pm.permMap))
		assert.True(t, pm.insert(udpAddr2, perm2))
		assert.Equal(t, 2, len(pm.permMap))
		assert.True(t, pm.insert(tcpAddr, perm3))
		assert.Equal(t, 3, len(pm.permMap))

		perms, ok := pm.find(udpAddr1)
		assert.True(t, ok)
		assert.Equal(t, perm1, perms)
		assert.Equal(t, permStateIdle, perms.st)

		perms, ok = pm.find(udpAddr2)
		assert.True(t, ok)
		assert.Equal(t, perm2, perms)
		assert.Equal(t, permStatePermitted, perms.st)

		perms, ok = pm.find(tcpAddr)
		assert.True(t, ok)
		assert.Equal(t, perm3, perms)
		assert.Equal(t, permStateIdle, perms.st)

		addrs := pm.permittedAddrs()
		assert.Equal(t, 1, len(addrs), "only the granted permission is listed")
		assert.Equal(t, udpAddr2.String(), addrs[0].String())

		pm.delete(tcpAddr)
		assert.Equal(t, 2, len(pm.permMap))
		pm.delete(udpAddr1)
		assert.Equal(t, 1, len(pm.permMap))
		pm.delete(udpAddr2)
		assert.Equal(t, 0, len(pm.permMap))
	})
}
