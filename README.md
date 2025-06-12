# freeathome

A client library for the BUSCH-JAEGER free@home local API implemented in Golang.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pgerke/freeathome)
![CI](https://img.shields.io/github/actions/workflow/status/pgerke/freeathome/ci.yaml?style=flat-square)
[![codecov](https://codecov.io/gh/pgerke/freeathome/branch/main/graph/badge.svg?token=UJQVXZ5PPM)](https://codecov.io/gh/pgerke/freeathome)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=pgerke_freeathome&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=pgerke_freeathome)
![License](https://img.shields.io/github/license/pgerke/freeathome?style=flat-square)

## Installation

To use this library in your own Go project, make sure your project is using Go modules. Then, run the following command to add the dependency:

```sh
go get github.com/pgerke/freeathome@latest
```

You can then import the package in your code:

```go
import "github.com/pgerke/freeathome"
```

This will give you access to the public API client and related utilities for interacting with a local free\@home SysAP.

## Features

The project initially covers the scope of the [TypeScript free@home API Client](https://github.com/pgerke/freeathome-local-api-client).

- Connect to your B+J System Access Point 2.0 and control it using the local API.
- 100% covered by automated unit tests
- Websocket communication with keepalive
- Get configuration
- Get device list
- Get device
- Create virtual device
- Get and set datapoints
- Trigger proxy device
- Set proxy device value
- Default and custom loggers!

## Usage Requirements

- A free@home System Access Point 2.0 running firmware > v3.0
- Local API has to be enabled for the user account to be used

## Documentation

The API documentation is available at https://pkg.go.dev/github.com/pgerke/freeathome.

## I found a bug, what do I do?

I'm happy to hear any feedback regarding the library or it's implementation, be it critizism, praise or rants. Please create a [GitHub issue](https://github.com/pgerke/freeathome/issues) or drop me an [email](mailto:info@philipgerke.com) if you would like to contact me.

I would especially appreciate, if you could report any issues you encounter while using the library. Issues I know about, I can probably fix.

If you want to submit a bug report, please check if the issue you have has already been reported. If you want to contribute additional information to the issue, please add it to the existing issue instead of creating another one. Duplicate issues will take time from bugfixing and thus delay a fix.

While creating a bug report, please make it easy for me to fix it by giving us all the details you have about the issue. Always include the version of the library and a short concise description of the issue. Besides that, there are a few other pieces of information that help tracking down bugs:

- The system environment in which the issue occurred
- Some steps to reproduce the issue, e.g. a code snippet
- The expected behaviour and how the failed failed to meet that expectation
- Anything else you think I might need

## I have a feature request, what do I do?

Please create a [GitHub issue](https://github.com/pgerke/freeathome/issues) or drop me an [email](mailto:info@philipgerke.com)!

## Non-Affiliation Disclaimer

This library is not endorsed by, directly affiliated with, maintained, authorized, or sponsored by Busch-Jaeger Elektro GmbH or ABB Asea Brown Boveri Ltd or . All product and company names are the registered trademarks of their original owners. The use of any trade name or trademark is for identification and reference purposes only and does not imply any association with the trademark holder of their product brand.

## License

The project is subject to the MIT license unless otherwise noted. A copy can be found in the root directory of the project [LICENSE](./LICENSE).

<hr>

Made with ❤️ by [Philip Gerke](https://github.com/pgerke)
