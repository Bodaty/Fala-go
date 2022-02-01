package cmd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/sethvargo/go-retry"

	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go/cmd/util/cmd/common"
	"github.com/onflow/flow-go/model/bootstrap"

	"github.com/onflow/flow-go/crypto"
	"github.com/onflow/flow-go/state/protocol"
	utilsio "github.com/onflow/flow-go/utils/io"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/state/protocol/inmem"
)

// GetProtocolSnapshot callback that will get latest finalized protocol snapshot
type GetProtocolSnapshot func(context.Context) (protocol.Snapshot, error)

// GetSnapshot will attempt to get the latest finalized protocol snapshot with the given flow configs
func GetSnapshot(ctx context.Context, client *client.Client) (*inmem.Snapshot, error) {
	b, err := client.GetLatestProtocolStateSnapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest finalized protocol state snapshot during pre-initialization: %w", err)
	}

	var snapshotEnc inmem.EncodableSnapshot
	err = json.Unmarshal(b, &snapshotEnc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal protocol state snapshot: %w", err)
	}

	snapshot := inmem.SnapshotFromEncodable(snapshotEnc)
	return snapshot, nil
}

// GetSnapshotAtEpochAndPhase will get the latest finalized protocol snapshot and check the current epoch and epoch phase.
// If we are past the target epoch and epoch phase we exit the retry mechanism immediately.
// If not check the snapshot at the specified interval until we reach the target epoch and phase.
func GetSnapshotAtEpochAndPhase(logger zerolog.Logger, startupEpoch uint64, startupEpochPhase flow.EpochPhase, retryInterval time.Duration, getSnapshot GetProtocolSnapshot) (protocol.Snapshot, error) {
	start := time.Now()
	var snapshot protocol.Snapshot

	constRetry, err := retry.NewConstant(retryInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to create retry mechanism: %w", err)
	}

	err = retry.Do(context.Background(), constRetry, func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		logger.Info().
			Dur("time-waiting", time.Since(start)).
			Msg("retrieving snapshot...")
		snapshot, err = getSnapshot(ctx)
		if err != nil {
			logger.Info().Err(err).Msg("failed to get snapshot")
			return retry.RetryableError(fmt.Errorf("failed to get protocol snapshot: %w", err))
		}

		// if we encounter any errors interpreting the snapshot something went wrong stop retrying
		currEpochCounter, err := snapshot.Epochs().Current().Counter()
		if err != nil {
			return fmt.Errorf("failed to get the current epoch counter: %w", err)
		}

		currEpochPhase, err := snapshot.Phase()
		if err != nil {
			return fmt.Errorf("failed to get the current epoch phase: %w", err)
		}

		// check if we are in or past the target epoch and phase
		if currEpochCounter > startupEpoch || (currEpochCounter == startupEpoch && currEpochPhase >= startupEpochPhase) {
			logger.Info().
				Dur("time-waiting", time.Since(start)).
				Uint64("current-epoch", currEpochCounter).
				Str("current-epoch-phase", currEpochPhase.String()).
				Msg("reached desired epoch and phase in dynamic startup pre-init")

			return nil
		}

		logger.Warn().
			Dur("time-waiting", time.Since(start)).
			Uint64("current-epoch", currEpochCounter).
			Str("current-epoch-phase", currEpochPhase.String()).
			Msgf("waiting for epoch %d and phase %s", startupEpoch, startupEpochPhase.String())

		return retry.RetryableError(fmt.Errorf("dynamic startup epoch and epoch phase not reached"))
	})
	if err != nil {
		return nil, fmt.Errorf("failed to wait for target epoch and phase: %w", err)
	}

	return snapshot, nil
}

// ValidateDynamicStartupFlags will validate flags necessary for dynamic node startup
// - assert dynamic-startup-access-publickey  is valid ECDSA_P256 public key hex
// - assert dynamic-startup-access-address is not empty
// - assert dynamic-startup-startup-epoch-phase is > 0 (EpochPhaseUndefined)
// - assert dynamic-startup-startup-epoch is > 0
func ValidateDynamicStartupFlags(accessPublicKey, accessAddress string, startPhase flow.EpochPhase) error {
	b, err := hex.DecodeString(strings.TrimPrefix(accessPublicKey, "0x"))
	if err != nil {
		return fmt.Errorf("invalid flag --dynamic-startup-access-publickey: %w", err)
	}

	_, err = crypto.DecodePublicKey(crypto.ECDSAP256, b)
	if err != nil {
		return fmt.Errorf("invalid flag --dynamic-startup-access-publickey: %w", err)
	}

	if accessAddress == "" {
		return fmt.Errorf("invalid flag --dynamic-startup-access-address can not be empty")
	}

	if startPhase <= flow.EpochPhaseUndefined {
		return fmt.Errorf("invalid flag --dynamic-startup-startup-epoch-phase unknow epoch phase")
	}

	return nil
}

// DynamicStartPreInit will attempt to save a the latest finalizes protocol snapshot
// if we have reached the target epoch and phase are
func DynamicStartPreInit(nodeConfig *NodeConfig) error {
	rootSnapshotPath := filepath.Join(nodeConfig.BootstrapDir, bootstrap.PathRootProtocolStateSnapshot)

	log := nodeConfig.Logger.With().Str("component", "dynamic_start").Logger()

	// root snapshot exists skip dynamic start up
	if rootSnapshotExists(rootSnapshotPath) {
		log.Info().Msg("starting up using existing root snapshot file")
		return nil
	}

	startupPhase := flow.GetEpochPhase(nodeConfig.BaseConfig.DynamicStartupEpochPhase)

	// validate dynamic startup flags
	err := ValidateDynamicStartupFlags(nodeConfig.BaseConfig.DynamicStartupANPubkey, nodeConfig.BaseConfig.DynamicStartupANAddress, startupPhase)
	if err != nil {
		return err
	}

	// get flow clinet with secure client connection to download protocol snapshot from access node
	config, err := common.NewFlowClientConfig(nodeConfig.DynamicStartupANAddress, nodeConfig.DynamicStartupANPubkey, false)
	if err != nil {
		return fmt.Errorf("failed to create flow client config for node dynamic startup pre-init: %w", err)
	}

	flowClient, err := common.FlowClient(config)
	if err != nil {
		return fmt.Errorf("failed to create flow client for node dynamic startup pre-init: %w", err)
	}

	getSnapshotFunc := func(ctx context.Context) (protocol.Snapshot, error) {
		return GetSnapshot(ctx, flowClient)
	}

	snapshot, err := GetSnapshotAtEpochAndPhase(
		nodeConfig.Logger,
		nodeConfig.BaseConfig.DynamicStartupEpoch,
		startupPhase,
		nodeConfig.BaseConfig.DynamicStartupSleepInterval,
		getSnapshotFunc,
	)
	if err != nil {
		return fmt.Errorf("failed to get snapshot at start up epoch (%d) and phase (%s): %w", nodeConfig.BaseConfig.DynamicStartupEpoch, startupPhase.String(), err)
	}

	nodeConfig.Logger.Info().Str("snapshot-path", rootSnapshotPath).Msg("writing root snapshot file")
	err = writeRootSnapshotFile(snapshot.(*inmem.Snapshot), rootSnapshotPath)
	if err != nil {
		return fmt.Errorf("failed to write root snapshot file: %w", err)
	}

	return nil
}

// rootSnapshotExists check if root snapshot file exists
func rootSnapshotExists(path string) bool {
	return utilsio.FileExists(path)
}

// writeRootSnapshotFile will write the snapshot to the path provided
func writeRootSnapshotFile(snapshot *inmem.Snapshot, path string) error {
	err := utilsio.WriteJSON(path, snapshot.Encodable())
	if err != nil {
		return fmt.Errorf("failed to write root snapshot file: %w", err)
	}

	return nil
}
