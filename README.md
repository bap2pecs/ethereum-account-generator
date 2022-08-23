## Ethereum Account Generator
Generate ethereum accounts that matches the specified regex pattern.

[![License](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/bap2pecs/ethereum-account-generator/main/LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/bap2pecs/ethereum-account-generator/pulls)

## How to build the image

- prerequisite: 
  - run `docker buildx create --use`
- run `docker buildx build --platform linux/amd64,linux/arm64 -t bap2pecs/ethereum-account-generator . ` within this repo to build images for different architecture
  - add option `--push` to push images into registry
- to push to docker hub (private repo):
  - `docker login`
  - `docker push bap2pecs/monitor`

## How to use the image
- first create an `.env` file by copying the content from `.env.example` and replacing the values
- to do a quick test, run `docker run -it --rm --env-file=.env bap2pecs/ethereum-account-generator` 
- using `docker-compose`:
  - download `docker-compose.yml` to the same directory where you created the `.env` file
  - run `docker-compose up -d` to start searching
  - run `docker-compose logs -f` to see the search result
  - run `docker-compose stop` to stop searching
    - then run `docker-compose logs | grep "to continue"` to find out where the search is stopped at
    - then replace the value of `START_POS` in `.env` with the grepped number
    - then restart searching with `docker-compose up -d`

## How to test from src file
- first create an `.env` file by copying the content from `.env.example` and replacing the values
- run `cd src && make run`

## Contribution
Thank you for considering contributing to the repo! Anyone is welcomed to submit pull request, even for the smallest fixes or typos.

If you'd like to contribute, please create a new branch and submit a pull request for review.

## License
[MIT](https://raw.githubusercontent.com/bap2pecs/ethereum-account-generator/main/LICENSE)