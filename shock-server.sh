#!/bin/sh -e
# Shock auto-start
#
# description: auto-starts Shock server
# processname: shock-server
# pidfile: /var/run/shock-server.pid
# logfile: /var/log/shock-server.log
# config: /etc/shock/shock-server.conf
 
NAME="shock-server"
LOG_FILE="/var/log/${NAME}.log"
PID_FILE="/etc/shock/data/pidfile"
CONF_FILE="/etc/shock/${NAME}.conf"

start() {
    echo -n "Starting $NAME... "
    if [ -f $PID_FILE ]; then
	    echo "is already running!"
    else
	    $NAME -conf $CONF_FILE > $LOG_FILE 2>&1 &
	    sleep 1
	    echo "(Done)"
    fi
    return 0
}
 
stop() {
    echo -n "Stopping $NAME... "
    if [ -f $PID_FILE ]; then
	    PIDN=`cat $PID_FILE`
	    kill $PIDN 2>&1
	    sleep 1
	    rm $PID_FILE
	    echo "(Done)"
    else
	    echo "can not stop, it is not running!"
    fi
    return 0
}

status() {
    if [ -f $PID_FILE ]; then
	    PIDN=`cat $PID_FILE`
	    PSTAT=`ps -p $PIDN | grep -v -w 'PID'`
	    if [ -z "$PSTAT" ]; then
	        echo "$NAME has pidfile ($PIDN) but is not running."
	    else
	        echo "$NAME is running with pid $PIDN."
	    fi
    else
	    echo "$NAME is not running."
    fi
    return 0
}

case "$1" in
    start)
	    start
	    ;;
    stop)
	    stop
	    ;;
    restart)
	    stop
	    sleep 5
	    start
	    ;;
    status)
	    status
	    ;;
    *)
	    echo "Usage: $0 (start | stop | restart | status)"
	    exit 1
	    ;;
esac
