#!/bin/bash -eu
# [START gke_frontend_genproto]

PATH=$PATH:$GOPATH/bin
protodir=../../protos

protoc --go_out=plugins=grpc:genproto -I $protodir $protodir/demo.proto

# [END gke_frontend_genproto]