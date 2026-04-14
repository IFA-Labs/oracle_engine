// This package is to be deprecated in favor of the CRE workflow which is more robust and has better error handling.
package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/utils"

	"oracle_engine/internal/database/timescale"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	importVerifier "oracle_engine/pkg/abi"
)

// / Relayer is a service that takes issuances requests
// / and sends to the contract
// / It also updates the status of the issuance request
// / in local db (also the asset instance)
type Relayer struct {
	cfg                    *config.Config
	contractToRoutineChMap map[string]chan *models.Issuance
	chainLocks             map[string]*sync.Mutex
	db                     *timescale.TimescaleDB
}

func New(config *config.Config, db *timescale.TimescaleDB) *Relayer {
	return &Relayer{
		cfg:                    config,
		contractToRoutineChMap: make(map[string]chan *models.Issuance),
		chainLocks:             make(map[string]*sync.Mutex),
		db:                     db,
	}
}

// / Start treat latest issuance with utmost priority
// / Start a go routine for each issuance
// / Each contract has its own go routine
func (r *Relayer) Start(ctx context.Context) error {
	bufferSize := r.cfg.RelayerBatch.ChannelBuffer
	if bufferSize <= 0 {
		bufferSize = 256
	}

	for _, ctrct := range r.cfg.Contracts {
		contractKey := r.contractKey(ctrct)
		r.contractToRoutineChMap[contractKey] = make(chan *models.Issuance, bufferSize)
		if _, ok := r.chainLocks[ctrct.ChainID]; !ok {
			r.chainLocks[ctrct.ChainID] = &sync.Mutex{}
		}
		go r.startRoutine(ctx, ctrct, r.contractToRoutineChMap[contractKey])
	}

	<-ctx.Done()
	logging.Logger.Info("Relayer routine stopped", zap.Error(ctx.Err()))
	return ctx.Err()

}

func (r *Relayer) AcceptIssuance(issuance *models.Issuance) error {
	logging.Logger.Debug("Issuance accepted", zap.String("assetID", issuance.Price.AssetID))
	if len(r.contractToRoutineChMap) == 0 {
		return fmt.Errorf("no relayer contract routines are active")
	}

	for contractKey, ch := range r.contractToRoutineChMap {
		logging.Logger.Info(
			"Sending issuance to contract channel",
			zap.String("assetID", issuance.Price.AssetID),
			zap.String("contract", contractKey),
		)
		ch <- issuance
	}
	return nil
}

func (r *Relayer) startRoutine(ctx context.Context, ctrct config.ContractConfig, ch <-chan *models.Issuance) {
	maxBatch := r.cfg.RelayerBatch.MaxIssuances
	if maxBatch <= 0 {
		maxBatch = 20
	}

	flushEvery := time.Duration(r.cfg.RelayerBatch.FlushIntervalSeconds) * time.Second
	if flushEvery <= 0 {
		flushEvery = 3 * time.Second
	}

	ticker := time.NewTicker(flushEvery)
	defer ticker.Stop()

	batch := make([]*models.Issuance, 0, maxBatch)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := r.ConveyBatchIssuancesToContract(ctx, batch, ctrct); err != nil {
			logging.Logger.Error("Failed to convey issuance batch", zap.Error(err), zap.String("contract", r.contractKey(ctrct)))
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case issuance, ok := <-ch:
			if !ok {
				flush()
				return
			}
			if issuance == nil {
				continue
			}
			batch = append(batch, issuance)
			if len(batch) >= maxBatch {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (r *Relayer) ConveyIssuanceToContract(ctx context.Context, issuance *models.Issuance, ctrct config.ContractConfig) error {
	return r.ConveyBatchIssuancesToContract(ctx, []*models.Issuance{issuance}, ctrct)
}

func (r *Relayer) ConveyBatchIssuancesToContract(ctx context.Context, issuances []*models.Issuance, ctrct config.ContractConfig) error {
	if len(issuances) == 0 {
		return nil
	}

	logging.Logger.Debug(
		"conveying issuance batch to contract",
		zap.Int("batchSize", len(issuances)),
		zap.String("chainId", ctrct.ChainID),
	)

	rpcUrl := ctrct.RPC
	if rpcUrl == "" {
		rpcUrl = os.Getenv("ALCHEMY_URL")
		if rpcUrl == "" {
			logging.Logger.Error("RPC URL not set")
			return fmt.Errorf("RPC URL not set")
		}
	}
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		logging.Logger.Error("Failed to connect to Ethereum client", zap.Error(err))
		return err
	}
	privateKey, err := crypto.HexToECDSA(r.cfg.PrivateKey)
	if err != nil {
		logging.Logger.Error("Failed to load private key", zap.Error(err))
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logging.Logger.Error("Failed to assert public key type", zap.Error(err))
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	chainLock := r.chainLocks[ctrct.ChainID]
	if chainLock == nil {
		chainLock = &sync.Mutex{}
		r.chainLocks[ctrct.ChainID] = chainLock
	}

	chainLock.Lock()
	defer chainLock.Unlock()

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		logging.Logger.Error("Failed to get nonce", zap.Error(err), zap.String("chainId", ctrct.ChainID))
		return err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		logging.Logger.Error("Failed to suggest gas price", zap.Error(err))
		return err
	}

	chainID, err := strconv.ParseInt(ctrct.ChainID, 10, 64)
	if err != nil {
		logging.Logger.Error("Failed to parse chain ID", zap.Error(err))
		return err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		logging.Logger.Error("Failed to create new keyed transactor", zap.Error(err))
		return err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasPrice = gasPrice

	address := common.HexToAddress(ctrct.Address)
	// Load the verifier contract
	contract, err := importVerifier.NewVerifier(address, client)
	if err != nil {
		logging.Logger.Error("Failed to load verifier contract", zap.Error(err))
		return fmt.Errorf("failed to load verifier contract: %w", err)
	}

	// Prepare inputs
	latestByAsset := r.latestIssuancesByAsset(issuances)
	assetIDs := make([]string, 0, len(latestByAsset))
	for assetID := range latestByAsset {
		assetIDs = append(assetIDs, assetID)
	}
	sort.Strings(assetIDs)

	assetIndex := make([][32]byte, 0, len(assetIDs))
	prices := make([]importVerifier.IIfaPriceFeedPriceFeed, 0, len(assetIDs))
	for _, assetID := range assetIDs {
		issuance := latestByAsset[assetID]
		assetIndex = append(assetIndex, utils.HexToBytes32(assetID))
		prices = append(prices, importVerifier.IIfaPriceFeedPriceFeed{
			Price:          utils.Float64ToBigInt(issuance.Price.Value),
			Decimal:        int8(issuance.Price.Expo),
			LastUpdateTime: uint64(issuance.Price.Timestamp.Unix()),
		})
	}

	tx, err := contract.SubmitPriceFeed(auth, assetIndex, prices)
	if err != nil {
		logging.Logger.Error(
			"Failed to submit price feed",
			zap.Int64("chainID", chainID),
			zap.Int("feedCount", len(prices)),
			zap.String("Contract", address.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to submit price feed: %w", err)
	}

	logging.Logger.Info(
		"Submitted price feed batch",
		zap.String("tx", tx.Hash().Hex()),
		zap.String("chainID", ctrct.ChainID),
		zap.Int("requestedIssuances", len(issuances)),
		zap.Int("submittedFeeds", len(prices)),
	)

	return nil
}

func (r *Relayer) latestIssuancesByAsset(issuances []*models.Issuance) map[string]*models.Issuance {
	latestByAsset := make(map[string]*models.Issuance)
	for _, issuance := range issuances {
		if issuance == nil {
			continue
		}
		existing, ok := latestByAsset[issuance.Price.AssetID]
		if !ok || issuance.Price.Timestamp.After(existing.Price.Timestamp) {
			latestByAsset[issuance.Price.AssetID] = issuance
		}
	}
	return latestByAsset
}

func (r *Relayer) contractKey(ctrct config.ContractConfig) string {
	return fmt.Sprintf("%s:%s", ctrct.ChainID, ctrct.Address)
}

// abigen --bin=./build/Store.bin --abi=./build/Store.abi --pkg=store --out=Store.go
