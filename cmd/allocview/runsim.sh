set +x

go build -o ../../testdata/leaking.exe ../../testdata/leaking.go
go build -o allocview.exe .
./allocview.exe -simulate $@
