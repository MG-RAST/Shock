#!/bin/sh

# usage: cmd filename

# read a file from dsmc (aka IBM TSM) to extract object IDs that are backed up


# config variables
# URL to server
SHOCK_SERVER_URL="https://shock.mg-rast.org"
# name of the location defined in locations.yaml
LOCATION_NAME="anltsm"
# NOTE: we assume authentication bits to be contain in the AUTH env variable


# write to SHOCK
write_location() {
id=$1
curl -s -X POST -H "$AUTH" "${SHOCK_SERVER_URL}/node/${id}/locations/${LOCATION_NAME}"
}

# write usage
usage() {
      echo "script usage: $(basename $0) [-v] [-h] filename" >&2
}


## main

#
while getopts 'vh' OPTION; do
  case "$OPTION" in
    v)
      verbose="1"
      ;;

      h)
      echo "$(basename $0) -h --> display this help" >&1
      usage
      exit 0
      ;;
    ?)
      usage
      exit 1
      ;;
  esac
done
shift "$(($OPTIND -1))"

# check if parameter is file
if [ ! -f $1 ]
then
	usage
	exit 1	
fi


newcount=0
existcount=0
missingcount=0
# read the file
WCOPY=/tmp/working_copy$$.txt
tail -n+14 $1  > ${WCOPY}

while read line; do 
    id=$(echo $line | awk ' { print $7 } ' | cut -d/ -f 9)
    if [[ $verbose == "1" ]]
    then 
	echo "working on $id"
    fi
    JSON=$(write_location ${id} )

    if echo ${JSON} |  grep -q 200  ; then 
      newcount=`expr $newcount + 1`
     elif echo ${JSON}| grep -q 500 ; then
      existcount=`expr $existcount + 1`
     elif echo ${JSON}| grep -q "Node not found" ; then
      missingcount=`expr $missingcount + 1`
    else
      echo "$(basename $0) can't write to ${SHOCK_SERVER_URL}; exiting (node: ${id})" >&2
      echo "RAW JSON: \n${JSON}\n"
     exit 1
     fi 

    # check to see if it is 200 or 500

done < ${WCOPY}
 
echo "wrote $newcount locations to ${SHOCK_SERVER_URL} (${existcount} already in place) (${missingcount} nodes missing)"

rm -f ${WCOPY}
