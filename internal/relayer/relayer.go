package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/utils"
	"os"
	"strconv"

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
	cfg                 *config.Config
	assetToRoutineChMap map[string]chan *models.Issuance
	db                  *timescale.TimescaleDB
}

func New(config *config.Config, db *timescale.TimescaleDB) *Relayer {
	return &Relayer{
		cfg:                 config,
		assetToRoutineChMap: make(map[string]chan *models.Issuance),
		db:                  db,
	}
}

// / Start treat latest issuance with utmost priority
// / Start a go routine for each issuance
// / Each contract has its own go routine
func (r *Relayer) Start(ctx context.Context) error {
	for _, asset := range r.cfg.Assets {
		assetId := utils.GenerateIDForAsset(asset.InternalAssetIdentity)
		r.assetToRoutineChMap[assetId] = make(chan *models.Issuance)
		go r.startRoutine(ctx, assetId)
	}

	for c := range ctx.Done() {
		logging.Logger.Info("Relayer routine stopped", zap.Any("cause", c))
		return ctx.Err()
	}
	return nil

}

func (r *Relayer) AcceptIssuance(issuance *models.Issuance) error {
	logging.Logger.Info("Issuance accepted", zap.String("assetID", issuance.Price.AssetID))
	if _, ok := r.assetToRoutineChMap[issuance.Price.AssetID]; !ok {
		return fmt.Errorf("no routine found for assetID: %s", issuance.Price.AssetID)
	}
	// Send the issuance to the corresponding channel
	// This will be picked up by the startRoutine function
	logging.Logger.Info("Sending issuance to asset channel", zap.String("assetID", issuance.Price.AssetID))
	r.assetToRoutineChMap[issuance.Price.AssetID] <- issuance
	return nil
}

func (r *Relayer) startRoutine(ctx context.Context, assetID string) {
	ch := r.assetToRoutineChMap[assetID]
	for issuance := range ch {
		for _, ctrct := range r.cfg.Contracts {
			go r.ConveyIssuanceToContract(ctx, issuance, ctrct)
		}
	}
}

func (r *Relayer) ConveyIssuanceToContract(ctx context.Context, issuance *models.Issuance, ctrct config.ContractConfig) error {
	logging.Logger.Info("Conveying issuance to contract", zap.String("assetID", issuance.Price.AssetID), zap.String("contract", ctrct.Address))
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

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logging.Logger.Error("Failed to get nonce", zap.Error(err))
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
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
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address := common.HexToAddress(ctrct.Address)
	// Load the verifier contract
	contract, err := importVerifier.NewVerifier(address, client)
	if err != nil {
		logging.Logger.Error("Failed to load verifier contract", zap.Error(err))
		return fmt.Errorf("failed to load verifier contract: %w", err)
	}

	// Prepare inputs
	assetIndex := [][32]byte{}
	var key [32]byte
	copy(key[:], []byte(issuance.Price.AssetID))
	assetIndex = append(assetIndex, key)

	prices := []importVerifier.IIfaPriceFeedPriceFeed{
		{
			Price:          big.NewInt(int64(issuance.Price.Value)),
			Decimal:        int8(issuance.Price.Expo),
			LastUpdateTime: uint64(issuance.Price.Timestamp.Unix()),
		},
	}

	tx, err := contract.SubmitPriceFeed(auth, assetIndex, prices)
	if err != nil {
		logging.Logger.Error("Failed to submit price feed", zap.Error(err))
		return fmt.Errorf("failed to submit price feed: %w", err)
	}

	logging.Logger.Info("Submitted price feed", zap.String("tx", tx.Hash().Hex()))

	return nil
}

// abigen --bin=./build/Store.bin --abi=./build/Store.abi --pkg=store --out=Store.go
