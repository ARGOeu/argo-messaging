[![Build Status](https://travis-ci.org/ARGOeu/argo-messaging.svg?branch=devel)](https://travis-ci.org/ARGOeu/argo-messaging)
# ARGO Messaging

## Development

1. Install Golang and bzr library
2. Create a new work space:

      `mkdir ~/go-workspace`
      `export GOPATH=~/go-workspace`
      `export PATH=$PATH:$GOPATH/bin`

  You may add the last `export` line into the `~/.bashrc` or the `~/.bash_profile` file to have `GOPATH` environment variable properly setup upon every login.

3. Get the latest version

      `go get github.com/ARGOeu/argo-messaging`

4. Get dependencies:

   Argo-messaging uses godep tool for dependency handling.
   To install godep tool issue:

      `go get github.com/tools/godep`
      `godep restore`
      `godep update ...`

5. To build the service use the following command:

      `go build`

6. To run the service use the following command:

      `./argo-messaging`

7. To run the unit-tests:

      `go test ./...`

8. To generate and serve godoc (@port 6060)

      `godoc -http=:6060`


## Credits

The ARGO Messaging Service is developed by [GRNET](http://www.grnet.gr)

The work represented by this software was partially funded by the EGI-ENGAGE project through the European Union (EU) Horizon 2020 program under Grant number 654142.
