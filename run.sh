#!/bin/sh
go run main.go keygen -f $1 -u username -p password
go run main.go client -sync -f $1

x=1
while [ $x -le 1000 ]
do
    go run main.go client -b -t 100,200,300 -f $1
done