
#!/bin/sh

export DIR=" /go/src/github.com/MG-RAST/Shock"

cd ${DIR}
#go get -d ./shock-server/ ./shock-client/
CGO_ENABLED=0 go install -installsuffix cgo $1 -v -ldflags="-X github.com/MG-RAST/Shock/shock-server/conf.VERSION=$(git describe --tags)" ./shock-server/


