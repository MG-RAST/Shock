#!/bin/bash

# converts shock-openapi.yaml into api.html

rm -f api.html index.html

set -x
set -e

docker run -ti --name swaggergenerate --rm -v ${PWD}:/local swaggerapi/swagger-codegen-cli-v3:3.0.11 generate \
    -i /local/shock-openapi.yaml \
    -l html2 \
    -o /local/

sleep 1

mv index.html api.html

