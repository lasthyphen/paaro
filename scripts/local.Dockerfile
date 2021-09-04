# syntax=docker/dockerfile:experimental

# This Dockerfile is meant to be used with the build_local_dep_image.sh script
# in order to build an image using the local version of coreth

# Changes to the minimum golang version must also be replicated in
# scripts/ansible/roles/golang_based/defaults/main.yml
# scripts/build_dijets.sh
# scripts/local.Dockerfile (here)
# Dockerfile
# README.md
# go.mod
FROM golang:1.15.5-buster

RUN mkdir -p /go/src/github.com/djt-labs

WORKDIR $GOPATH/src/github.com/djt-labs
COPY paaro paaro
COPY coreth coreth

WORKDIR $GOPATH/src/github.com/djt-labs/paaro
RUN ./scripts/build_dijets.sh
RUN ./scripts/build_coreth.sh ../coreth $PWD/build/plugins/evm

RUN ln -sv $GOPATH/src/github.com/djt-labs/dijets-byzantine/ /paaro
