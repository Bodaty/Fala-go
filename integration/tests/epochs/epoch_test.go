package epochs

import (
	"context"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go/integration/utils"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/onflow/flow-go/integration/testnet"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/state/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEpochs(t *testing.T) {
	suite.Run(t, new(Suite))
}

// TestViewsProgress asserts epoch state transitions over two full epochs
// without any nodes joining or leaving.
func (s *Suite) TestViewsProgress() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// phaseCheck is a utility struct that contains information about the
	// final view of each epoch/phase.
	type phaseCheck struct {
		epoch     uint64
		phase     flow.EpochPhase
		finalView uint64 // the final view of the phase as defined by the EpochSetup
	}

	phaseChecks := []*phaseCheck{}
	// iterate through two epochs and populate a list of phase checks
	for counter := 0; counter < 2; counter++ {

		// wait until the access node reaches the desired epoch
		var epoch protocol.Epoch
		var epochCounter uint64
		for epoch == nil || epochCounter != uint64(counter) {
			snapshot, err := s.client.GetLatestProtocolSnapshot(ctx)
			require.NoError(s.T(), err)
			epoch = snapshot.Epochs().Current()
			epochCounter, err = epoch.Counter()
			require.NoError(s.T(), err)
		}

		epochFirstView, err := epoch.FirstView()
		require.NoError(s.T(), err)
		epochDKGPhase1Final, err := epoch.DKGPhase1FinalView()
		require.NoError(s.T(), err)
		epochDKGPhase2Final, err := epoch.DKGPhase2FinalView()
		require.NoError(s.T(), err)
		epochDKGPhase3Final, err := epoch.DKGPhase3FinalView()
		require.NoError(s.T(), err)
		epochFinal, err := epoch.FinalView()
		require.NoError(s.T(), err)

		epochViews := []*phaseCheck{
			{epoch: epochCounter, phase: flow.EpochPhaseStaking, finalView: epochFirstView},
			{epoch: epochCounter, phase: flow.EpochPhaseSetup, finalView: epochDKGPhase1Final},
			{epoch: epochCounter, phase: flow.EpochPhaseSetup, finalView: epochDKGPhase2Final},
			{epoch: epochCounter, phase: flow.EpochPhaseSetup, finalView: epochDKGPhase3Final},
			{epoch: epochCounter, phase: flow.EpochPhaseCommitted, finalView: epochFinal},
		}

		for _, v := range epochViews {
			s.BlockState.WaitForSealedView(s.T(), v.finalView)
		}

		phaseChecks = append(phaseChecks, epochViews...)
	}

	s.net.StopContainers()

	consensusContainers := make([]*testnet.Container, 0)
	for _, c := range s.net.Containers {
		if c.Config.Role == flow.RoleConsensus {
			consensusContainers = append(consensusContainers, c)
		}
	}

	for _, c := range consensusContainers {
		containerState, err := c.OpenState()
		require.NoError(s.T(), err)

		// create a map of [view] => {epoch-counter, phase}
		lookup := map[uint64]struct {
			epochCounter uint64
			phase        flow.EpochPhase
		}{}

		final, err := containerState.Final().Head()
		require.NoError(s.T(), err)

		var h uint64
		for h = 0; h <= final.Height; h++ {
			snapshot := containerState.AtHeight(h)

			head, err := snapshot.Head()
			require.NoError(s.T(), err)

			epoch := snapshot.Epochs().Current()
			currentEpochCounter, err := epoch.Counter()
			require.NoError(s.T(), err)
			currentPhase, err := snapshot.Phase()
			require.NoError(s.T(), err)

			lookup[head.View] = struct {
				epochCounter uint64
				phase        flow.EpochPhase
			}{
				currentEpochCounter,
				currentPhase,
			}
		}

		for _, v := range phaseChecks {
			item := lookup[v.finalView]
			assert.Equal(s.T(), v.epoch, item.epochCounter, "wrong epoch at view %d", v.finalView)
			assert.Equal(s.T(), v.phase, item.phase, "wrong phase at view %d", v.finalView)
		}
	}
}

// TestEpochJoin will asserts that a node can successfully join the protocol by staking during
// the Epoch Staking Phase and asserting that node info read from the staking table match the
// node info generated by the test
func (s *Suite) TestEpochJoin() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env := utils.LocalnetEnv()

	// stake a new node
	info := s.StakeNode(ctx, env, flow.RoleConsensus)

	// get node info from staking table
	nodeInfoCDC := s.ExecuteGetNodeInfoScript(ctx, env, info.NodeID)
	nodeInfo, ok := nodeInfoCDC.(cadence.Struct)
	require.True(s.T(), ok)

	// make sure node info we generated matches what we get from the flow staking table
	nodeID := string(nodeInfo.Fields[0].(cadence.String))
	require.Equal(s.T(), info.NodeID.String(), nodeID, "expected node ID generated in test to equal node ID from staking table ")

	result := s.SetApprovedNodesScript(ctx, env, info.NodeID)
	require.NoError(s.T(), result.Error)

	s.net.StopContainers()
}
