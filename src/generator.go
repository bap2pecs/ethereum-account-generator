package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type PathPos struct {
	account, index uint32
}

func (pos1 PathPos) cmp(pos2 PathPos) int {
	if pos1.account < pos2.account {
		return -1
	} else if pos1.account > pos2.account {
		return 1
	}

	if pos1.index > pos2.index {
		return 1
	} else if pos1.index < pos2.index {
		return -1
	} else {
		return 0
	}
}

// warning: this function will not work properly when:
//
//	pos.account == math.MaxUint32 && pos.index+step > math.MaxUint32
func (pos PathPos) inc(step uint32) PathPos {
	if pos.index+step >= pos.index {
		return PathPos{pos.account, pos.index + step}
	}
	return PathPos{pos.account + 1, pos.index + step}
}

func generateHDAccount(wallet *hdwallet.Wallet, pos PathPos) (address, key string) {
	// see https://en.bitcoin.it/wiki/BIP_0044
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/%d'/0/%d", pos.account, pos.index))
	acc, err := wallet.Derive(path, false)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := wallet.PrivateKeyHex(acc)
	if err != nil {
		log.Fatal(err)
	}

	return acc.Address.Hex(), privateKey
}

func printAddress(pos PathPos, address, key string) {
	fmt.Printf("%d:%d: %s %s\n", pos.account, pos.index, address, key)
}

var searchPattern = os.Getenv("SEARCH_PATTERN")

func isAllLeadingZeroAddress(address string) bool {
	match, err := regexp.MatchString(searchPattern, address)
	if err != nil {
		log.Fatal(err)
	}
	return match
}

func mineHDAccount(wallet *hdwallet.Wallet, pos PathPos, step uint32, trackerSlot *PathPos) {
	for i := pos; true; i = i.inc(step) {
		address, key := generateHDAccount(wallet, i)
		if isAllLeadingZeroAddress((address)) {
			printAddress(i, address, key)
		}
		*trackerSlot = i
	}
}

func minInSlice(values []PathPos) PathPos {
	min := values[0]
	for _, v := range values {
		if min.cmp(v) > 0 {
			min = v
		}
	}
	return min
}

func registerOnExit(tracker []PathPos) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nLatest searched account and index (in each slot):")
		for _, pos := range tracker {
			fmt.Println(pos.account, pos.index)
		}
		fmt.Println("to continue: ", minInSlice(tracker).inc(1))
		os.Exit(1)
	}()
}

func parseMnemonic() string {
	return os.Getenv("MNEMONIC")
}

func parseStartAccountAndIndex() PathPos {
	startAccount, err := strconv.ParseUint(os.Getenv("START_ACCOUNT"), 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	startIndex, err := strconv.ParseUint(os.Getenv("START_INDEX"), 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return PathPos{uint32(startAccount), uint32(startIndex)}
}

func main() {
	// parse arguments from ENV variables
	mnemonic := parseMnemonic()
	startPos := parseStartAccountAndIndex()

	// schedule the number of go routines to be equal to the number of available
	// CPU cores
	thread := uint32(runtime.GOMAXPROCS(0))
	tracker := make([]PathPos, thread)

	// print some useful info on exit
	registerOnExit(tracker)

	// create the wallet from mnemonic
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	// searching for addresses staring with many zeros
	var wg sync.WaitGroup
	for i := uint32(0); i < thread; i++ {
		wg.Add(1)
		go func(i uint32) {
			defer wg.Done()
			mineHDAccount(wallet, startPos.inc(i), thread, &tracker[i])
		}(i)
	}
	wg.Wait()
}
