#!/bin/bash

cd $GOPATH/src/github.com/bongo227/Furlang/
alias help="printf 'build - Builds the compiler\nclean - Cleans up the build directory\ntest - Runs go tests\nhelp - Prints the help text\nitest - Runs intergration tests\nlint - Runs go linter\ndead - runs dead code test\nftest - Runs full test suite\n'"
alias build="go build -o build/furlang"
alias clean="rm -f build/*"
alias test="go test ./..."
alias itest="bash makeTests.sh && bats build/tests.bats"
alias lint="golint ./..."
alias dead="unused ./..."
alias ftest="test && itest && lint && dead"
clear
help