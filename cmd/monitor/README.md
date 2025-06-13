# free@home Monitor

A monitor application that connects to the free@home System Access Point's local API websocket and logs device updates using the free@home Golang library.

The monitor application output will be written to STDOUT, while the device logs and messages from the System Access Point will be written to STDERR in the `logfmt` format.

## Usage Requirements

- A free@home System Access Point 2.0 running firmware > v3.0
- Local API has to be enabled for the user account to be used

## Running the monitor

Regardless of how you want to run the monitor you'll need to configure the System Access Point connection. The monitor will take the hostname and credentials from the following environment variables:

- `SYSAP_HOST`: The host name or IP address of the free@home System Access Point.
- `SYSAP_USER_ID`: The user id of the local API user. This is probably a GUID.
- `SYSAP_PASSWORD`: The password corresponding to the user ID.

You can either set the environment variables manually, or you can put them in an `.env` file like so:

```
SYSAP_HOST=192.168.178.123
SYSAP_USER_ID=01234567-89ab-cdef-0123-456789abcdef
SYSAP_PASSWORD=s3cr3t_p4ssW0rD
```

## Run locally

### Prerequisites

- The version of Golang as defined in the [go.mod](./../../go.mod) file.
- Environment variables are set as described above

### Build and Run

1. In the repository base directory install the dependencies by running `go mod tidy`.
1. Then you can build the monitor application by running `go build -o monitor ./cmd/monitor`
1. With the environment variables set, start the monitor applicatiton with `./monitor`.
1. To stop the monitor send an interrupt signal by pressing <kbd>Ctrl</kbd>+<kbd>C</kbd>. This triggers the graceful shutdown and may take a moment to complete.
1. To force a shutdown you can send a second SIGINT by pressing <kbd>Ctrl</kbd>+<kbd>C</kbd> again.

## Run in Docker

Pull the latest image from the GitHub Container Registry by calling `docker pull ghcr.io/pgerke/freeathome-monitor:latest`. Then you can run

```sh
docker run -it -e SYSAP_HOST="192.168.178.123" -e SYSAP_USER_ID="01234567-89ab-cdef-0123-456789abcdef" -e SYSAP_PASSWORD="s3cr3t_p4ssW0rD" ghcr.io/pgerke/freeathome-monitor:latest
```

Alternatively you can create an `.env` file as shown above and provide it to the container by calling:

```sh
docker run -it --env-file .env ghcr.io/pgerke/freeathome-monitor:latest
```

## I have a feature request or found a bug, what do I do?

Please create a [GitHub issue](https://github.com/pgerke/freeathome/issues)!

## Non-Affiliation Disclaimer

This library is not endorsed by, directly affiliated with, maintained, authorized, or sponsored by Busch-Jaeger Elektro GmbH or ABB Asea Brown Boveri Ltd. All product and company names are the registered trademarks of their original owners. The use of any trade name or trademark is for identification and reference purposes only and does not imply any association with the trademark holder of their product brand.

## License

The monitor application is subject to the MIT license unless otherwise noted.

<hr>

Made with ❤️ by [Philip Gerke](https://github.com/pgerke)
