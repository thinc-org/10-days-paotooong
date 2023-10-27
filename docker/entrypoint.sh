#!/bin/sh

show_usage() {
	echo "$0 [proxy|grpc]"
	exit 1
}

if [ $# -ne 1 ]; then
	show_usage
fi

case $1 in
	grpc)
		echo "Starting grpc server"
		/app/grpc
	;;
	proxy)
		echo "Starting http server"
		/app/proxy
	;;
	*)
		echo "Invalid option"
		show_usage
	;;
esac

