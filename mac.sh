#!/usr/bin/env bash

echo "##########################"
echo "#   TEST FOR HYPERCHAIN  #"
echo "##########################"

echo "kill the bind port process"

kill -9 $(lsof -i :8001 | awk 'NR>=2{print $2}')
kill -9 $(lsof -i :8002 | awk 'NR>=2{print $2}')
kill -9 $(lsof -i :8003 | awk 'NR>=2{print $2}')
kill -9 $(lsof -i :8004 | awk 'NR>=2{print $2}')

#rebuild the application
echo "rebuild the application"

#govendor build

echo "run the application"
# change the terminal program here!!
open -n -a Terminal "(./hyperchain -o 1 -l 8081)"
#gnome-terminal -x bash -c "(./hyperchain -o 1 -l 8081)"
#gnome-terminal -x bash -c "(./hyperchain -o 1 -l 8081)"
#gnome-terminal -x bash -c "(./hyperchain -o 2 -l 8082)"
#gnome-terminal -x bash -c "(./hyperchain -o 3 -l 8083)"
#gnome-terminal -x bash -c "(./hyperchain -o 4 -l 8084)"

echo "All process are running background"