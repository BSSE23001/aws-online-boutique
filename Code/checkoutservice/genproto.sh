#!/bin/bash -eu
# [START gke_checkoutservice_genproto]

PATH=$PATH:$GOPATH/bin
protodir=../../protos

protoc --go_out=plugins=grpc:genproto -I $protodir $protodir/demo.proto

# [END gke_checkoutservice_genproto]