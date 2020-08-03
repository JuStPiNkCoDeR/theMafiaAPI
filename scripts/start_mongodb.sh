#!/usr/bin/env bash

STAGE="development"

for i in "$@"
do
case ${i} in
    -s=*|--stage=*)
    STAGE="${i#*=}"
    shift
    ;;
esac
done

echo "Staring mongodb for stage - $STAGE..."