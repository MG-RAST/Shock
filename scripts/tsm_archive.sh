#!/bin/sh

# usage: ./tsm_backup [-v]
# audience: Shock admin
# connect with Shock to retrieve list of files to be moved to TSM 
# connect with TSM to get list of files already in TSM
# for each file in Shock list,
#     check if file is already in TSM (with JSON{"id": "${LOCATION_NAME}", "stored": = "true" } )
#     submit via dsmc for archiving and add JSON to /node/${id}/location/${LOCATION_NAME}/ with { "id": "${LOCATION_NAME}", "stored": "false" }
# 

# send data items in Shock output to a TSM instance
# the Shockoutput will be of the form


# config of TSM is via the run time environment (e.g. the dsmc utilities and the server side config)


### ################################################################################
### ################################################################################
### ################################################################################
### Config variables start here

# config variables
# URL to server
SHOCK_SERVER_URL="https://shock.mg-rast.org"
# the DATA directory of the shock server
SHOCK_DATA_PATH="/dpool/mgrast/shock/data"
# name of the location defined in locations.yaml
LOCATION_NAME="anltsm"
# name of the dump file for TSM data
TSM_DUMP="/var/tmp/backup_list_${LOCATION_NAME}.txt"
# TSM servername
TSM_SERVER=TSM_CELS
# NOTE: we assume authentication bits to be contain in the AUTH env variable
WCOPY=${SHOCK_DATA_PATH}/$(basename $0)_wcopy.$$.txt
OUTCOPY=${SHOCK_DATA_PATH}/$(basename $0)_output.$$.txt
LOCKF=$${SHOCK_DATA_PATH}/$(basename $0).lock

### no more config
# AUTH is set externally
### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################

### return the age of a file in hours
function cleanup() {

# clean up
rm -f ${WCOPY} ${OUTCOPY} 
rm -f ${LOCKF} ${TSM_DUMP}

}

### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################


### return the age of a file in hours
function fileage() {
if [ ! -f $1 ]; then
  echo "file $1 does not exist"
        exit 1
fi
MAXAGE=$(bc <<< '24*60*60') # seconds in 28 hours
# file age in seconds = current_time - file_modification_time.
FILEAGE=$(($(date +%s) - $(stat -c '%Y' "$1")))
test $FILEAGE -lt $MAXAGE && {  # this is a very ugly hack and needs to return the actual hours..
    echo "23"
    exit 0
}
  echo "25"
}

### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################

###  extract a list of all items in TSM backup once every day
function update_TSM_dump () {

local filename=$1
local cachefiledate=$(fileage ${filename})

# check if cace
if [ -f $filename ] && [[ ${cachefiledate} -lt 24 ]]
then
  if [ ${verbose} == "1" ] ; then
    echo "using cached nodes file ($filename)"
  fi
else
  # capture nova output in file
  if [ ${verbose} == "1" ] ; then
    echo "creating new DUMP file ($filename)"
  fi
  dsmc query archive "${SHOCK_DATA_PATH}/*/*/*/*/*" > $filename
  chmod g+w ${filename} 2>/dev/null

fi
}

### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################

### set location with stored == false, indicating that data is in flight to TSM
function write_location() {
id=$1

local val=false
local JSON_STRING='{"id":"'"$LOCATION_NAME"'","stored":"'"$val"'"}'

curl -s -X POST -H "$AUTH" "${SHOCK_SERVER_URL}/node/${id}/locations/ -d ${JSON_STRING}"
}

### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################

### set Location as verified in Shock, confirming the data for said node is in TSM
verify_location() {
id=$1

local val=true
local JSON_STRING='{"id":"'"$LOCATION_NAME"'","stored":"'"$val"'"}'

curl -s -X POST -H "$AUTH" "${SHOCK_SERVER_URL}/node/${id}/locations/ -d ${JSON_STRING}" > $
}

### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################

#### write usage info
function usage() {
      echo "script usage: $(basename $0) [-v] [-h] -d <TSM_dumpfile> filename" >&2
      echo "connect with Shock to retrieve list of files to be moved to location" >&2
}

### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################

### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################

## main
# in case of error or interruption, clean up the temp files and exist cleanly
trap 'cleanup; exit 1' 0 1 2 3 15

#
while getopts 'vfh' OPTION; do
  case "$OPTION" in
    v)
      verbose="1"
      ;;
    f)
      force="1"
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
if [ ! -f $1 ] ; then
	usage
	exit 1	
fi

if ["${force}x" == "x" ] && [ -e ${LOCKF} ]; then 
  echo " [$(basename $0)] Lockfile ${LOCKF} exists; exiting"
  exit 1
fi

touch ${LOCKF}

if [[ $verbose == "1" ]] ;then 
  echo "Settings:"
  echo "SHOCK_SERVER_URL:\t\t${SHOCK_SERVER_URL}"
  echo "SHOCK_DATA_PATH:\t${SHOCK_SERVER_URL}" 
  echo "LOCATION_NAME:\t\t${LOCATION_NAME}"
  echo "TSM_DUMP:\t\t${TSM_DUMP}"
fi

# check if the dsmc command is available
if [ ! -x "$(which dsmc)" ] ; then
  echo " [$(basename $0)] requires the IBM TSM dsmc command to be installed, configured and available in PATH"
  exit 1
fi

if [ ! "${force}x" -eq "x" ] ; then 
  if [ -x ${WCOPY} ] || [ -x ${OUTCOPY} ] ; then
     echo " [$(basename $0)] Output files from last run exist, exiting. (check: ${WCOPY} and/or ${OUTCOPY}) "
    exit 1
  fi
fi
 

# download the file of nodes that require submission to DSMC from SHOCK
curl -s -X POST -H "$AUTH" "${SHOCK_SERVER_URL}/location/${LOCATION_NAME}/missing" | jq .data[].id | tr -d \"  > ${WCOPY}
if [ $? != 0 ] ; then 
  echo " [$(basename $0)] Can't connect to ${SHOCK_SERVER_URL} or disk full"
  exit 1
fi


## read a file dumped by shock
writecount=0
verifycount=0
missingcount=0

while read line; do 
    id=$(echo $line)  
    if [[ $verbose == "1" ]] ; then 
	    echo "working on $id"
    fi

    # add the data files and the idx directory to the request file
    DATAFILE="${SHOCK_DATA_PATH}/*/*/*/*/${id}.data"
    INDEX="${SHOCK_DATA_PATH}/*/*/*/*/${id}/idx"


    # check if all data and index are in the backup already
    if [ fgrep -q ${DATAFILE} DSMCDB ] && [ fgrep -q ${INDEXFILE} DSMCDB ] ; then
      if [[ $verbose == "1" ]] ; then 
	      echo "$id already found in TSM"
      fi
      JSON=$(verify_location ${id} )
      if echo ${JSON} |  grep -q 200  ; then 
        verifycount=`expr $verifycount + 1`
      elif echo ${JSON}| grep -q "Node not found" ; then
        missingcount=`expr $missingcount + 1`
      else
        echo "$(basename $0) can't write to ${SHOCK_SERVER_URL}; exiting (node: ${id})" >&2
        echo "RAW JSON: \n${JSON}\n"
        exit 1
     fi 

    else  # add data and index to request file
      if [[ $verbose == "1" ]] ; then 
	      echo "${id} NOT found in TSM"
      fi

      # check if node is already requested
      JSON=$(curl -s -X GET -H "$AUTH" "${SHOCK_SERVER_URL}/node/${id}/locations/${LOCATION_NAME}/" )

      if echo ${JSON} |  grep -q locations.stored="false" ; then 
        # already requested skip to next ${id}
      else 
         # write names to request file
         echo "${DATAFILE}" >> ${OUTCOPY}
         echo "${INDEXFILE}" >> ${OUTCOPY}

         JSON=$(write_location ${id} )

         if echo ${JSON} |  grep -q 200  ; then 
            writecount=$(expr $writecount + 1 )
         else
            echo " [$(basename $0)] can't write to ${SHOCK_SERVER_URL}; exiting (node: ${id})" >&2
            echo "RAW JSON: \n${JSON}\n"
            exit 1
          fi 
      fi  
done <${WCOPY}

if [[ ${verbose == "1" } ]] ; then
  echo "found ${writecount} items to add to TSM"
  echo "found ${verifycount} items to confirm as in TSM"
  echo "found ${missingcount} nodes missing in MongoDB)"
fi

# run the command to request archiving

  if [[ $verbose == "1" ]] ; then
    echo "running dsmc archive -filelist=${OUTCOPY} > /dev/null"
  fi

# capture the return value and report any errors
 ret=$(dsmc archive -filelist=${OUTCOPY}) 
 if [ $? != 0 ] ; then
    echo "FAILED: dsmc archive -filelist=${OUTCOPY} "
    echo "OUTPUT: ${ret}"
    cleanup
    exit 1
  fi

# run cleanup function
cleanup
exit 0

