package gasprice

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

// GetGasTipFunc provides a general interface for retrieving the closest estimate for gas tip to allow for
// timely execution of a transaction.
type GetGasTipFunc = func() (*big.Int, error)

// NoopGetGasTipFunc returns nil gas tip and nil error.
func NoopGetGasTipFunc() GetGasTipFunc {
	return func() (*big.Int, error) {
		return nil, nil
	}
}

// GetGasTipWithEthClient returns a GetGasTipFunc using an eth client.
func GetGasTipWithEthClient(eth *ethclient.Client) GetGasTipFunc {
	return func() (*big.Int, error) {
		chainId, err := eth.ChainID(context.Background())

		if err != nil {
			return nil, err
		}

		if chainId.Cmp(big.NewInt(31337)) == 0 {
			// If we're on a local chain, we don't need to tip
			return big.NewInt(10000000), nil
		}

		gt, err := eth.SuggestGasTipCap(context.Background())
		if err != nil {
			return nil, err
		}
		return gt, nil
	}
}
