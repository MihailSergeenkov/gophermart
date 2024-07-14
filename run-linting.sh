#! /bin/bash

if [ -z ${GOLANGCI_LINT_CONFIG_FILE+x} ]; then
    docker run --rm \
        -v $(pwd):/app \
        --workdir /app \
        golangci/golangci-lint \
            bash -c "golangci-lint run -v"
else
    docker run --rm \
        -v $(pwd):/app \
        --workdir /app \
        golangci/golangci-lint \
            bash -c "golangci-lint -c=$GOLANGCI_LINT_CONFIG_FILE run" > lint-report.ugly.json
    cat lint-report.ugly.json | jq > lint-report.json
    rm -f lint-report.ugly.json
fi
