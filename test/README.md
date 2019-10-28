
# start test environment
 
```bash
#ensure latest container is built
(cd ../; docker build -t mgrast/shock .)

#build test image
docker build -t mgrast/shock-pytest .

#start shock
docker-compose up
```

# run test
```bash
docker run -ti --rm --network=shock-test_default mgrast/shock-pytest py.test -k test_shock
```


# (run test against production)
```bash
docker run -ti --rm --env "SHOCK_URL=https://shock.mg-rast.org" --env "SHOCK_ADMIN_AUTH=${SHOCK_ADMIN_AUTH}" --env "SHOCK_USER_AUTH=${SHOCK_USER_AUTH}" mgrast/shock-pytest py.test -k test_shock
```