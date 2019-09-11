#!/bin/sh

# usage: ./tsm_restore [-v] node(s)
# audience: Shock admin
# purpose: restore file(s) from a TSM location into the data path of a shock server

# example:
# "./tsm_restore -v 0001a988-f2a2-4c55-a42e-dd28d42d0344 0001c1ed-58a3-44ca-8be5-f05a5d37da5b"

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
TSM_DUMP="/var/tmp/restore_list_${LOCATION_NAME}.txt"
# TSM servername
TSM_SERVER=TSM_CELS
# NOTE: we assume authentication bits to be contain in the AUTH env variable
WCOPY=${SHOCK_DATA_PATH}/$(basename $0)_wcopy.$$.txt
OUTCOPY=${SHOCK_DATA_PATH}/$(basename $0)_output.$$.txt
LOCKF=$${SHOCK_DATA_PATH}/$(basename $0).lock

### no more config
### ################################################################################
### ################################################################################
function cleanup() {
rm -f ${WCOPY} ${OUTCOPY} 
rm -f ${LOCKF} ${TSM_DUMP}
}
### ################################################################################
### ################################################################################

### check_node for { "stored" : "true" } indicating that data is in TSM
function check_node() {

local loc_id=$1
local node_id=$2

set -x
# check if node is already requested
cmd="curl -s -X GET -H \"$AUTH\" \"${SHOCK_SERVER_URL}/node/${node_id}/locations/${loc_id} \" "

${cmd}

# is the node store in TSM?
if echo ${JSON} |  grep -q locations.stored="true" ; then 
    # if [[ $verbose == "1" ]] ; then 
	#   echo "[$(basename $0)] node ${id} found store at ${loc_id}"
    # fi
    return 0
else
    # if [[ $verbose == "1" ]] ; then 
    #   echo "[$(basename $0)] node ${id} is NOT stored at ${loc_id}"
    # fi 
    return 1
fi
}
### ################################################################################
### ################################################################################
### ################################################################################
### ################################################################################

#### write usage info
function usage() {
      echo "script usage: $(basename $0) [-v] [-h] node1 [node2 node3] ..." >&2
      echo "connect with TSM to restore nodes to the correct data path" >&2
}

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



# check if at least one argument is provided
if [ $# -lt 1 ] ; then
	usage
	exit 1	
fi

if [ "${force}x" == "x" ] && [ -e ${LOCKF} ]; then 
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
# if [ ! -x "$(which dsmc)" ] ; then
#   echo " [$(basename $0)] requires the IBM TSM dsmc command to be installed, configured and available in PATH"
#   exit 1
# fi


# remove any remaining tmp files
rm -f ${TSM_DUMP}
### create correct paths in a temporary filelist NOTE: this requires the WILDCARDSareliteral Option to be set for DSMC
for node in $@
do
    ret= check_node ${node} ${LOCATION_NAME}
    if [[ ${ret} -eq 0 ]] ; then 
        echo "${SHOCK_DATA_PATH}/*/*/*/*/${name}/*" >> ${TSM_DUMP}
    else
        if [[ $verbose == "1" ]] ; then
        echo "[$(basename $0)] NODE ${name} not present in ${LOCATION_NAME}"
        fi
    fi
done

# restore command to original location
cmd="dsmc1 restore -filelist=${TSM_DUMP} -subdir=yes"

# run the command to request archiving
if [[ $verbose == "1" ]] ; then
   echo "[$(basename $0)] running ${cmd}"
fi

# capture the return value and report any errors
ret=$(${cmd})
if [[ $? -ne 0 ]] ; then
   echo "[$(basename $0)] FAILED: ${cmd}"
  echo "[$(basename $0)] OUTPUT: ${ret}"
   cleanup
   exit 1
fi

cleanup
exit 0

