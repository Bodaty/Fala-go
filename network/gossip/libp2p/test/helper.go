package test

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/module/mock"
	"github.com/dapperlabs/flow-go/network/codec/json"
	"github.com/dapperlabs/flow-go/network/gossip/libp2p"
	protocol "github.com/dapperlabs/flow-go/protocol/mock"
)

func CreateSubnets(nodesNum, subnetNum, linkNum int) (map[int][]*libp2p.Network, map[int][]flow.Identifier, error) {
	if nodesNum < subnetNum || nodesNum%subnetNum != 0 {
		return nil, nil, fmt.Errorf("number of subnets should divide number of nodes")
	}

	// subSize of subnets
	subSize := nodesNum / subnetNum
	if subSize < linkNum {
		return nil, nil,
			fmt.Errorf("number of links (%d) greater than subnet size (%d)", subSize, linkNum)
	}

	// map of subnet ids
	subnetIdList := groupIDs(CreateIDs(nodesNum), subSize, linkNum)

	// nets keeps a map of subnets networks
	nets := make(map[int][]*libp2p.Network)
	ids := make(map[int][]flow.Identifier)

	// creates topology over the nodes of each subnet
	for index, subids := range subnetIdList {
		logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
		mws, err := CreateMiddleware(logger, subids)
		if err != nil {
			return nil, nil, err
		}

		// creates a network for that subnet of ids
		subnet, subidf, err := CreateNetworks(mws, subids, nodesNum, false)
		if err != nil {
			return nil, nil, err
		}

		// adds subnet to the map of subnets
		nets[index] = subnet
		ids[index] = subidf
	}

	// allows nodes to find each other
	time.Sleep(10 * time.Second)

	return nets, ids, nil
}

// groupIDs groups ids into sub-many groups
func groupIDs(ids []*flow.Identity, subNum, linkNum int) map[int][]*flow.Identity {
	sIndx := -1
	subnets := make(map[*flow.Identity]int)

	// keeps ids of the nodes in each subnet
	subnetIdList := make(map[int][]*flow.Identity)
	// clusters ids into subnets
	for index, id := range ids {
		// checks if we reach end of current subnet
		if index%subNum == 0 {
			// moves to next subnet
			sIndx++
		}

		// assigns subnet index of the net
		subnets[id] = sIndx
		if subnetIdList[sIndx] == nil {
			// initializes list
			subnetIdList[sIndx] = make([]*flow.Identity, 0)
		}

		// adding nodes' id to the current Identity
		subnetIdList[sIndx] = append(subnetIdList[sIndx], id)
	}

	// creates links between subnets
	if linkNum > 0 {
		for this := range subnetIdList {
			for other := range subnetIdList {
				if this == other {
					continue
				}

				// adds links from this subnet to other subnet
				for i := 0; i < linkNum; i++ {
					subnetIdList[other] = append(subnetIdList[other], subnetIdList[this][i])
				}
			}
		}
	}

	return subnetIdList
}

// helper offers a set of functions that are shared among different tests
// CreateIDs creates and initializes count-many flow identifiers instances
func CreateIDs(count int) []*flow.Identity {
	identities := make([]*flow.Identity, 0)
	for i := 0; i < count; i++ {
		// defining id of node
		var nodeID [32]byte
		nodeID[0] = byte(i + 1)
		identity := &flow.Identity{
			NodeID: nodeID,
		}
		identities = append(identities, identity)
	}
	return identities
}

// createNetworks receives a slice of middlewares their associated flow identifiers,
// and for each middleware creates a network instance on top
// it returns the slice of created middlewares
// csize is the receive cache size of the nodes
// dryrun is a boolean for topology generation tests, in case it is switched to true, the ids is directly
// mocked for every node as their state, otherwise, ids is only used to assign the ids of the nodes, and the
// state is constructed based on the IPs collected from libp2p
func CreateNetworks(mws []*libp2p.Middleware, ids flow.IdentityList, csize int, dryrun bool) ([]*libp2p.Network, []flow.Identifier, error) {
	count := len(mws)
	nets := make([]*libp2p.Network, 0)

	// creates and mocks the state
	var snapshot *SnapshotMock
	if dryrun {
		snapshot = &SnapshotMock{ids: ids}
	} else {
		snapshot = &SnapshotMock{ids: flow.IdentityList{}}
	}
	state := &protocol.State{}
	state.On("Final").Return(snapshot)

	for i := 0; i < count; i++ {
		// creates and mocks me
		me := &mock.Local{}
		me.On("NodeID").Return(ids[i].NodeID)
		logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
		net, err := libp2p.NewNetwork(logger, json.NewCodec(), state, me, mws[i], csize)
		if err != nil {
			return nil, nil, fmt.Errorf("could not create error %w", err)
		}

		nets = append(nets, net)
	}

	// if dryrun then don't actually start the network just return the network objects
	if dryrun {
		return nets, snapshot.ids.NodeIDs(), nil
	}

	for _, net := range nets {
		<-net.Ready()
	}

	for i, m := range mws {
		// retrieves IP and port of the middleware
		ip, port := m.GetIPPort()

		// mocks an identity for the middleware
		id := &flow.Identity{
			NodeID:  ids[i].NodeID,
			Address: fmt.Sprintf("%s:%s", ip, port),
			Role:    flow.RoleCollection,
			Stake:   0,
		}
		snapshot.ids = append(snapshot.ids, id)
	}

	return nets, snapshot.ids.NodeIDs(), nil
}

// CreateMiddleware receives an ids slice and creates and initializes a middleware instances for each id
func CreateMiddleware(log zerolog.Logger, identities []*flow.Identity) ([]*libp2p.Middleware, error) {
	count := len(identities)
	mws := make([]*libp2p.Middleware, 0)
	for i := 0; i < count; i++ {
		// creating middleware of nodes
		mw, err := libp2p.NewMiddleware(log, json.NewCodec(), "0.0.0.0:0", identities[i].NodeID)
		if err != nil {
			return nil, err
		}

		mws = append(mws, mw)
	}
	return mws, nil
}

type SnapshotMock struct {
	ids flow.IdentityList
}

func (s *SnapshotMock) Identities(filters ...flow.IdentityFilter) (flow.IdentityList, error) {
	return s.ids, nil
}

func (s *SnapshotMock) Identity(nodeID flow.Identifier) (*flow.Identity, error) {
	return nil, fmt.Errorf(" not implemented")
}

func (s *SnapshotMock) Commit() (flow.StateCommitment, error) {
	return nil, fmt.Errorf(" not implemented")
}

func (s *SnapshotMock) Clusters() (*flow.ClusterList, error) {
	return nil, fmt.Errorf(" not implemented")
}

func (s *SnapshotMock) Head() (*flow.Header, error) {
	return nil, fmt.Errorf(" not implemented")
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
