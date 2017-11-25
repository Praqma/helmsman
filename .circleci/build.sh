#!/bin/bash -x
echo "running tests ..."
go test -v

if [ $? -eq 0 ]; then
    echo "building ..."
    go build

    if [ $? -eq 0 ]; then
        echo "cleaning after tests ..."
        rm hello.world hello.world1 helmsman
        git clean -fd
        echo "releasing ..."
        goreleaser --release-notes release_notes.md | tee /dev/tty | grep -o "error"
        if [ $? -eq 0 ]; then
            echo "goreleaser experienced an error and no new releases made. That is Ok!"
        else
            echo "New release was successfully made."    
        fi
    else
        echo "Build failed!!"
        exit 9   
    fi 
else
    echo "tests failed ... Aborting!"
    exit 9
fi