// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"net"

	"github.com/pion/stun/v4"
)

type mockClient struct {
	writeTo            func(data []byte, to net.Addr) (int, error)
	performTransaction func(msg *stun.Message, to net.Addr, dontWait bool, context context.Context) (TransactionResult, error)
	onDeallocated      func(relayedAddr net.Addr)
	sendErr            error // returned by StartTransaction, as a failed local write
	onStart            func(msg *stun.Message)
}

func (c *mockClient) WriteTo(data []byte, to net.Addr) (int, error) {
	if c.writeTo != nil {
		return c.writeTo(data, to)
	}

	return 0, nil
}

func (c *mockClient) PerformTransactionWithContext(msg *stun.Message, to net.Addr, dontWait bool, ctx context.Context) (TransactionResult, error) {
	if c.performTransaction != nil {
		return c.performTransaction(msg, to, dontWait, ctx)
	}

	return TransactionResult{}, errFake
}

func (c *mockClient) StartTransaction(msg *stun.Message, to net.Addr) (func(context.Context) (TransactionResult, error), error) {
	if c.sendErr != nil {
		return nil, c.sendErr
	}
	if c.onStart != nil {
		c.onStart(msg)
	}

	return func(ctx context.Context) (TransactionResult, error) {
		return c.PerformTransactionWithContext(msg, to, false, ctx)
	}, nil
}

func (c *mockClient) OnDeallocated(relayedAddr net.Addr) {
	if c.onDeallocated != nil {
		c.onDeallocated(relayedAddr)
	}
}
