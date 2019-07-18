################################################################################
# Go
################################################################################

# Go workspace
export GOPATH=$HOME/Desktop/w/go

export GOBIN=$HOME/Desktop/w/go/bin

# Add Go bin folder to path
# Ensure path is only added once when reloading
if [[ $PATH != *":$GOPATH/bin"* ]]; then
  export PATH=$PATH:$GOPATH/bin
fi

export GO111MODULE=on
