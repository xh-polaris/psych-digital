#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)
BinaryName=psych-digital
echo "$CURDIR/bin/${BinaryName}"
exec $CURDIR/bin/${BinaryName}