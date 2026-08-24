if [ $# -ne 1 ]; then
	echo "fuck you"
	exit 1
fi

oarid=$1
mapfile -t workers < workers.$oarid


for wrkr in ${workers[@]}; do
	echo $wrkr > tmp.wrkr
	./c2.sh kill-all-engines tmp.wrkr
	wait
	./c2.sh kill-gateway tmp.wrkr
	wait
	./c2.sh run-gateway tmp.wrkr micro-ccmesg-box
	./c2.sh run-engines tmp.wrkr micro-ccmesg-box tmp.wrkr
	sleep 1
	./c2.sh run-launcher $wrkr micro-ccmesg-box 1 tmp.wrkr
	./c2.sh run-launcher $wrkr micro-ccmesg-box 2 tmp.wrkr
	./c2.sh run-launcher $wrkr micro-ccmesg-box 3 tmp.wrkr
done

rm tmp.wrkr
