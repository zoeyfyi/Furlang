#!/bin/bash

# Create a build directory if one doesnt exsist
if [ ! -d "build" ]; then
    mkdir build
fi

# Clean build directory
rm build/* 2> /dev/null

# Build a bats testing file

# Add bats enviroment
echo "#!/usr/bin/env bats
" > build/tests.bats

# Add each test
TEST_DIR="tests/*"
for f in $TEST_DIR
do
  # Get the name of the file without folder or extension
  name=$(basename $f)
  name=${name%.*}

  # Create a test for the file
  echo "@test \"test: $name\" {
    run ./furlang $f
    echo \"\$output\"
    [ \"\$status\" -eq 0 ]

    run lli build/ben.ll
    echo \"\$output\"
    [ \"\$status\" -eq 123 ]
  }
  " >> build/tests.bats
done

# Run the tests
if [[ $* == *--tap* ]]; then
  bats --tap build/tests.bats
else
  bats build/tests.bats
fi


# Clean up build directory
rm build/*
