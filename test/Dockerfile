
#docker build -t mgrast/shock-pytest -f Dockerfile_pytest .
#docker run -ti --rm --network=shock-test_default mgrast/shock-pytest

FROM alpine

RUN apk update && apk add \
    python3

RUN pip3 install --upgrade pip

RUN pip3 install \
    pytest \
    requests 

COPY testdata /testing/testdata/
COPY test_shock.py /testing/
WORKDIR /testing


# example single test:
# py.test -k test_querynode_name

# example all tests:
# py.test -k test_shock

# 
# curl -X POST -H "Authorization: ${SHOCK_AUTH}" -F 'attributes_str={test:"hello"}' ${SHOCK_URL}/node
