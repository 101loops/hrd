#!/bin/bash

ARCH="386"
VERSION="1.9.3"

if [[ `uname -a` == *x86_64* ]]
then
    ARCH="amd64"
fi

FILE=go_appengine_sdk_linux_$ARCH-$VERSION.zip
echo "downloading '$FILE'"

wget https://commondatastorage.googleapis.com/appengine-sdks/featured/$FILE -nv
unzip -q $FILE -d .