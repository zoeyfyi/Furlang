#!/usr/bin/env bats

if [ ! -d "build" ]; then
    mkdir build
fi

go build -tags='llvm' -o=furlang compiler.go

@test "main example" {
    run ./furlang examples/main.fur
    echo "$output"
    [ "$status" -eq 0 ]

    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
}

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

@test "blocks example" {
    run ./furlang examples/blocks.fur
    echo "$output"
    [ "$status" -eq 0 ]

    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
}

@test "if example" {
    run ./furlang examples/if.fur
    echo "$output"
    [ "$status" -eq 0 ]

    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
}