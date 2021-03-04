#!/bin/bash

rm -f ./release/*

GOOS=linux   GOARCH=amd64 go build -o ./build/release/hunched-dog__linux_amd64   ./
GOOS=linux   GOARCH=386   go build -o ./build/release/hunched-dog__linux_386     ./
GOOS=darwin  GOARCH=amd64 go build -o ./build/release/hunched-dog__darwin_amd64  ./
GOOS=windows GOARCH=amd64 go build -o ./build/release/hunched-dog__windows_amd64 ./
GOOS=windows GOARCH=386   go build -o ./build/release/hunched-dog__windows_386   ./
