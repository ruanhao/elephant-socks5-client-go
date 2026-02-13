#!/usr/bin/env bash
# -*- coding: utf-8 -*-
#
# Description:

set -e

while getopts "vruc" opt; do
    case $opt in
        v)                     # verbose
            set -x
            verbose=1
            ;;
        r)                     # release
            type=Release
            ;;
        u)                     # unittest
            ut=1
            ;;
        c)                     # just compile
            just_compile=1
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            exit 1
    esac
done

shift $((OPTIND-1))

if [ -n "$ut" ]; then
    go test -v ./...
    exit 0
fi

BUILD_OPTIONS=("--race") # enable race detection
if [ -n "$verbose" ]; then
    BUILD_OPTIONS+=("-v")
fi
if [ "${type}" == "Release" ]; then
    BUILD_OPTIONS+=("-ldflags=-s -w")
fi


if [ -z "$just_compile" ]; then
    go run ./...
else
    go build "${BUILD_OPTIONS[@]}" -o elephant ./...
fi
