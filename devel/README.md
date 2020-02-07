

# change to devel dir
cd devel


# build the server from the
docker build --force-rm --no-cache --rm -t mgrast/shock ..


# start the server
docker-compose up


# in another shell (on the same machine) jump into the container
docker exec -ti `docker ps | fgrep mgrast/shock | awk ' { print $1 } ' ` ash

# start the server with the appropriate commands
shock-server --hosts=shock-mongo --expire_wait=1 --cache_path=/usr/local/shock/cache —debugauth=true --debuglevel=true
