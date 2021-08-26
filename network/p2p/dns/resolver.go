package dns

import (
	"context"
	"net"
	"sync"
	"time"
	_ "unsafe" // for linking runtimeNano

	madns "github.com/multiformats/go-multiaddr-dns"

	"github.com/onflow/flow-go/engine"
	"github.com/onflow/flow-go/module"
)

//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

// Resolver is a cache-based dns resolver for libp2p.
// - DNS cache implementation notes:
//   - Generic / possibly expected functionality NOT implemented:
//     - Caches domains for TTL seconds as given by upstream DNS resolver, e.g. [1].
//     - Possibly pre-expire cached domains so no connection time resolve delay.
//   - Actual / pragmatic functionality implemented below:
//     - Caches domains for global (not individual domain record TTL) TTL seconds.
//     - Cached IP is returned even if cached entry expired; so no connection time resolve delay.
//     - Detecting expired cached domain triggers async DNS lookup to refresh cached entry.
// [1] https://en.wikipedia.org/wiki/Name_server#Caching_name_server
type Resolver struct {
	sync.Mutex
	c              *cache
	res            madns.BasicResolver // underlying resolver
	collector      module.ResolverMetrics
	unit           *engine.Unit
	processingIPs  map[string]struct{}
	processingTXTs map[string]struct{}
}

// optFunc is the option function for Resolver.
type optFunc func(resolver *Resolver)

// WithBasicResolver is an option function for setting the basic resolver of this Resolver.
func WithBasicResolver(basic madns.BasicResolver) optFunc {
	return func(resolver *Resolver) {
		resolver.res = basic
	}
}

// WithTTL is an option function for setting the time to live for cache entries.
func WithTTL(ttl time.Duration) optFunc {
	return func(resolver *Resolver) {
		resolver.c.ttl = ttl
	}
}

// NewResolver is the factory function for creating an instance of this resolver.
func NewResolver(collector module.ResolverMetrics, opts ...optFunc) (*madns.Resolver, error) {
	resolver := &Resolver{
		res:            madns.DefaultResolver,
		c:              newCache(),
		collector:      collector,
		processingIPs:  map[string]struct{}{},
		processingTXTs: map[string]struct{}{},
		unit:           engine.NewUnit(),
	}

	for _, opt := range opts {
		opt(resolver)
	}

	return madns.NewResolver(madns.WithDefaultResolver(resolver))
}

// Ready initializes the resolver and returns a channel that is closed when the initialization is done.
func (r *Resolver) Ready() <-chan struct{} {
	return r.unit.Ready()
}

// Done terminates the resolver and returns a channel that is closed when the termination is done
func (r *Resolver) Done() <-chan struct{} {
	return r.unit.Done()
}

// LookupIPAddr implements BasicResolver interface for libp2p for looking up ip addresses through resolver.
func (r *Resolver) LookupIPAddr(ctx context.Context, domain string) ([]net.IPAddr, error) {
	started := runtimeNano()

	addr, err := r.lookupIPAddr(ctx, domain)

	r.collector.DNSLookupDuration(
		time.Duration(runtimeNano() - started))
	return addr, err
}

// lookupIPAddr encapsulates the logic of resolving an ip address through cache.
func (r *Resolver) lookupIPAddr(ctx context.Context, domain string) ([]net.IPAddr, error) {
	addr, exists, expired := r.c.resolveIPCache(domain)

	if !exists {
		r.collector.OnDNSCacheMiss()
		return r.lookupResolverForIPAddr(ctx, domain)
	}

	if expired && r.shouldResolveIP(domain) {
		r.unit.Launch(func() {
			_, err := r.lookupResolverForIPAddr(ctx, domain)
			if err != nil {
				r.collector.OnDNSCacheInvalidated()
			}
			r.doneResolvingIP(domain)
		})
	}

	r.collector.OnDNSCacheHit()
	return addr, nil
}

// lookupResolverForIPAddr queries the underlying resolver for the domain and updates the cache if query is successful.
func (r *Resolver) lookupResolverForIPAddr(ctx context.Context, domain string) ([]net.IPAddr, error) {
	addr, err := r.res.LookupIPAddr(ctx, domain)
	if err != nil {
		return nil, err
	}

	r.c.updateIPCache(domain, addr) // updates cache

	return addr, nil
}

// LookupTXT implements BasicResolver interface for libp2p.
func (r *Resolver) LookupTXT(ctx context.Context, txt string) ([]string, error) {

	started := runtimeNano()

	addr, err := r.lookupTXT(ctx, txt)

	r.collector.DNSLookupDuration(
		time.Duration(runtimeNano() - started))
	return addr, err
}

// lookupIPAddr encapsulates the logic of resolving a txt through cache.
func (r *Resolver) lookupTXT(ctx context.Context, txt string) ([]string, error) {
	addr, exists, expired := r.c.resolveTXTCache(txt)

	if !exists {
		r.collector.OnDNSCacheMiss()
		return r.lookupResolverForTXTAddr(ctx, txt)
	}

	if expired && r.shouldResolveTXT(txt) {
		r.unit.Launch(func() {
			_, err := r.lookupResolverForTXTAddr(ctx, txt)
			if err != nil {
				r.collector.OnDNSCacheInvalidated()
			}
			r.doneResolvingTXT(txt)
		})

	}

	r.collector.OnDNSCacheHit()
	return addr, nil
}

// lookupResolverForIPAddr queries the underlying resolver for the domain and updates the cache if query is successful.
func (r *Resolver) lookupResolverForTXTAddr(ctx context.Context, txt string) ([]string, error) {
	addr, err := r.res.LookupTXT(ctx, txt)
	if err != nil {
		return nil, err
	}

	r.c.updateTXTCache(txt, addr) // updates cache

	return addr, nil
}

func (r *Resolver) shouldResolveIP(domain string) bool {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.processingIPs[domain]; !ok {
		r.processingIPs[domain] = struct{}{}
		return true
	}

	return false
}

func (r *Resolver) doneResolvingIP(domain string) {
	r.Lock()
	defer r.Unlock()

	delete(r.processingIPs, domain)
}

func (r *Resolver) doneResolvingTXT(txt string) {
	r.Lock()
	defer r.Unlock()

	delete(r.processingTXTs, txt)
}

func (r *Resolver) shouldResolveTXT(txt string) bool {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.processingIPs[txt]; !ok {
		r.processingTXTs[txt] = struct{}{}
		return true
	}

	return false
}
