//go:build ignore
// +build ignore

package main

import (
"fmt"
"math/big"
"github.com/ethereum/go-ethereum/accounts/abi"
)

type IIfaPriceFeedPriceFeed struct {
	Price          *big.Int
	Decimal        int8
	LastUpdateTime uint64
}

type IfaPriceFeedVerifierSumissionData struct {
	Assesetindex [][32]byte
	Price        []IIfaPriceFeedPriceFeed
}

func main() {
	tupleType, err := abi.NewType(
"tuple", "",
[]abi.ArgumentMarshaling{
{Name: "assesetindex", Type: "bytes32[]"},
{Name: "price", Type: "(int256,int8,uint64)[]"},
},
)
	if err != nil {
		panic(err)
	}
	args := abi.Arguments{
		{Name: "ifaPriceFeedVerifierSumissionData", Type: tupleType},
	}

	in := IfaPriceFeedVerifierSumissionData{
		Assesetindex: [][32]byte{},
		Price: []IIfaPriceFeedPriceFeed{
			{Price: big.NewInt(100), Decimal: 18, LastUpdateTime: 1000},
		},
	}

	_, err = args.Pack(in)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Success")
	}
}
