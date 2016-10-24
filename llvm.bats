#!/usr/bin/env bats

# Create a build directory if one doesnt exsist
if [ ! -d "build" ]; then
    mkdir build
fi

# Run each test in the tests folder
TEST_DIR="tests/*"
for f in $TEST_DIR
do
  # Get the name of the file without folder or extension
  name=$(basename $f)
  name=${name%.*}

  # Test the file
  @test "test: $name" {
    run ./furlang $f
    echo "$output"
    [ "$status" -eq 0 ]

    run lli build/ben.ll
    echo "$output"
    [ "$status" -eq 123 ]
  }
done
