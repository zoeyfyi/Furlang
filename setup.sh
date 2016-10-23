#!/bin/bash
# alias.sh

# Call source setup.sh to setup the aliases in your terminal

alias build="go build -o furlang"
alias clean="rm build/* && rm furlang"
alias compile="./furlang"
alias run="lli build/ben.ll"
alias ret="echo \$?"
alias dockerbuild="docker build -t furlang ."
alias testg="go test -tags 'nollvm' ./..."
alias test="docker run -v $GOPATH/src/bitbucket.com/bongo227/furlang:/go/src/bitbucket.com/bongo227/furlang furlang"