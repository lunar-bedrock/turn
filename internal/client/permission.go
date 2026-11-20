// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package client

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/turn/v5/internal/ipnet"
)

type permState int32

const (
	permStateIdle permState = iota
	permStateRequesting
	permStatePermitted
)

type permission struct {
	addr  net.Addr
	st    permState    // Thread-safe (atomic op)
	mutex sync.RWMutex // Thread-safe

	// err is the last failed request, returned to writes until retryAt. Guarded by mutex.
	err     error
	retryAt time.Time
}

func (p *permission) setState(state permState) {
	atomic.StoreInt32((*int32)(&p.st), int32(state))
}

func (p *permission) state() permState {
	return permState(atomic.LoadInt32((*int32)(&p.st)))
}

// Thread-safe permission map.
type permissionMap struct {
	permMap map[string]*permission
	mutex   sync.RWMutex
}

func (m *permissionMap) insert(addr net.Addr, p *permission) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	p.addr = addr
	m.permMap[ipnet.FingerprintAddr(addr)] = p

	return true
}

func (m *permissionMap) find(addr net.Addr) (*permission, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	p, ok := m.permMap[ipnet.FingerprintAddr(addr)]

	return p, ok
}

func (m *permissionMap) delete(addr net.Addr) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.permMap, ipnet.FingerprintAddr(addr))
}

// deleteExpiredDenials forgets rejected permissions whose retry time has passed, so a long-lived
// allocation does not keep every peer it was ever denied.
func (m *permissionMap) deleteExpiredDenials(now time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for key, p := range m.permMap {
		p.mutex.RLock()
		expired := p.state() == permStateIdle && p.err != nil && !now.Before(p.retryAt)
		p.mutex.RUnlock()
		if expired {
			delete(m.permMap, key)
		}
	}
}

// permittedAddrs returns the addresses whose permission the server has granted.
func (m *permissionMap) permittedAddrs() []net.Addr {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	addrs := []net.Addr{}
	for _, p := range m.permMap {
		if p.state() == permStatePermitted {
			addrs = append(addrs, p.addr)
		}
	}

	return addrs
}

func (m *permissionMap) addrs() []net.Addr {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	addrs := []net.Addr{}
	for _, p := range m.permMap {
		addrs = append(addrs, p.addr)
	}

	return addrs
}

func newPermissionMap() *permissionMap {
	return &permissionMap{
		permMap: map[string]*permission{},
	}
}
