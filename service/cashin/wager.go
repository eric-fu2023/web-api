package cashin

type WagerCalculator func(amount int64) int64

func DefaultWager(amount int64) int64 {
	return amount
}

func DoubleWager(amount int64) int64 {
	return amount*2
}

func GetWagerFromAmount(amount int64,fn WagerCalculator) int64 {
	return fn(amount)
}