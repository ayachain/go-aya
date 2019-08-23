#!/usr/bin/env bash

protoc --go_out=. ./assets/assets.proto

protoc --go_out=. ./block/block.proto

protoc --go_out=. ./chaininfo/chaininfo.proto

protoc --go_out=. ./electoral/electoral.proto

protoc --go_out=. ./indexes/index.proto

protoc --go_out=. ./minined/minined.proto

protoc --go_out=. ./node/node.proto

protoc --go_out=. ./receipt/receipt.proto

protoc --go_out=. ./transaction/transaction.proto
