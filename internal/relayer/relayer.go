package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"oracle_engine/internal/config"
	"oracle_engine/internal/models"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// / Relayer is a service that takes issuances requests
// / and sends to the contract
// / It also updates the status of the issuance request
// / in local db (also the asset instance)
type Relayer struct {
	cfg                 *config.Config
	assetToRoutineChMap map[string]chan *models.Issuance
}

func New(config *config.Config) *Relayer {
	return &Relayer{
		cfg:                 config,
		assetToRoutineChMap: make(map[string]chan *models.Issuance),
	}
}

// / Start treat latest issuance with utmost priority
// / Start a go routine for each issuance
// / Each contract has its own go routine
func (r *Relayer) Start(ctx context.Context, issuanceCh chan *models.Issuance) error {
	for _, asset := range r.cfg.Assets {
		r.assetToRoutineChMap[asset.InternalAssetIdentity] = make(chan *models.Issuance)
		go r.startRoutine(ctx, asset.InternalAssetIdentity)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case issuance := <-issuanceCh:
			r.assetToRoutineChMap[issuance.Price.AssetID] <- issuance
		}
	}
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
	client, err := ethclient.Dial(ctrct.RPC)
	if err != nil {
		return err
	}
	privateKey, err := crypto.HexToECDSA(r.cfg.PrivateKey)
	if err != nil {
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	chainID, err := strconv.ParseInt(ctrct.ChainID, 10, 64)
	if err != nil {
		return err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		return err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address := common.HexToAddress(ctrct.Address)
	// instance, err := store.NewStore(address, client)
	// if err != nil {
	//   log.Fatal(err)
	// }

	// just set up asset key as bytes here
	var assetID [32]byte
	copy(assetID[:], []byte(issuance.Price.AssetID))

	return nil
}

// abigen --bin=./build/Store.bin --abi=./build/Store.abi --pkg=store --out=Store.go
