#!/bin/bash

helpInputs()
{
    echo ""
        echo "Usage: $0 [-i idoEmcoBranch] [-e emcoBranch]"
        exit 1 # Exit script after printing help
}

cloneTestRepos() {
    
        export EMCODIR=$(mktemp -d)
        # Begin script in case all parameters are correct
        echo "Testing IDO-EMCO branch $IDOEMCOBRANCH"
        echo "Testing EMCO Branch $EMCOBRANCH"
        echo "Checking out both branches first"
        EMCO_ROOT=$EMCODIR
        EMCO_PROJ=IDO-EMCO
        
        cd $EMCODIR
        git clone --recurse-submodules -b $IDOEMCOBRANCH https://github.com/otcshare/$EMCO_PROJ && cd $EMCO_PROJ
        git submodule foreach git checkout $EMCOBRANCH
        (make deploy && make test) >> $EMCODIR/test_report 2>&1
        if [ "$?" -eq "0" ]; then
          echo "Test is SUCCESSFUL"
          /bin/rm -rf $EMCODIR
        else
          echo "Test FAILED. Look inside $EMCODIR"
          exit 1
        fi
}

export EMCOBRANCH=$(git rev-parse --abbrev-ref HEAD)
export IDOEMCOBRANCH=main

while getopts "i:e:" opt
do
   case "$opt" in
      i ) IDOEMCOBRANCH="$OPTARG" ;;
      e ) EMCOBRANCH="$OPTARG" ;;
      ? ) helpInputs ;; # Print helpInput in case parameter is non-existent
   esac
done

cloneTestRepos
