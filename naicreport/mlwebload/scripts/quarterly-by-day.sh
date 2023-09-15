#!/bin/bash
#
# This is meant to be run in a directory that has everything, symlinked if necessary

mkdir -p ./output
./naicreport ml-webload -tag quarterly -sonalyze ./sonalyze -config-file ./ml-nodes.json -output-path ./output -data-path ./data -from 90d -daily
