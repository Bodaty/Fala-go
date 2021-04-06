package requester

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/engine"
	mockfetcher "github.com/onflow/flow-go/engine/verification/fetcher/mock"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/model/verification"
	mempool "github.com/onflow/flow-go/module/mempool/mock"
	"github.com/onflow/flow-go/module/mock"
	"github.com/onflow/flow-go/network/mocknetwork"
	"github.com/onflow/flow-go/utils/unittest"

	protocol "github.com/onflow/flow-go/state/protocol/mock"
)

// RequesterEngineTestSuite encapsulates data structures for running unittests on requester engine.
type RequesterEngineTestSuite struct {
	// modules
	log             zerolog.Logger
	handler         *mockfetcher.ChunkDataPackHandler // contains callbacks for handling received chunk data packs.
	retryInterval   time.Duration                     // determines time in milliseconds for retrying chunk data requests.
	pendingRequests *mempool.ChunkRequests            // used to store all the pending chunks that assigned to this node
	state           *protocol.State                   // used to check the last sealed height
	con             *mocknetwork.Conduit              // used to send chunk data request, and receive the response

	// identities
	verIdentity *flow.Identity // verification node
}

// setupTest initiates a test suite prior to each test.
func setupTest() *RequesterEngineTestSuite {
	r := &RequesterEngineTestSuite{
		log:             unittest.Logger(),
		handler:         &mockfetcher.ChunkDataPackHandler{},
		retryInterval:   100 * time.Millisecond,
		pendingRequests: &mempool.ChunkRequests{},
		state:           &protocol.State{},
		verIdentity:     unittest.IdentityFixture(unittest.WithRole(flow.RoleVerification)),
		con:             &mocknetwork.Conduit{},
	}

	return r
}

// newRequesterEngine returns a requester engine for testing.
func newRequesterEngine(t *testing.T, s *RequesterEngineTestSuite) *Engine {
	net := &mock.Network{}
	// mocking the network registration of the engine
	net.On("Register", engine.RequestChunks, testifymock.Anything).
		Return(s.con, nil).
		Once()

	e, err := New(s.log, s.state, net, s.retryInterval, s.pendingRequests, s.handler)
	require.NoError(t, err)

	testifymock.AssertExpectationsForObjects(t, net)

	return e
}

// TestHandleChunkDataPack_HappyPath evaluates the happy path of receiving a requested chunk data pack.
// The chunk data pack should be passed to the registered handler, and the resources should be cleaned up.
func TestHandleChunkDataPack_HappyPath(t *testing.T) {
	s := setupTest()
	e := newRequesterEngine(t, s)

	response := unittest.ChunkDataResponseFixture()
	originID := unittest.IdentifierFixture()

	// we have a request pending for this response chunk ID
	s.pendingRequests.On("ByID", response.ChunkDataPack.ChunkID).Return(&verification.ChunkRequestStatus{}, true).Once()
	// we remove pending request on receiving this response
	s.pendingRequests.On("Rem", response.ChunkDataPack.ChunkID).Return(true).Once()

	s.handler.On("HandleChunkDataPack", originID, &response.ChunkDataPack, &response.Collection).Return().Once()

	err := e.Process(originID, response)
	require.Nil(t, err)

	testifymock.AssertExpectationsForObjects(t, s.pendingRequests, s.con, s.handler)
}

// TestHandleChunkDataPack_NonExistingRequest evaluates that receiving a chunk data pack response that does not have any request attached
// is dropped without passing it to the handler.
func TestHandleChunkDataPack_NonExistingRequest(t *testing.T) {
	s := setupTest()
	e := newRequesterEngine(t, s)

	response := unittest.ChunkDataResponseFixture()
	originID := unittest.IdentifierFixture()

	// we have a request pending for this response chunk ID
	s.pendingRequests.On("ByID", response.ChunkDataPack.ChunkID).Return(nil, false).Once()

	err := e.Process(originID, response)
	require.Nil(t, err)

	testifymock.AssertExpectationsForObjects(t, s.pendingRequests, s.con)
	s.handler.AssertNotCalled(t, "HandleChunkDataPack")
	s.pendingRequests.AssertNotCalled(t, "Rem")
}

// TestHandleChunkDataPack_NonExistingRequest evaluates that failing to remove a received chunk data pack's request
// from the memory terminates the procedure of handling a chunk data pack without passing it to the handler.
// The request for a chunk data pack may be removed from the memory if duplicate copies of a requested chunk data pack arrive
// concurrently. Then the mutex lock on pending requests mempool allows only one of those requested chunk data packs to remove the
// request and pass to handler. While handling the other ones gracefully terminated.
func TestHandleChunkDataPack_FailedRequestRemoval(t *testing.T) {
	s := setupTest()
	e := newRequesterEngine(t, s)

	response := unittest.ChunkDataResponseFixture()
	originID := unittest.IdentifierFixture()

	// we have a request pending for this response chunk ID
	s.pendingRequests.On("ByID", response.ChunkDataPack.ChunkID).Return(&verification.ChunkRequestStatus{}, true).Once()
	// however by the time we try remove it, the request status has gone.
	// this can happen when duplicate chunk data packs are coming concurrently.
	// the concurrency is safe with pending requests mempool lock.
	s.pendingRequests.On("Rem", response.ChunkDataPack.ChunkID).Return(false).Once()

	err := e.Process(originID, response)
	require.Nil(t, err)

	testifymock.AssertExpectationsForObjects(t, s.pendingRequests, s.con)
	s.handler.AssertNotCalled(t, "HandleChunkDataPack")
}
