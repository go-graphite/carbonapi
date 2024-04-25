package gosnowth

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Logger values implement the behavior used by SnowthClient for logging,
// if the client has been assigned a logger with this interface.
type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

// SnowthNode values represent a snowth node. An IRONdb cluster consists of
// several nodes.  A SnowthNode has a URL to the API of that Node, an identifier,
// and a current topology.  The identifier is how the node is identified within
// the cluster, and the topology is the current topology that the node falls
// within.  A topology is a set of nodes that distribute data amongst each other.
type SnowthNode struct {
	url             *url.URL
	identifier      string
	currentTopology string
	semVer          string
}

// GetURL returns the *url.URL for a given SnowthNode. This is useful if you
// need the raw connection string of a given snowth node, such as when making a
// proxy for a snowth node.
func (sn *SnowthNode) GetURL() *url.URL {
	return sn.url
}

// SemVer returns a string containing the semantic version of IRONdb the node
// is currently running.
func (sn *SnowthNode) SemVer() string {
	return sn.semVer
}

// GetCurrentTopology return the hash string representation of the
// node's current topology.
func (sn *SnowthNode) GetCurrentTopology() string {
	return sn.currentTopology
}

// httpClient values are used to define the behavior needed from HTTP client
// values.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config values represent configuration information SnowthClient values.
type Config struct {
	DialTimeout    time.Duration `json:"dial_timeout,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty"`
	WatchInterval  time.Duration `json:"watch_interval,omitempty"`
	Retries        int64         `json:"retries,omitempty"`
	ConnectRetries int64         `json:"connect_retries,omitempty"`
	Servers        []string      `json:"servers,omitempty"`
	DenyHosts      []string      `json:"deny_hosts,omitempty"`
	CtxKeyTraceID  interface{}   `json:"-"`
}

// NewConfig creates and initializes a new SnowthClient configuration value
// using default values.
func NewConfig(servers ...string) *Config {
	return &Config{
		DialTimeout:    500 * time.Millisecond,
		Servers:        servers,
		Timeout:        10 * time.Second,
		WatchInterval:  30 * time.Second,
		Retries:        0,
		ConnectRetries: -1,
	}
}

// SnowthClient values provide client functionality for accessing IRONdb.
type SnowthClient struct {
	sync.RWMutex
	c httpClient

	// timeout is the maximum duration that a snowth request is allowed to run.
	timeout time.Duration

	// retries is used to determine weather or not to retry requests which
	// fail due to timeouts or other non-connection problems.
	retries int64

	// connRetries is used to determine weather or not to retry requests which
	// fail to snowth nodes due to connection problems.
	connRetries int64

	// in order to keep track of healthy nodes within the cluster,
	// we have two lists of SnowthNode types, active and inactive.
	activeNodes   []*SnowthNode
	inactiveNodes []*SnowthNode

	// denyHosts are a list of hosts to always keep inactive.
	denyHosts []string

	// watchInterval is the duration between checks to tell if a node is active
	// or inactive.
	watchInterval time.Duration

	// If log output is desired, a value matching the Logger interface can be
	// assigned.  If this is nil, no log output will be attempted.
	log Logger

	// request is an assignable middleware function which modifies the request
	// before it is used by SnowthClient to connect with IRONdb. Tracing headers
	// or other context information can be added by this function.
	request func(r *http.Request) error

	// dumpRequests and traceRequests are settings from the environment
	// GOSNOWTH_DUMP_REQUESTS and GOSNOWTH_TRACE_REQUESTS respectively.
	// Set to a path `/data/fetch` or `*` for all paths.
	// Dump: full request w/payload is emitted to stdout
	// Trace: httptrace of request
	dumpRequests  string
	traceRequests string

	// ctxKeyTraceID is a context key to use to retrieve trace ID's from
	// contexts passed to gosnowth functions.
	ctxKeyTraceID interface{}

	// current topology
	currentTopology         string
	currentTopologyCompiled *Topology
}

// NewClient creates and performs initial setup of a new SnowthClient.
func NewClient(ctx context.Context, cfg *Config,
	logs ...Logger,
) (*SnowthClient, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = time.Second * 10
	}

	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = time.Millisecond * 500
	}

	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   cfg.DialTimeout,
				KeepAlive: cfg.Timeout,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			DisableKeepAlives:     true,
			MaxConnsPerHost:       0,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   1,
			IdleConnTimeout:       cfg.Timeout,
			TLSHandshakeTimeout:   cfg.DialTimeout,
			ExpectContinueTimeout: cfg.DialTimeout,
		},
	}

	sc := &SnowthClient{
		c:             client,
		activeNodes:   []*SnowthNode{},
		inactiveNodes: []*SnowthNode{},
		watchInterval: cfg.WatchInterval,
		timeout:       cfg.Timeout,
		retries:       cfg.Retries,
		connRetries:   cfg.ConnectRetries,
		dumpRequests:  os.Getenv("GOSNOWTH_DUMP_REQUESTS"),
		traceRequests: os.Getenv("GOSNOWTH_TRACE_REQUESTS"),
		denyHosts:     cfg.DenyHosts,
		ctxKeyTraceID: cfg.CtxKeyTraceID,
	}

	if len(logs) > 0 {
		sc.log = logs[0]
	}

	// Initial setup of the client may continue in the background after
	// the client has been returned to the caller.
	cCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)

	doneCh := make(chan struct{}, 1)

	defer close(doneCh)

	go func(ctx context.Context, sc *SnowthClient, cfg *Config) {
		remaining := len(cfg.Servers)

		nodeCh := make(chan *SnowthNode, remaining)

		defer close(nodeCh)

		errCh := make(chan error, remaining)

		defer close(errCh)

		for _, addr := range cfg.Servers {
			go func(addr string) {
				url, err := url.Parse(addr)
				if err != nil {
					errCh <- fmt.Errorf("unable to parse server url %s: %w",
						addr, err)

					return
				}

				for _, dh := range cfg.DenyHosts {
					if url.Host == dh {
						errCh <- fmt.Errorf("deny host found in servers: %s",
							url.Host)

						return
					}
				}

				node := &SnowthNode{url: url}

				stats, err := sc.GetStatsNodeContext(ctx, node)
				if err != nil {
					// This node had an error, put on inactive list.
					sc.AddNodes(node)

					errCh <- fmt.Errorf("unable to get status of node %s: %w",
						addr, err)

					return
				}

				node.identifier = stats.Identity()
				node.currentTopology = stats.CurrentTopology()
				node.semVer = stats.SemVer()

				select {
				case <-ctx.Done():
					errCh <- fmt.Errorf("timeout getting status of node %s: %w",
						addr, ctx.Err())
				default:
					nodeCh <- node
				}
			}(addr)
		}

		first := true

		for remaining > 0 {
			select {
			case err := <-errCh:
				sc.LogErrorf(err.Error())
			case node := <-nodeCh:
				sc.AddNodes(node)
				sc.ActivateNodes(node)

				if first {
					sc.Lock()

					sc.currentTopology = node.currentTopology

					sc.Unlock()

					doneCh <- struct{}{}

					first = false
				}
			}

			remaining--
		}
	}(cCtx, sc, cfg)

	for {
		select {
		case <-ctx.Done():
			cancel()

			return nil, fmt.Errorf("no snowth nodes could be activated")
		case <-doneCh:
			return sc, nil
		}
	}
}

// Retries gets the number of retries a SnowthClient will attempt when
// errors other than connection errors occur with a snowth node.
// Retires will repeat the request with exponential backoff until this number
// of retries is reached.
func (sc *SnowthClient) Retries() int64 {
	sc.RLock()
	defer sc.RUnlock()

	return sc.retries
}

// SetRetries sets the number of retries a SnowthClient will attempt when
// errors other than connection errors occur with a snowth node.
// Retires will repeat the request with exponential backoff until this number
// of retries is reached.
func (sc *SnowthClient) SetRetries(num int64) {
	sc.Lock()
	defer sc.Unlock()
	sc.retries = num
}

// ConnectRetries gets the number of retries a SnowthClient will attempt when
// connection errors occur to a snowth node. When a connection error occurs
// the affected node will be deactivated, then a retries will happen on
// another node. A value of -1 will retry until no nodes are available,
// The watch routine can reactivate nodes that have been deactivated by
// retries when their connectivity is restored.
func (sc *SnowthClient) ConnectRetries() int64 {
	sc.RLock()
	defer sc.RUnlock()

	return sc.connRetries
}

// SetConnectRetries sets the number of retries a SnowthClient will attempt when
// connection errors occur to a snowth node. When a connection error occurs
// the affected node will be deactivated, then a retries will happen on
// another node. A value of -1 will retry until no nodes are available,
// The watch routine can reactivate nodes that have been deactivated by
// retries when their connectivity is restored.
func (sc *SnowthClient) SetConnectRetries(num int64) {
	sc.Lock()
	defer sc.Unlock()
	sc.connRetries = num
}

// SetRequestFunc sets an optional middleware function that is used to modify
// the HTTP request before it is used by SnowthClient to connect with IRONdb.
// Tracing headers or other context information provided by the user of this
// library can be added by this function.
func (sc *SnowthClient) SetRequestFunc(f func(r *http.Request) error) {
	sc.Lock()
	defer sc.Unlock()
	sc.request = f
}

// SetWatchInterval sets the interval at which the watch process executes.
func (sc *SnowthClient) SetWatchInterval(d time.Duration) {
	sc.Lock()
	defer sc.Unlock()
	sc.watchInterval = d
}

// SetLog assigns a logger to the snowth client.
func (sc *SnowthClient) SetLog(log Logger) {
	sc.Lock()
	defer sc.Unlock()
	sc.log = log
}

// LogInfof writes a log entry at the information level.
func (sc *SnowthClient) LogInfof(format string, args ...interface{}) {
	if sc.log != nil {
		sc.log.Infof(format, args...)
	}
}

// LogWarnf writes a log entry at the warning level.
func (sc *SnowthClient) LogWarnf(format string, args ...interface{}) {
	if sc.log != nil {
		sc.log.Warnf(format, args...)
	}
}

// LogErrorf writes a log entry at the error level.
func (sc *SnowthClient) LogErrorf(format string, args ...interface{}) {
	if sc.log != nil {
		sc.log.Errorf(format, args...)
	}
}

// LogDebugf writes a log entry at the debug level.
func (sc *SnowthClient) LogDebugf(format string, args ...interface{}) {
	if sc.log != nil {
		sc.log.Debugf(format, args...)
	}
}

// Topology returns the currently active topology.
func (sc *SnowthClient) Topology() (*Topology, error) {
	if sc.currentTopologyCompiled != nil {
		return sc.currentTopologyCompiled, nil
	}

	var lasterr error = nil

	for _, node := range sc.ListActiveNodes() {
		if topology, lasterr := sc.GetTopologyInfo(node); lasterr == nil {
			sc.currentTopologyCompiled = topology

			return topology, nil
		}
	}

	return nil, lasterr
}

// FindMetricNodeIDs returns (possibly) as list of uuid node identifiers that
// own the metric.
func (sc *SnowthClient) FindMetricNodeIDs(uuid, metric string) []string {
	topo, err := sc.Topology()
	if topo == nil || err != nil {
		return make([]string, 0)
	}

	results, err := topo.FindMetricNodeIDs(uuid, metric)
	if results == nil || err != nil {
		return make([]string, 0)
	}

	return results
}

// isNodeActive checks to see if a given node is active or not taking into
// account the ability to get the node state, gossip information and the gossip
// age of the node. If the age is larger than 10 the node is considered
// inactive.
func (sc *SnowthClient) isNodeActive(ctx context.Context,
	node *SnowthNode,
) bool {
	sc.RLock()
	dhosts := sc.denyHosts
	sc.RUnlock()

	for _, dh := range dhosts {
		if node.GetURL().Host == dh {
			sc.LogWarnf("deny host from active node check: %s",
				node.GetURL().Host)

			return false
		}
	}

	if node.identifier == "" || node.semVer == "" {
		// go get state to figure out identity
		stats, err := sc.GetStatsNodeContext(ctx, node)
		if err != nil {
			// error means we failed, node is not active
			sc.LogWarnf("unable to get the state of the node: %s", err.Error())

			return false
		}

		node.identifier = stats.Identity()
		node.semVer = stats.SemVer()
		sc.LogDebugf("retrieved state of node: %s -> %s",
			node.GetURL().Host, node.identifier)
	}

	gossip, err := sc.GetGossipInfo(node)
	if err != nil {
		sc.LogWarnf("unable to get the gossip info of the node: %s",
			err.Error())

		return false
	}

	age := float64(100)

	for _, entry := range []GossipDetail(*gossip) {
		if entry.ID == node.identifier {
			age = entry.Age

			break
		}
	}

	if age > 10.0 {
		sc.LogWarnf("gossip age expired: %s -> %d", node.GetURL().Host, age)

		return false
	}

	return true
}

// WatchAndUpdate watches gossip data for all nodes, and move the nodes to
// the active or inactive pools as required.  It returns a function to cancel
// the operation if needed. It accepts a context value as an argument which
// will also cancel the operation if the context is cancelled or expired. If
// context cancellation is not needed, nil can be passed as the argument.
func (sc *SnowthClient) WatchAndUpdate(ctx context.Context) {
	sc.RLock()
	wi := sc.watchInterval
	to := sc.timeout
	sc.RUnlock()

	if wi <= time.Duration(0) {
		return
	}

	go func(wi time.Duration) {
		tick := time.NewTimer(wi)

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				sc.LogDebugf("updating active snowth nodes")

				for _, node := range sc.ListActiveNodes() {
					tcCtx, tcCancel := context.WithTimeout(ctx, to)

					topology, err := sc.GetTopologyInfoContext(tcCtx, node)

					tcCancel()

					if err != nil {
						sc.LogErrorf("error getting topology info "+
							"from node %s: %w", node.GetURL().Host, err)
					} else if topology == nil || len(topology.Nodes) == 0 {
						sc.LogErrorf("received no topology info "+
							"from node %s", node.GetURL().Host)
					} else {
						for _, topoNode := range topology.Nodes {
							pcCtx, pcCancel := context.WithTimeout(ctx, to)

							sc.populateNodeInfo(pcCtx,
								node.GetCurrentTopology(), topoNode)

							pcCancel()
						}

						break
					}
				}

				for _, node := range sc.ListInactiveNodes() {
					icCtx, icCancel := context.WithTimeout(ctx, to)

					if sc.isNodeActive(icCtx, node) {
						sc.LogDebugf("moving snowth node to active list: %s",
							node.GetURL().Host)

						sc.ActivateNodes(node)
					}

					icCancel()
				}

				for _, node := range sc.ListActiveNodes() {
					acCtx, acCancel := context.WithTimeout(ctx, to)

					if !sc.isNodeActive(acCtx, node) {
						sc.LogWarnf("moving snowth node to inactive list: %s",
							node.GetURL().Host)

						sc.DeactivateNodes(node)
					}

					acCancel()
				}

				tick = time.NewTimer(wi)
			}
		}
	}(wi)
}

// populateNodeInfo populates an existing node with details from the topology.
// If a node doesn't exist, it will be added to the list of active nodes.
func (sc *SnowthClient) populateNodeInfo(ctx context.Context, hash string,
	topology TopologyNode,
) {
	sc.Lock()

	found := false

	for i := 0; i < len(sc.activeNodes); i++ {
		if sc.activeNodes[i].identifier == topology.ID {
			found = true

			url := url.URL{
				Scheme: "http",
				Host: fmt.Sprintf("%s:%d", topology.Address,
					topology.APIPort),
			}

			sc.activeNodes[i].url = &url
			sc.activeNodes[i].currentTopology = hash

			break
		}
	}

	for i := 0; i < len(sc.inactiveNodes); i++ {
		if sc.inactiveNodes[i].identifier == topology.ID {
			found = true

			url := url.URL{
				Scheme: "http",
				Host: fmt.Sprintf("%s:%d", topology.Address,
					topology.APIPort),
			}

			sc.inactiveNodes[i].url = &url
			sc.inactiveNodes[i].currentTopology = hash

			break
		}
	}

	if sc.currentTopology != hash {
		sc.currentTopology = hash
		sc.currentTopologyCompiled = nil
	}

	dhosts := sc.denyHosts

	sc.Unlock()

	if !found {
		newNode := &SnowthNode{
			identifier: topology.ID,
			url: &url.URL{
				Scheme: "http",
				Host: fmt.Sprintf("%s:%d", topology.Address,
					topology.APIPort),
			},
			currentTopology: hash,
		}

		for _, dh := range dhosts {
			if newNode.GetURL().Host == dh {
				sc.LogWarnf("deny host found from topology: %s",
					newNode.GetURL().Host)

				return
			}
		}

		stats, err := sc.GetStatsNodeContext(ctx, newNode)
		if err != nil {
			// This node is not returning stats, put it on the inactive list.
			sc.AddNodes(newNode)

			return
		}

		newNode.identifier = stats.Identity()
		newNode.semVer = stats.SemVer()

		sc.AddNodes(newNode)
		sc.ActivateNodes(newNode)
	}
}

// ActivateNodes makes provided nodes active.
func (sc *SnowthClient) ActivateNodes(nodes ...*SnowthNode) {
	sc.Lock()
	defer sc.Unlock()

	in := []*SnowthNode{}
	match := false

	for _, iv := range sc.inactiveNodes {
		match = false

		for _, v := range nodes {
			if v.GetURL().String() == iv.GetURL().String() {
				match = true

				break
			}
		}

		if !match {
			in = append(in, iv)
		}
	}

	sc.inactiveNodes = in
	an := []*SnowthNode{}

	for _, v := range nodes {
		match = false

		for _, av := range sc.activeNodes {
			if v.GetURL().String() == av.GetURL().String() {
				match = true

				break
			}
		}

		if !match {
			an = append(an, v)
		}
	}

	sc.activeNodes = append(sc.activeNodes, an...)
}

// DeactivateNodes makes provided nodes inactive.
func (sc *SnowthClient) DeactivateNodes(nodes ...*SnowthNode) {
	sc.Lock()
	defer sc.Unlock()

	an := []*SnowthNode{}
	match := false

	for _, av := range sc.activeNodes {
		match = false

		for _, v := range nodes {
			if v.GetURL().String() == av.GetURL().String() {
				match = true

				break
			}
		}

		if !match {
			an = append(an, av)
		}
	}

	sc.activeNodes = an
	in := []*SnowthNode{}

	for _, v := range nodes {
		match = false

		for _, iv := range sc.inactiveNodes {
			if v.GetURL().String() == iv.GetURL().String() {
				match = true

				break
			}
		}

		if !match {
			in = append(in, v)
		}
	}

	sc.inactiveNodes = append(sc.inactiveNodes, in...)
}

// AddNodes adds node values to the inactive node list.
func (sc *SnowthClient) AddNodes(nodes ...*SnowthNode) {
	sc.Lock()
	defer sc.Unlock()

	in := []*SnowthNode{}
	match := false

	for _, v := range nodes {
		match = false

		for _, iv := range sc.inactiveNodes {
			if v.GetURL().String() == iv.GetURL().String() {
				match = true

				break
			}
		}

		if !match {
			in = append(in, v)
		}
	}

	sc.inactiveNodes = append(sc.inactiveNodes, in...)
}

// ListInactiveNodes lists all of the currently inactive nodes.
func (sc *SnowthClient) ListInactiveNodes() []*SnowthNode {
	sc.RLock()
	defer sc.RUnlock()

	result := []*SnowthNode{}
	result = append(result, sc.inactiveNodes...)

	return result
}

// ListActiveNodes lists all of the currently active nodes.
func (sc *SnowthClient) ListActiveNodes() []*SnowthNode {
	sc.RLock()
	defer sc.RUnlock()

	result := []*SnowthNode{}
	result = append(result, sc.activeNodes...)

	return result
}

// GetActiveNode returns a random active node in the cluster.
func (sc *SnowthClient) GetActiveNode(idsets ...[]string) *SnowthNode {
	sc.RLock()
	defer sc.RUnlock()

	if len(sc.activeNodes) == 0 {
		return nil
	}

	for _, ids := range idsets {
		for _, id := range ids {
			for _, node := range sc.activeNodes {
				if node.identifier == id {
					return node
				}
			}
		}
	}

	return sc.activeNodes[time.Now().UnixNano()%int64(len(sc.activeNodes))]
}

// DoRequest sends a request to IRONdb.
// If the client is set to retry using other nodes on network failures, this
// will perform those retries.
func (sc *SnowthClient) DoRequest(node *SnowthNode,
	method string, url string, body io.Reader,
	headers http.Header,
) (io.Reader, http.Header, error) {
	return sc.DoRequestContext(context.Background(), node, method, url,
		body, headers)
}

// DoRequestContext is the context aware version of DoRequest.
// If the client is set to retry using other nodes on network failures, this
// will perform those retries.
func (sc *SnowthClient) DoRequestContext(ctx context.Context, node *SnowthNode,
	method string, url string, body io.Reader,
	headers http.Header,
) (io.Reader, http.Header, error) {
	retries := sc.Retries()
	if retries < 0 {
		retries = 0
	}

	bBody := []byte{}

	var err error

	if body != nil {
		bBody, err = io.ReadAll(body)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read request body: %w", err)
		}
	}

	cr := sc.ConnectRetries()
	nodes := []*SnowthNode{node}

	for _, n := range sc.ListActiveNodes() {
		if n.GetURL().String() != node.GetURL().String() {
			nodes = append(nodes, n)
		}
	}

	var bdy io.Reader

	var hdr http.Header

	var status int

	traceID, ok := ctx.Value(sc.ctxKeyTraceID).(string)
	if !ok {
		traceID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	for r := int64(0); r < retries+1; r++ {
		connRetries := cr
		sns := nodes

		for len(sns) > 0 {
			n := int64(0)
			reqMsg := "attempting"

			if connRetries != cr {
				n = time.Now().UnixNano() % int64(len(sns))
				reqMsg = "retrying"
			}

			sn := sns[n]

			sns[n] = sns[len(sns)-1]
			sns = sns[:len(sns)-1]

			if sn == nil {
				continue
			}

			surl := sc.getURL(sn, url)

			sc.LogDebugf("gosnowth %s request "+
				"[traceID: %s, retry: %d, connRetry: %d]: %s %s %s",
				reqMsg, traceID, r, (cr - connRetries), method, surl,
				string(bBody))

			start := time.Now()

			bdy, hdr, status, err = sc.do(ctx, sn, method, surl,
				bytes.NewBuffer(bBody), headers, traceID)

			sc.LogDebugf("gosnowth request complete "+
				"[traceID: %s, retry: %d, connRetry: %d]: %s %s latency: %+v",
				traceID, r, (cr - connRetries), method, surl,
				time.Since(start))

			if err == nil {
				return bdy, hdr, nil
			}

			sc.LogWarnf("gosnowth request error "+
				"[retry: %d, connRetry: %d]: %s %s traceID: %s %+v",
				r, (cr - connRetries), method, surl, traceID, err)

			// Stop retrying other nodes if he context deadline was reached
			// or the context has been canceled.
			select {
			case <-ctx.Done():
				return bdy, hdr, ctx.Err()
			default:
			}

			// Do not retry 4xx status errors since these indicate a problem
			// with the request.
			if status >= http.StatusBadRequest &&
				status < http.StatusInternalServerError {
				return bdy, hdr, err
			}

			if strings.Contains(err.Error(), "cannot parse") ||
				strings.Contains(err.Error(), "User facing error") {
				return bdy, hdr, err
			}

			if connRetries == 0 {
				break
			}

			connRetries--
		}

		if r < retries {
			time.Sleep(time.Millisecond * time.Duration(100*2^r))
		}
	}

	return bdy, hdr, err
}

// do sends a request to IRONdb.
func (sc *SnowthClient) do(ctx context.Context,
	node *SnowthNode,
	method, url string,
	body io.Reader,
	headers http.Header,
	traceID string,
) (io.Reader, http.Header, int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	r, err := http.NewRequest(method, sc.getURL(node, url), body)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	sc.RLock()
	traceReq := sc.traceRequests != "" && (sc.traceRequests == "*" ||
		strings.HasPrefix(r.URL.Path, sc.traceRequests))
	dumpReq := sc.dumpRequests != "" && (sc.dumpRequests == "*" ||
		strings.HasPrefix(r.URL.Path, sc.dumpRequests))
	sc.RUnlock()

	r.Close = true

	for key, values := range headers {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}

	// Send a header telling snowth to use the gosnowth timeout - 1 second.
	if sc.timeout > 0 {
		if (sc.timeout - time.Second) > 0 {
			to := sc.timeout - time.Second

			r.Header.Set("X-Snowth-Timeout", to.String())
		} else {
			r.Header.Set("X-Snowth-Timeout", sc.timeout.String())
		}
	}

	r = r.WithContext(ctx)

	sc.RLock()

	rf := sc.request

	sc.RUnlock()

	if rf != nil {
		if err := rf(r); err != nil {
			return nil, nil, 0, fmt.Errorf("unable to process request: %w", err)
		}

		if r == nil {
			return nil, nil, 0, fmt.Errorf("invalid request after processing")
		}
	}

	if traceReq {
		ctrace := &httptrace.ClientTrace{
			GetConn: func(hostPort string) {
				sc.LogDebugf("gosnowth TRACE-%s: connecting %s\n",
					traceID, hostPort)
			},
			GotConn: func(info httptrace.GotConnInfo) {
				sc.LogDebugf("gosnowth TRACE-%s: connected %+v\n",
					traceID, info)
			},
			PutIdleConn: func(err error) {
				sc.LogDebugf("gosnowth TRACE-%s: put conn back in idle pool, err: %v\n",
					traceID, err)
			},
			GotFirstResponseByte: func() {
				sc.LogDebugf("gosnowth TRACE-%s: got first byte\n", traceID)
			},
			Got100Continue: func() {
				sc.LogDebugf("gosnowth TRACE-%s: got 100 Continue\n", traceID)
			},
			Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
				sc.LogDebugf("gosnowth TRACE-%s: %d %+v\n",
					traceID, code, header)

				return nil
			},
			DNSStart: func(info httptrace.DNSStartInfo) {
				sc.LogDebugf("gosnowth TRACE-%s: dns start %+v\n",
					traceID, info)
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				sc.LogDebugf("gosnowth TRACE-%s: dns done %+v\n",
					traceID, info)
			},
			ConnectStart: func(network, addr string) {
				sc.LogDebugf("gosnowth TRACE-%s: dialing %s/%s\n",
					traceID, network, addr)
			},
			ConnectDone: func(network, addr string, err error) {
				sc.LogDebugf("gosnowth TRACE-%s: dial complete %s/%s err: %v\n",
					traceID, network, addr, err)
			},
			WroteHeaderField: func(key string, values []string) {
				sc.LogDebugf("gosnowth TRACE-%s: wrote header %s: %+v\n",
					traceID, key, values)
			},
			WroteHeaders: func() {
				sc.LogDebugf("gosnowth TRACE-%s: wrote all headers\n", traceID)
			},
			Wait100Continue: func() {
				sc.LogDebugf("gosnowth TRACE-%s: waiting for '100 Continue' from server\n",
					traceID)
			},
			WroteRequest: func(info httptrace.WroteRequestInfo) {
				sc.LogDebugf("gosnowth TRACE-%s: wrote request %s %s - info: %+v\n",
					traceID, r.Method, r.URL.Path, info)
			},
		}

		r = r.WithContext(httptrace.WithClientTrace(r.Context(), ctrace))
	}

	if dumpReq {
		dump, err := httputil.DumpRequestOut(r, true)
		if err != nil {
			sc.LogErrorf("gosnowth error: %v traceID: %s", err, traceID)
		}

		sc.LogDebugf("gosnowth request dump: %s", string(dump))
	}

	sc.RLock()

	cli := sc.c

	sc.RUnlock()

	resp, err := cli.Do(r)
	if err != nil {
		return nil, nil, http.StatusInternalServerError,
			fmt.Errorf("failed to perform request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, resp.StatusCode,
			fmt.Errorf("unable to read response body: %w", err)
	}

	newTopo := resp.Header.Get("X-Topo-0")

	if newTopo != "" {
		sc.RLock()

		if newTopo != sc.currentTopology || newTopo != node.currentTopology {
			sc.RUnlock()

			sc.Lock()

			sc.currentTopology = newTopo
			sc.currentTopologyCompiled = nil

			node.currentTopology = newTopo

			sc.Unlock()
		} else {
			sc.RUnlock()
		}
	}

	if traceReq {
		msg := string(res[0:64]) + "..."
		if resp.StatusCode != http.StatusOK {
			msg = string(res)
		}

		sc.LogDebugf("gosnowth TRACE-%s: complete %s - %s\n",
			traceID, resp.Status, msg)
	}

	if resp.StatusCode != http.StatusOK {
		return bytes.NewBuffer(res), resp.Header, resp.StatusCode,
			fmt.Errorf("error returned from IRONdb (%s): [%d] %s",
				r.URL.Host, resp.StatusCode, string(res))
	}

	return bytes.NewBuffer(res), resp.Header, resp.StatusCode, nil
}

// getURL resolves the URL with a reference for a particular node.
func (sc *SnowthClient) getURL(node *SnowthNode, ref string) string {
	return resolveURL(node.url, ref)
}
