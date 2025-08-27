#!/bin/sh

nohup ./main  & echo $! > main.pid
CHILD_PID=$( cat main.pid )
(
    while true
    do 
      if [ -f /opt/exit-signals/EXIT ]
      then
        kill -TERM ${CHILD_PID}
        echo "Killed ${CHILD_PID} because the main container terminated."
      fi
      sleep 1
      done
) &
wait ${CHILD_PID}
