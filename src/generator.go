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

func generateHDAccount(wallet *hdwallet.Wallet, index uint32) (address, key string) {
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", index))
	account, err := wallet.Derive(path, false)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := wallet.PrivateKeyHex(account)
	if err != nil {
		log.Fatal(err)
	}

	return account.Address.Hex(), privateKey
}

func printAccount(index uint32, address, key string) {
	fmt.Printf("%d: %s %s\n", index, address, key)
}

var searchPattern = os.Getenv("SEARCH_PATTERN")

func isAllLeadingZeroAddress(address string) bool {
	match, err := regexp.MatchString(searchPattern, address)
	if err != nil {
		log.Fatal(err)
	}
	return match
}

func mineHDAccount(wallet *hdwallet.Wallet, startIndex uint32, step int, trackerSlot *uint32) {
	for i := startIndex; true; i += uint32(step) {
		address, key := generateHDAccount(wallet, i)
		if isAllLeadingZeroAddress((address)) {
			printAccount(i, address, key)
		}
		*trackerSlot = i
	}
}

func minInSlice(values []uint32) uint32 {
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func registerOnExit(tracker []uint32) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nLatest searched index (in each slot):")
		for _, slot := range tracker {
			fmt.Println(slot)
		}
		fmt.Println("to continue: ", minInSlice(tracker)+1)
		os.Exit(1)
	}()
}

func parseMnemonic() string {
	return os.Getenv("MNEMONIC")
}

func parseStartIndex() uint32 {
	startIndex, err := strconv.ParseUint(os.Getenv("START_INDEX"), 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return uint32(startIndex)
}

func main() {
	// parse arguments from ENV variables
	mnemonic := parseMnemonic()
	startIndex := parseStartIndex()

	// schedule the number of go routines to be equal to the number of available
	// CPU cores
	thread := runtime.GOMAXPROCS(0)
	tracker := make([]uint32, thread)

	// print some useful info on exit
	registerOnExit(tracker)

	// create the wallet from mnemonic
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	// searching for addresses staring with many zeros
	var wg sync.WaitGroup
	for i := startIndex; i < startIndex+uint32(thread); i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			mineHDAccount(wallet, i, thread, &tracker[i%uint32(thread)])
		}()
	}
	wg.Wait()
}
