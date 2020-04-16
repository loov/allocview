set +x

go build -o allocview.exe .
./allocview.exe -simulate $@
