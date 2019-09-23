
# testing


```bash
#start shock
docker-compose up

#build test image
docker build -t mgrast/shock-pytest -f Dockerfile_pytest .

# run test
docker run -ti --rm --network=shock-test_default mgrast/shock-pytest py.test -k test_shock

```

