package settings

import "fmt"

type CacheKeyPrefix string

func (c CacheKeyPrefix) Apply(key string) string {
	return fmt.Sprintf("%s:%s", string(c), key)
}

func (c CacheKeyPrefix) ApplyWithSuffix(key, suffix string) string {
	return fmt.Sprintf("%s:%s:%s", string(c), key, suffix)
}

const MARKET_DATA_CACHE = CacheKeyPrefix("MDC")
