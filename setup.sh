#!/bin/bash
# alias.sh

# Call source setup.sh to setup the aliases in your terminal

alias build="go build -o furlang -tags 'llvm'"
alias buildn="go build -o furlang -tags 'nollvm'"
alias compile="sudo ./furlang"
alias run="lli build/ben.ll"
alias ret="echo \$?"
alias dockerbuild="docker build -t furlang ."
alias test="docker run -v $GOPATH/src/bitbucket.com/bongo227/furlang:/go/src/bitbucket.com/bongo227/furlang furlang"