# api2

> 🌎 The new API server for Quaver.

**api2** is the most up-to-date (v2) and open source version of the Quaver web API.

As endpoints are made available in v2, it is recommended to use it instead of its v1 counterpart, as v1 will be made obsolete as we begin to update usage in-game and on our website.

**This application is being developed for internal network use. As such, no support will be provided for the usage of this software.**

## Requirements

- Go 1.19
- MariaDB / MySQL
- Redis
- Steam API Key
- Steam Publisher API Key
- Postman (for testing)
  
## Setup

- Install `Go 1.19` or later.
- Clone the repository.
- Copy `config.example.json` and make a file named `config.json`
- Fill out the config file with the appropriate details.
- Start the server with `go run .` or your method of choice.
- The server is now available at `http://localhost:8080` (or your desired port).

## Testing

There are two standards when it comes to testing the API — via built in go tests and Postman for request handlers.

### Unit Tests

Built in go tests are used for any unit testing. Simply run `go test` with the appropriate test(s) you would like to run. 

### Integration Tests

We test all request handlers through a [Postman collection]().

## LICENSE

This software is licensed under the **GNU Affero General Public License v3.0.** Please see the LICENSE file for more information.