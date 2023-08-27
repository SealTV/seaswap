package main

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/sealtv/seaswap/abi/uniswapV2ERC20"
	"github.com/sealtv/seaswap/abi/uniswapV2Factory"
	"github.com/sealtv/seaswap/abi/uniswapV2Router02"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const ethURL = "https://mainnet.infura.io/v3/65ddf5ea8f7e4f01a00e0c7657405483"

const UniswapFactoryAddress = "0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"
const UniswapRouterAddress = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"

func init() {
	pflag.String("eth-url", ethURL, "Ethereum node URL")

	pflag.String("poolAddress", "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852", "Uniswap Pool Address")
	pflag.String("fromToken", "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", "From Token Address")
	pflag.String("toToken", "0xdac17f958d2ee523a2206206994597c13d831ec7", "To Token Address")
	pflag.Int64("fromAmount", 1, "From Token Amount")

	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
}
func main() {
	cli, err := ethclient.Dial(viper.GetString("eth-url"))
	if err != nil {
		log.Fatalf("cannot connect to Ethereum node: %v", err)
	}
	defer cli.Close()

	uniswap, err := newUniswap(cli,
		viper.GetString("poolAddress"),
		viper.GetString("fromToken"),
		viper.GetString("toToken"),
	)

	if err != nil {
		log.Fatalf("cannot create uniswap: %v", err)
	}

	isPullValid, err := uniswap.checkIsPullValid()
	if err != nil {
		log.Fatalf("cannot check is pull valid: %v", err)
	}

	if !isPullValid {
		log.Fatalf("pull is not valid")
	}

	fromTokenDecimal, err := uniswap.getFromTokenDecimals()
	if err != nil {
		log.Fatalf("cannot get FROM token decimal: %v", err)
	}

	fromExp := new(big.Int).Exp(big.NewInt(10), fromTokenDecimal, nil)

	toTokenDecimal, err := uniswap.getToTokenDecimals()
	if err != nil {
		log.Fatalf("cannot get TO token decimal: %v", err)
	}

	toExp := new(big.Int).Exp(big.NewInt(10), toTokenDecimal, nil)

	amonutsOut, err := uniswap.getAmountsOut(new(big.Int).Mul(big.NewInt(viper.GetInt64("fromAmount")), fromExp))
	if err != nil {
		log.Fatalf("cannot get amountsOut: %v", err)
	}

	if len(amonutsOut) != 2 {
		log.Fatalf("invalid amountsOut")
	}

	outAmount := new(big.Float).Quo(new(big.Float).SetInt(amonutsOut[1]), new(big.Float).SetInt(toExp))

	fromTokenSymbol, err := uniswap.getFromTokenSymbol()
	if err != nil {
		log.Fatalf("cannot get from token symbol: %v", err)
	}

	toTokenSymbol, err := uniswap.getToTokenSymbol()
	if err != nil {
		log.Fatalf("cannot get to token symbol: %v", err)
	}

	log.Printf("You can swap %v %s -> %s %s", viper.GetInt64("fromAmount"), fromTokenSymbol, outAmount.String(), toTokenSymbol)

}

type uniswap struct {
	poolAddress      common.Address
	fromTokenAddress common.Address
	toTokenAddress   common.Address

	fact   *uniswapV2Factory.UniswapV2FactoryCaller
	router *uniswapV2Router02.UniswapV2Router02Caller

	fromToken *uniswapV2ERC20.UniswapV2ERC20Caller
	toToken   *uniswapV2ERC20.UniswapV2ERC20Caller
}

func newUniswap(cli *ethclient.Client, poolAddress, fromToken, toToken string) (*uniswap, error) {
	fromTokenAddress := common.HexToAddress(fromToken)
	toTokenAddres := common.HexToAddress(toToken)

	if fromTokenAddress.Cmp(toTokenAddres) == 0 {
		return nil, errors.New("fromToken and toToken cannot be the same")
	}

	fact, err := uniswapV2Factory.NewUniswapV2FactoryCaller(common.HexToAddress(UniswapFactoryAddress), cli)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create uniswap factory")
	}

	router, err := uniswapV2Router02.NewUniswapV2Router02Caller(common.HexToAddress(UniswapRouterAddress), cli)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create uniswap router")
	}

	from, err := uniswapV2ERC20.NewUniswapV2ERC20Caller(fromTokenAddress, cli)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create uniswapv2")
	}

	to, err := uniswapV2ERC20.NewUniswapV2ERC20Caller(toTokenAddres, cli)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create uniswapv2")
	}

	return &uniswap{
		poolAddress:      common.HexToAddress(poolAddress),
		fromTokenAddress: fromTokenAddress,
		toTokenAddress:   toTokenAddres,

		fact:      fact,
		router:    router,
		fromToken: from,
		toToken:   to,
	}, nil
}

func (u *uniswap) checkIsPullValid() (bool, error) {
	pair, err := u.fact.GetPair(nil, u.fromTokenAddress, u.toTokenAddress)
	if err != nil {
		return false, errors.Wrap(err, "cannot get pair")
	}

	if pair.Cmp(u.poolAddress) != 0 {
		log.Printf("expected pair address %s, got %s", u.poolAddress.String(), pair.String())
		return false, nil
	}

	return true, nil
}

func (u *uniswap) getAmountsOut(amountIn *big.Int) ([]*big.Int, error) {
	return u.router.GetAmountsOut(nil, amountIn, []common.Address{
		u.fromTokenAddress,
		u.toTokenAddress,
	})
}

func (u *uniswap) getFromTokenDecimals() (*big.Int, error) {
	return getTokenDecimals(u.fromToken)
}

func (u *uniswap) getToTokenDecimals() (*big.Int, error) {
	return getTokenDecimals(u.toToken)
}

func (u *uniswap) getFromTokenSymbol() (string, error) {
	return getSymbol(u.fromToken)
}

func (u *uniswap) getToTokenSymbol() (string, error) {
	return getSymbol(u.toToken)
}

func getTokenDecimals(token *uniswapV2ERC20.UniswapV2ERC20Caller) (*big.Int, error) {
	decimals, err := token.Decimals(nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get decimals")
	}

	return big.NewInt(int64(decimals)), nil
}

func getSymbol(token *uniswapV2ERC20.UniswapV2ERC20Caller) (string, error) {
	symbol, err := token.Symbol(nil)
	if err != nil {
		return "", errors.Wrap(err, "cannot get symbol")
	}

	return symbol, nil
}
