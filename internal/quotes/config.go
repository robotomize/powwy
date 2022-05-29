package quotes

import "time"

type Config struct {
	HashCashDifficult       int           `env:"HASH_CASH_DIFFICULT,default=5"`
	HashCashExpiredDuration time.Duration `env:"HASH_CASH_EXPIRED_DURATION,default=10m"`
}
