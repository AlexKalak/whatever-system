package service

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const mempoolCallDataABIJSON = `[
	{"type":"function","name":"swapExactTokensForTokens","inputs":[{"name":"amountIn","type":"uint256"},{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"swapTokensForExactTokens","inputs":[{"name":"amountOut","type":"uint256"},{"name":"amountInMax","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"swapExactETHForTokens","inputs":[{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"swapTokensForExactETH","inputs":[{"name":"amountOut","type":"uint256"},{"name":"amountInMax","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"swapExactTokensForETH","inputs":[{"name":"amountIn","type":"uint256"},{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"swapETHForExactTokens","inputs":[{"name":"amountOut","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","inputs":[{"name":"amountIn","type":"uint256"},{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"swapExactETHForTokensSupportingFeeOnTransferTokens","inputs":[{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"swapExactTokensForETHSupportingFeeOnTransferTokens","inputs":[{"name":"amountIn","type":"uint256"},{"name":"amountOutMin","type":"uint256"},{"name":"path","type":"address[]"},{"name":"to","type":"address"},{"name":"deadline","type":"uint256"}],"outputs":[]},
	{"type":"function","name":"exactInputSingle","inputs":[{"name":"params","type":"tuple","components":[{"name":"tokenIn","type":"address"},{"name":"tokenOut","type":"address"},{"name":"fee","type":"uint24"},{"name":"recipient","type":"address"},{"name":"deadline","type":"uint256"},{"name":"amountIn","type":"uint256"},{"name":"amountOutMinimum","type":"uint256"},{"name":"sqrtPriceLimitX96","type":"uint160"}]}],"outputs":[]},
	{"type":"function","name":"exactOutputSingle","inputs":[{"name":"params","type":"tuple","components":[{"name":"tokenIn","type":"address"},{"name":"tokenOut","type":"address"},{"name":"fee","type":"uint24"},{"name":"recipient","type":"address"},{"name":"deadline","type":"uint256"},{"name":"amountOut","type":"uint256"},{"name":"amountInMaximum","type":"uint256"},{"name":"sqrtPriceLimitX96","type":"uint160"}]}],"outputs":[]},
	{"type":"function","name":"exactInput","inputs":[{"name":"params","type":"tuple","components":[{"name":"path","type":"bytes"},{"name":"recipient","type":"address"},{"name":"deadline","type":"uint256"},{"name":"amountIn","type":"uint256"},{"name":"amountOutMinimum","type":"uint256"}]}],"outputs":[]},
	{"type":"function","name":"exactOutput","inputs":[{"name":"params","type":"tuple","components":[{"name":"path","type":"bytes"},{"name":"recipient","type":"address"},{"name":"deadline","type":"uint256"},{"name":"amountOut","type":"uint256"},{"name":"amountInMaximum","type":"uint256"}]}],"outputs":[]},
	{"type":"function","name":"multicall","inputs":[{"name":"data","type":"bytes[]"}],"outputs":[]},
	{"type":"function","name":"swap","inputs":[{"name":"amount0Out","type":"uint256"},{"name":"amount1Out","type":"uint256"},{"name":"to","type":"address"},{"name":"data","type":"bytes"}],"outputs":[]},
	{"type":"function","name":"swap","inputs":[{"name":"recipient","type":"address"},{"name":"zeroForOne","type":"bool"},{"name":"amountSpecified","type":"int256"},{"name":"sqrtPriceLimitX96","type":"uint160"},{"name":"data","type":"bytes"}],"outputs":[]},
	{"type":"function","name":"mint","inputs":[{"name":"recipient","type":"address"},{"name":"tickLower","type":"int24"},{"name":"tickUpper","type":"int24"},{"name":"amount","type":"uint128"},{"name":"data","type":"bytes"}],"outputs":[]},
	{"type":"function","name":"burn","inputs":[{"name":"tickLower","type":"int24"},{"name":"tickUpper","type":"int24"},{"name":"amount","type":"uint128"}],"outputs":[]}
]`

type chainMempoolCallDataParser struct {
	contractABI abi.ABI
}

func newChainMempoolCallDataParser() (chainMempoolCallDataParser, error) {
	contractABI, err := abi.JSON(strings.NewReader(mempoolCallDataABIJSON))
	if err != nil {
		return chainMempoolCallDataParser{}, err
	}
	return chainMempoolCallDataParser{contractABI: contractABI}, nil
}

func (p *chainMempoolCallDataParser) parse(data []byte) (ChainMempoolCallData, bool, error) {
	if len(data) < 4 {
		return ChainMempoolCallData{}, false, nil
	}

	method, err := p.contractABI.MethodById(data[:4])
	if err != nil {
		return ChainMempoolCallData{}, false, nil
	}

	fmt.Println("Found by id: ", method)

	args := make(map[string]any)
	if err := method.Inputs.UnpackIntoMap(args, data[4:]); err != nil {
		return ChainMempoolCallData{}, false, err
	}

	return ChainMempoolCallData{
		Method:    method.RawName,
		Signature: method.Sig,
		Selector:  "0x" + hex.EncodeToString(data[:4]),
		Args:      normalizeCallDataValue(args).(map[string]any),
	}, true, nil
}

func normalizeCallDataValue(value any) any {
	switch v := value.(type) {
	case common.Address:
		return v.String()
	case *big.Int:
		if v == nil {
			return "0"
		}
		return v.String()
	case []byte:
		return "0x" + hex.EncodeToString(v)
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[key] = normalizeCallDataValue(val)
		}
		return out
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		out := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out[i] = normalizeCallDataValue(rv.Index(i).Interface())
		}
		return out
	case reflect.Struct:
		out := make(map[string]any, rv.NumField())
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			if rt.Field(i).PkgPath != "" {
				continue
			}
			out[rt.Field(i).Name] = normalizeCallDataValue(rv.Field(i).Interface())
		}
		return out
	case reflect.Ptr:
		if rv.IsNil() {
			return nil
		}
		return normalizeCallDataValue(rv.Elem().Interface())
	}

	return fmt.Sprint(value)
}
