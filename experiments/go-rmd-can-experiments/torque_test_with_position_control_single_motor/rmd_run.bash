#!/bin/bash

if [ $# -ne 1 ]; then
    echo "Usage: $0 <float_value>"
    exit 1
fi

FLOAT_VALUE=$1

go run main.go 1 $FLOAT_VALUE
go run main.go 2 $FLOAT_VALUE
go run main.go 3 $FLOAT_VALUE
go run main.go 4 $FLOAT_VALUE
