package status_test

import (
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/state_synchronization/requester/status"
	synctest "github.com/onflow/flow-go/module/state_synchronization/requester/unittest"
	"github.com/onflow/flow-go/storage"
	"github.com/onflow/flow-go/utils/unittest"
)

// TestSealedBlockReader evaluates that block reader correctly reads stored finalized blocks from the blocks storage and
// protocol state.
func TestSealedBlockReader(t *testing.T) {
	withReader(t, 10, func(reader *status.SealedBlockReader, blocks []*flow.Block) {
		// the last block seals its parent
		lastSealedBlock := blocks[len(blocks)-2]

		// head of the reader is the last sealed block
		head, err := reader.Head()
		assert.NoError(t, err)
		assert.Equal(t, head, lastSealedBlock.Header.Height, "head does not match last sealed block")

		// retrieved blocks from block reader should be the same as the original blocks stored in it.
		// all except the last block should be sealed
		lastIndex := len(blocks)
		for _, actual := range blocks[:lastIndex-1] {
			index := actual.Header.Height
			job, err := reader.AtIndex(index)
			assert.NoError(t, err)

			retrieved, err := status.JobToBlock(job)
			assert.NoError(t, err)
			assert.Equal(t, actual.ID(), retrieved.ID())
		}

		// ensure the last block returns a NotFound error
		job, err := reader.AtIndex(uint64(lastIndex))
		assert.Nil(t, job)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})
}

// withReader is a test helper that sets up a block reader.
// It also provides a chain of specified number of finalized blocks ready to read by block reader, i.e., the protocol state is extended with the
// chain of blocks and the blocks are stored in blocks storage.
func withReader(
	t *testing.T,
	blockCount int,
	withBlockReader func(*status.SealedBlockReader, []*flow.Block),
) {
	require.Equal(t, blockCount%2, 0, "block count for this test should be even")
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {

		blocks := make([]*flow.Block, blockCount)
		blocksByHeight := make(map[uint64]*flow.Block, blockCount)

		var seals []*flow.Header
		parent := unittest.GenesisFixture().Header
		for i := 0; i < blockCount; i++ {
			seals = []*flow.Header{parent}
			height := uint64(i) + 1

			blocks[i] = unittest.BlockWithParentAndSeals(parent, seals)
			blocksByHeight[height] = blocks[i]

			parent = blocks[i].Header
		}

		snapshot := synctest.MockProtocolStateSnapshot(synctest.WithHead(seals[0]))
		state := synctest.MockProtocolState(synctest.WithSnapshot(snapshot))
		headerStorage := synctest.MockBlockHeaderStorage(synctest.WithByHeight(blocksByHeight))

		reader := status.NewSealedBlockReader(state, headerStorage)

		withBlockReader(reader, blocks)
	})
}
