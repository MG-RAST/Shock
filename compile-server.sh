
#!/bin/sh
set -x

CGO_ENABLED=0 go install -installsuffix cgo $1 -v -ldflags="-X github.com/MG-RAST/Shock/shock-server/conf.VERSION=$(git describe --tags --long)" ./shock-server/


