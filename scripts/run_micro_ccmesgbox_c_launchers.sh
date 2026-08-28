if [ $# -ne 1 ]; then
	echo "fuck you"
	exit 1
fi

oarid=$1
mapfile -t workers < workers.$oarid
echo "[*] Killing redis"
./c2.sh remote-kill redis.$oarid redis
echo "[*] Killing cache servers... "
./c2.sh remote-kill workers.$oarid hz-server
echo "[*] Killing scheduler... "
./c2.sh remote-kill scheduler.$oarid scheduler

sleep 2

echo "[*] Restarting infrastructure"
./c2.sh redis-setup redis.$oarid
./c2.sh ccmesh-redis-run redis.$oarid

sleep 1

./c2.sh verify-redis redis.$oarid

./c2.sh ccmesh-server-run workers.$oarid
./c2.sh run-scheduler scheduler.$oarid

./c2.sh verify-scheduler scheduler.$oarid
./c2.sh verify-cache-servers workers.$oarid

echo "[*] Restarting application: micro-ccmesh"
./c2.sh kill-all-engines workers.$oarid
./c2.sh kill-gateway gateway.$oarid
./c2.sh kill-snitch gateway.$oarid

./c2.sh run-gateway gateway.$oarid micro-ccmesg-box
./c2.sh run-snitch gateway.$oarid

sleep 5

./c2.sh run-engines workers.$oarid micro-ccmesg-box gateway.$oarid

for wrkr in ${workers[@]}; do
	./c2.sh run-launcher $wrkr micro-ccmesg-box 1 gateway.$oarid
	./c2.sh run-launcher $wrkr micro-ccmesg-box 2 gateway.$oarid
	./c2.sh run-launcher $wrkr micro-ccmesg-box 3 gateway.$oarid
	./c2.sh run-launcher $wrkr micro-ccmesg-box 4 gateway.$oarid
	./c2.sh run-launcher $wrkr micro-ccmesg-box 5 gateway.$oarid
	./c2.sh run-launcher $wrkr micro-ccmesg-box 6 gateway.$oarid
done
