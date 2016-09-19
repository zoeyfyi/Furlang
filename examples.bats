#!/usr/bin/env bats

if [ ! -d "build" ]; then
    mkdir build
fi

go build -tags='llvm' -o=furlang compiler.go

@test "function example" {
    run ./furlang examples/function.fur
    echo "$output"
    [ "$status" -eq 0 ]
    
    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
}

@test "float example" {
    run ./furlang examples/float.fur
    echo "$output"
    [ "$status" -eq 0 ]

    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
}

@test "main example" {
    run ./furlang examples/main.fur
    echo "$output"
    [ "$status" -eq 0 ]

    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
}

@test "returns example" {
    run ./furlang examples/returns.fur
    echo "$output"
    [ "$status" -eq 0 ]

    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
}

@test "rpn example" {
    run ./furlang examples/rpn.fur
    echo "$output"
    [ "$status" -eq 0 ]

    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
}