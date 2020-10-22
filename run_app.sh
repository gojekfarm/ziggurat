#!/usr/bin/env bash

go build -race
./ziggurat-go --config="./config/config.sample.yaml"
