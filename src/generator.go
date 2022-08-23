package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"syscall"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func generateHDAccount(wallet *hdwallet.Wallet, pos int) (address, key string) {
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", pos))
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

func printAccount(pos int, address, key string) {
	fmt.Printf("%d: %s %s\n", pos, address, key)
}

var searchPattern = os.Getenv("SEARCH_PATTERN")

func isAllLeadingZeroAddress(address string) bool {
	match, err := regexp.MatchString(searchPattern, address)
	if err != nil {
		log.Fatal(err)
	}
	return match
}

func mineHDAccount(wallet *hdwallet.Wallet, startPos int, step int, trackerSlot *int) {
	for i := startPos; true; i += step {
		address, key := generateHDAccount(wallet, i)
		if isAllLeadingZeroAddress((address)) {
			printAccount(i, address, key)
		}
		*trackerSlot = i
	}
}

func minInSlice(values []int) int {
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func registerOnExit(tracker []int) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nLatest searched positions (in each slot):")
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

func parseStartPos() int {
	startPos, err := strconv.Atoi(os.Getenv("START_POS"))
	if err != nil {
		log.Fatal(err)
	}
	return startPos
}

func main() {
	// parse arguments from ENV variables
	mnemonic := parseMnemonic()
	startPos := parseStartPos()

	// schedule the number of go routines to be equal to the number of available
	// CPU cores
	thread := runtime.GOMAXPROCS(0)
	tracker := make([]int, thread)

	// print some useful info on exit
	registerOnExit(tracker)

	// create the wallet from mnemonic
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	// searching for addresses staring with many zeros
	for i := startPos; i < startPos+thread; i++ {
		go mineHDAccount(wallet, i, thread, &tracker[i%thread])
	}
	var input string
	fmt.Scanln(&input)
}
