OpenBCI golang server
=====================

OpenBCI golang server allows users to control, visualize and store data collected from the OpenBCI microcontroller.

Requirements
------------

* Go <https://golang.org> (Confirmed compatible with v1.3 and v1.4)
* FFTW 3.x <http://fftw.org> (Perhaps check your package manager first)
* The openbci-golang-server binary for your distribution not yet provided
* FTDI Virtual Com Port Driver <http://www.ftdichip.com/Drivers/VCP.htm>

To test and contribute code
---------------------------

* A Go developement environment <https://golang.org/doc/code.html>  
* The following additional go packages:  

        go get github.com/gorilla/websocket  
        go get github.com/tarm/goserial  
        go get github.com/runningwild/go-fftw  
        go get github.com/orfjackal/gospec (For running the go-fftw tests)  
  
* To clone, test, build and run the server run:  

        git clone https://github.com/kevinjos/openbci-golang-server.git  
        cd openbci-golang-server  
        go test  
        go build  
        ./openbci-golang-server  

* By default, the server points to <http://localhost:8888>

