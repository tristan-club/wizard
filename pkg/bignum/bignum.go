package bignum

import (
	"fmt"
	"github.com/trustwallet/go-primitives/numbers"
	"math/big"
	"strings"
)

func AddDecimal(i *big.Int, d int, p int64) {
	if p == 0 {
		p = 10
	}
	var decimals, pow = big.NewInt(int64(d)), big.NewInt(p)
	pow.Exp(pow, decimals, nil)
	i.Mul(i, pow)
}

func FloatAddDecimal(f *big.Float, d int, p int64) *big.Int {
	if p == 0 {
		p = 10
	}
	var pow, dec = big.NewInt(p), big.NewInt(int64(d))
	pow.Exp(pow, dec, nil)
	f.Mul(f, new(big.Float).SetInt(pow))
	i := new(big.Int)
	i, _ = f.Int(i)
	return i
}

func CutDecimal(i *big.Int, d int, p int) string {

	var decimals, pow = big.NewInt(int64(d)), big.NewInt(10)
	pow.Exp(pow, decimals, nil)
	bigF := new(big.Float).SetInt(i)
	bigF.Quo(bigF, new(big.Float).SetInt(pow))
	//log.Debug().Msgf("got float amount %s", bigF.String())
	//log.Debug().Msgf("got handled float amount %s", bigF.Text('f', p))
	return bigF.Text('f', p)
	// todo 更好地格式化余额方式
	if i.BitLen() > 28 {
		return bigF.Text('f', p)
	} else {
		f, _ := new(big.Float).SetString(bigF.Text('f', p))
		return f.String()
	}

}
func HandleDecimal(amount *big.Int, decimals int) string {
	balanceCutDecimal := CutDecimal(amount, decimals, 4)
	return balanceCutDecimal
}

func HandleDecimalStr(amount string, decimals int) (string, error) {
	//return numbers.DecimalExp(amount, -18)
	b, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		//log.Error().Msgf("decimal str handle invalid balance,chain_info type %d, chain_info id %d address %s, amount %s", amount)
		return "0", fmt.Errorf("invalid amount %s", amount)
	}
	resp := CutDecimal(b, decimals, 4)
	return resp, nil
}

func HandleAddDecimal(value string, decimal int) (*big.Int, bool) {
	//if _, ok := new(big.Float).SetString(value); !ok {
	//	return nil, false
	//}
	numberExped := numbers.DecimalExp(value, decimal)
	if strings.Contains(numberExped, ".") {
		numFloat, ok := new(big.Float).SetString(numberExped)
		if !ok {
			return nil, false
		}
		bigNum, _ := numFloat.Int(nil)
		bigNum.Add(bigNum, new(big.Int).SetUint64(1))
		return bigNum, true
	} else {
		return new(big.Int).SetString(numbers.DecimalExp(value, decimal), 10)
	}

}

func DecimalExp(dec string, exp int) string {
	return numbers.DecimalExp(dec, exp)
}
