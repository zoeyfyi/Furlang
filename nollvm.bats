#!/usr/bin/env bats

if [ ! -d "build" ]; then
    mkdir build
fi

go get github.com/oleiade/lane
go get github.com/bongo227/cmap
go get github.com/fatih/color
go get github.com/bongo227/cmap
go build -tags='nollvm' -o=furlang compiler.go

@test "main example" {
    run ./furlang examples/main.fur
    echo "$output"
    [ "$status" -eq 0 ]
}

@test "function example" {
    run ./furlang examples/function.fur
    echo "$output"
    [ "$status" -eq 0 ]
}

@test "float example" {
    run ./furlang examples/float.fur
    echo "$output"
    [ "$status" -eq 0 ]
}

@test "returns example" {
    run ./furlang examples/returns.fur
    echo "$output"
    [ "$status" -eq 0 ]
}

@test "rpn example" {
    run ./furlang examples/rpn.fur
    echo "$output"
    [ "$status" -eq 0 ]
}

@test "blocks example" {
    run ./furlang examples/blocks.fur
    echo "$output"
    [ "$status" -eq 0 ]
}

@test "if example" {
    run ./furlang examples/if.fur
    echo "$output"
    [ "$status" -eq 0 ]
}