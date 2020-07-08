package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// MockServices is a that tolds different mock services to be passed
// around easily for testing setup

// UintToPaddedString converts an int to string of length 19 by padding with 0
func UintToPaddedString(num int64) string {
	return fmt.Sprintf("%019d", num)
}

// GetTickChannelID is used to get the channel id for OHLCV data streaming
// it takes pairname, duration and units of data streaming
func GetTickChannelID(bt, qt string, unit string, duration int64) string {
	pair := GetPairKey(bt, qt)
	return fmt.Sprintf("%s::%d::%s", pair, duration, unit)
}

// GetPairKey return the pair key identifier corresponding to two
func GetPairKey(bt, qt string) string {
	return strings.ToLower(fmt.Sprintf("%s::%s", bt, qt))
}

func GetTradeChannelID(bt, qt string) string {
	return strings.ToLower(fmt.Sprintf("%s::%s", bt, qt))
}

func GetOHLCVChannelID(bt, qt string, unit string, duration int64) string {
	pair := GetPairKey(bt, qt)
	return fmt.Sprintf("%s::%d::%s", pair, duration, unit)
}

func GetOrderBookChannelID(bt, qt string) string {
	return strings.ToLower(fmt.Sprintf("%s::%s", bt, qt))
}

func Retry(retries int, fn func() error) error {
	if err := fn(); err != nil {
		retries--
		if retries <= 0 {
			return err
		}

		// preventing thundering herd problem (https://en.wikipedia.org/wiki/Thundering_herd_problem)
		time.Sleep(time.Second)

		return Retry(retries, fn)
	}

	return nil
}

func PrintJSON(x interface{}) {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Print(string(b), "\n")
}

func JSON(x interface{}) string {
	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	return fmt.Sprint(string(b), "\n")
}

func MaxIntMap(m map[string]int) (string, int) {
	var maxKey string = ""
	var maxValue int = 0

	for key, val := range m {
		if val > maxValue {
			maxValue = val
			maxKey = key
		}
	}

	return maxKey, maxValue
}

func PrintError(msg string, err error) {
	log.Printf("\n%v: %v\n", msg, err)
}

// Util function to handle unused variables while testing
func Use(...interface{}) {

}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
