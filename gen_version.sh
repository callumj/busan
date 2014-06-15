#!/bin/bash
MY_PATH="`dirname \"$0\"`"

cur_version=`cat ${MY_PATH}/VERSION`

s=$'Hello\nWorld'
go_src=$'package utils\nconst BusanVersion string = "'"${cur_version}"$'"'

echo "$go_src"

echo "$go_src" > ${MY_PATH}/utils/version.go