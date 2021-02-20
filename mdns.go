// Package mdns is an mdns router
package mdns

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	mdns "github.com/unistack-org/micro-register-mdns/v3/util"
	"github.com/unistack-org/micro/v3/router"
)

// NewRouter returns an initialized dns router
func NewRouter(opts ...router.Option) router.Router {
	options := router.NewOptions(opts...)
	for _, o := range opts {
		o(&options)
	}
	return &mdnsRouter{options}
}

type mdnsRouter struct {
	options router.Options
}

func (m *mdnsRouter) Init(opts ...router.Option) error {
	for _, o := range opts {
		o(&m.options)
	}
	return nil
}

func (m *mdnsRouter) Options() router.Options {
	return m.options
}

func (m *mdnsRouter) Table() router.Table {
	return nil
}

func (m *mdnsRouter) Lookup(opts ...router.QueryOption) ([]router.Route, error) {
	options := router.NewQuery(opts...)

	// check to see if we have the port provided in the service, e.g. go-micro-srv-foo:8000
	service, port, err := net.SplitHostPort(options.Service)
	if err != nil {
		service = options.Service
	}

	// query for the host
	entries := make(chan *mdns.ServiceEntry)

	p := mdns.DefaultParams(service)
	//p.Timeout = time.Millisecond * 100
	p.Entries = entries
	ctx, cancel := context.WithTimeout(m.options.Context, time.Millisecond*100)
	defer cancel()

	// check if we're using our own network
	if len(options.Network) > 0 {
		p.Domain = options.Network
	}

	// do the query
	if err := mdns.Query(ctx, p); err != nil {
		return nil, err
	}

	var routes []router.Route

	// compose the routes based on the entries
	for e := range entries {
		addr := e.Host
		// prefer ipv4 addrs
		if len(e.AddrV4) > 0 {
			addr = e.AddrV4.String()
			// else use ipv6
		} else if len(e.AddrV6) > 0 {
			addr = "[" + e.AddrV6.String() + "]"
		} else if len(addr) == 0 {
			continue
		}

		pt := 443

		if e.Port > 0 {
			pt = e.Port
		}

		// set the port
		if len(port) > 0 {
			pt, _ = strconv.Atoi(port)
		}

		routes = append(routes, router.Route{
			Service: service,
			Address: fmt.Sprintf("%s:%d", addr, pt),
			Network: p.Domain,
		})
	}

	return routes, nil
}

func (m *mdnsRouter) Watch(opts ...router.WatchOption) (router.Watcher, error) {
	return nil, nil
}

func (m *mdnsRouter) Close() error {
	return nil
}

func (m *mdnsRouter) Name() string {
	return m.options.Name
}

func (m *mdnsRouter) String() string {
	return "mdns"
}
