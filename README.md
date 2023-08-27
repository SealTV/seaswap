# seaswap

This application is trying to calculate amount of TokenOut after swapping from TokenFrom. Also the app checks input PoolAddres for availability to swap.

The app checks the following conditions:

- PoolAddres is valid
- TokenFrom is valid
- TokenTo is valid
- TokenFrom is in PoolAddres
- TokenTo is in PoolAddres
- TokenFrom is not equal TokenTo

The app convert fromAmount to wei and calculate amount of TokenOut after swapping from TokenFrom.

## input params

1. poolAddres - address of uniswap_v2 pool in ethereum (ex [pull addres 0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852](https://etherscan.io/address/0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852))
2. fromToken - address of token that you want to swap (ex [WETH](https://etherscan.io/token/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2))
3. toToken - address of token that you want to get after swap (ex [USDT](https://etherscan.io/token/0xdac17f958d2ee523a2206206994597c13d831ec7))
4. fromAmount - amount of fromToken that you want to swap (ex 1)

## task

You need to write a program (in golang) that accepts following params: address of uniswap_v2 pool in ethereum (ex [pull addres 0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852](https://etherscan.io/address/0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852)), inputToken address, outputToken address, inputAmount. Program should return outputAmount that corresponding uniswap_v2 pool will return if you try to swap inputAmount of fromToken. All math calculations should be done inside your program (not calling external services for results).

## example

- PoolID: 0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852
- FromToken: 0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2
- ToToken: 0xdac17f958d2ee523a2206206994597c13d831ec7
- InputAmount: 1e18

## How to run from source

```bash
go run main.go

# 2023/08/27 23:02:24 You can swap 1 WETH -> 1650.607226 USDT
```

### set custom params

```bash
go run main.go --poolAddress 0x0357079bbeCADD7bd4B7a9f418014679Fc4e3926 --fromToken 0x111111111117dC0aa78b770fA6A738034120C302 --toToken 0xdAC17F958D2ee523a2206206994597C13D831ec7 --fromAmount 10
# 2023/08/27 23:10:26 You can swap 10 1INCH -> 1.976283 USDT
```
