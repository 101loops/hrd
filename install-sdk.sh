#!/bin/bash

UPDATE_URL="https://appengine.google.com/api/updatecheck"
VERSION=$(echo $(curl ${UPDATE_URL}) | sed -E 's/release: \"(.+)\"(.*)/\1/g')

ARCH="386"
if [[ `uname -a` == *x86_64* ]]
then
    ARCH="amd64"
fi

FILE=go_appengine_sdk_linux_$ARCH-$VERSION.zip
echo "downloading '$FILE'"

wget https://commondatastorage.googleapis.com/appengine-sdks/featured/$FILE -nv
unzip -q $FILE -d .