if [ $# -ne 1 ]; then
	echo "fuck you"
	exit 1
fi

oarid=$1
mapfile -t workers < workers.$oarid

./c2.sh kill-all-engines workers.$oarid
./c2.sh kill-gateway gateway.$oarid

./c2.sh run-gateway gateway.$oarid micro-ccmesg
wait
./c2.sh run-engines workers.$oarid micro-ccmesg gateway.$oarid

for wrkr in ${workers[@]}; do
	./c2.sh run-launcher $wrkr micro-ccmesg 1 gateway.$oarid
	./c2.sh run-launcher $wrkr micro-ccmesg 2 gateway.$oarid
	./c2.sh run-launcher $wrkr micro-ccmesg 3 gateway.$oarid
done
