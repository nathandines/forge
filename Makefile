.PHONY: run test clean

bin/stack:
	go build -o bin/stack

test:
	cd stacklib && \
	go test -cover

clean:
	rm -rf bin
