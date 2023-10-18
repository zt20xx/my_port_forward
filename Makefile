.PHONY: all clean

all: myclient myserver

myclient: myclient/myc.go myclient/myc.ini
	cd myclient && go build -o ../cmd/myclient myc.go
	cp myclient/myc.ini cmd/

myserver: myserver/mys.go myserver/mys.ini
	cd myserver && go build -o ../cmd/myserver mys.go
	cp myserver/mys.ini cmd/

clean:
	rm -rf cmd/myclient cmd/myc.ini cmd/myserver cmd/mys.ini

