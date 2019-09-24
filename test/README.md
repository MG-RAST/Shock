
# testing

```bash
#build test image
docker build -t mgrast/shock-pytest .

#start shock
docker-compose up


# run test
docker run -ti --rm --network=shock-test_default mgrast/shock-pytest py.test -k test_shock

```

