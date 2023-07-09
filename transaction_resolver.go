package msmptd

import "net"

// Resolver returns net.Resolver being used for this transaction
func (t *Transaction) Resolver() *net.Resolver {
	if t.server != nil {
		if t.server.Resolver != nil {
			return t.server.Resolver
		}
	}
	return net.DefaultResolver
}
