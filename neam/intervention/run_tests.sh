#!/bin/bash
cd "$(dirname "$0")"
go test -v 2>&1 | head -200
