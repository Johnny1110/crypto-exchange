package legacy

import (
	"fmt"
	"testing"
	"time"
)

func Test_GetTimeBucket(t *testing.T) {
	now := time.Now()
	openTime := now.Truncate(24 * time.Hour).Unix()
	fmt.Println(openTime)
}

func Test_GetNextTimeBucket(t *testing.T) {
	now := time.Now()
	openTime := now.Truncate(1 * time.Hour).Unix()
	fmt.Println("now:", openTime)
	next := now.Add(1 * time.Hour)
	nextTime := next.Truncate(1 * time.Hour).Unix()
	fmt.Println("next:", nextTime)
}
