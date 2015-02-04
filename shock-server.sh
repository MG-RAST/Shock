#!/bin/sh -e
# Shock auto-start
#
# description: auto-starts Shock server
# processname: shock-server
# pidfile: /var/run/shock-server.pid
# logfile: /var/log/shock-server.log
# config: /etc/shock/shock-server.conf
 
NAME="shock-server"
PID_FILE="/var/run/${NAME}.pid"
LOG_FILE="/var/log/${NAME}.log"
CONF_FILE="/etc/shock/${NAME}.conf"

start() {
    echo -n "Starting $NAME... "
    if [ -f $PID_FILE ]; then
	    echo "is already running!"
    else
	    $NAME -conf $CONF_FILE > $LOG_FILE 2>&1 &
	    sleep 2
	    echo `ps -ef | grep -v grep | grep 'shock-server' | awk '{print $2}'` > $PID_FILE
	    echo "(Done)"
    fi
    return 0
}
 
stop() {
    echo -n "Stopping $NAME... "
    if [ -f $PID_FILE ]; then
	    PIDN=`cat $PID_FILE`
	    kill $PIDN 2>&1
	    sleep 2
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
	    echo "$NAME is running with pid $PIDN."
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
