go build -o ../../testdata/leaking.exe ../../testdata/leaking.go
go build -o allocview.exe .
GODEBUG=allocfreetrace=1 ../../testdata/leaking.exe  3>&1 1>&2 2>&3 3>&- | ./allocview.exe