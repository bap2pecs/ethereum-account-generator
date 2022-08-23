build:
	go build -o bin/generator generator.go

clean:
	rm -rf bin/

run:
	export MNEMONIC=$$(awk -F'=' '$$1 == "MNEMONIC" {print $$2 }' ../.env); \
	export START_POS=$$(awk -F'=' '$$1 == "START_POS" {print $$2 }' ../.env); \
	export SEARCH_PATTERN=$$(awk -F'=' '$$1 == "SEARCH_PATTERN" {print $$2 }' ../.env); \
	go run generator.go