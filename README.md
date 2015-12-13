OpenBCI golang server
=====================

OpenBCI golang server allows users to control, visualize and store data collected from the OpenBCI microcontroller.

Installation Requirements
-------------------------

* FFTW 3.x <http://fftw.org>
* FTDI Virtual Com Port Driver <http://www.ftdichip.com/Drivers/VCP.htm>
* Node package manager
* Go <https://golang.org>
* The following additional go packages:  

        go get github.com/gorilla/websocket  
        go get github.com/runningwild/go-fftw  
        go get github.com/orfjackal/gospec
        go get github.com/kevinjos/openbci-driver
        go get github.com/tarm/serial  
  

Installation Guide
------------------

        $ make all


Notes
-----

* By default, the server points to <http://localhost:8888>
* Builds on linux
* Node package manager alias used in Makefile may vary by distribution
