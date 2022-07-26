package bignum

import (
	"fmt"
	"math/big"
	"testing"
)

func TestAdd(t *testing.T) {
	i := big.NewInt(100000000000)
	d := 18
	AddDecimal(i, d, 10)
	fmt.Println(i)
	fmt.Println(i.String())
}

func TestCut(t *testing.T) {

	//var exp = `/?=[1-9-.]/g`
	//var exp = `/[1-9]/g`
	i := big.NewInt(101000000000000001)

	f := CutDecimal(i, 18, 4)
	t.Log(f)
	z, _ := new(big.Float).SetString(f)
	t.Log(z.String())
}

func TestFloatAdd(t *testing.T) {
	f, _ := new(big.Float).SetString("23123.222")
	i := FloatAddDecimal(f, 18, 10)
	t.Log(i)
}

func TestBigIntCut(t *testing.T) {
	b, _ := new(big.Float).SetString("12345678901234567890")

	b.SetPrec(18)
	s := fmt.Sprintf("%.4f", b)
	//fmt.Println(b.Text(10))
	//s := fmt.Sprintf("%10.4f", b)

	fmt.Println(s)
}
