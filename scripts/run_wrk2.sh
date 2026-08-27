set -e

if [ $# -ne 1 ]; then
	exit 1
fi

if [ ! -d "data" ]; then
	echo "[!] data folder not found. calling mkdir"
	mkdir data
fi

rps=$1
echo "[*] Warming up"
./wrk2/wrk -t2 -c64 -d30 -R${rps} --latency http://127.0.0.1:3000 > /dev/null
echo "[*] Benchmarking"
./wrk2/wrk -t2 -c64 -d120 -R${rps} --latency http://127.0.0.1:3000 > data/${rps}_wrk2.txt
