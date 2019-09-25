#ensure latest container is built
(cd ../; docker build -t mgrast/shock .)

#build test image
docker build -t mgrast/shock-pytest .

#start shock
docker-compose up


# run test
docker run -ti --rm --network=shock-test_default mgrast/shock-pytest py.test -k test_shock
