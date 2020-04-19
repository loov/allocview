
testdata-graph:
	go build -o ./graph.exe ./testdata/graph
	go build -o ./allocview.exe .
	./allocview.exe ./graph.exe

testdata-leaking:
	go build -o ./graph.exe ./testdata/graph
	go build -o ./allocview.exe .
	./allocview.exe ./graph.exe
