package abis

import _ "embed"

//go:embed events_abi_uniswap_v3.json
var EventsABIUniswapV3String string

//go:embed events_abi_pancakeswap_v3.json
var EventsABIPancakeswapV3String string

//go:embed events_abi_sushiswap_v3.json
var EventsABISushiswapV3String string

//go:embed events_abi_uniswap_v2.json
var EventsABIUniswapV2String string

const ERC20MetadataABIString = `[{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"}]`
