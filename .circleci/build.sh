#!/bin/bash -x
echo "running tests ..."
go test 

if [ $? -eq 0 ]; then
    echo "releasing ..."
    goreleaser | grep -o "error"
    if [ $? -eq 0 ]; then
        goreleaser | grep -0 "already_exists"
        if [ $? -gt 0 ]; then
            echo "ERROR: failed to release. Terminating."
            Exit 9
        else
            echo "No new releases made. That is Ok!"
        fi    
    else
        echo "New release was successfully made."    
    fi
else
    echo "tests failed ... Aborting!"
    Exit 9
fi