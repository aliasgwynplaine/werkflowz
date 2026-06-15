set -e
go build src/init/init.go
go build src/finish/finish.go
go build src/pingpong/pingpong.go
echo "Done!"
