# api-v2

> 🌎 The new API server for Quaver.

**api2** is the most up-to-date (v2) and open source version of the Quaver web API.

As endpoints are made available in v2, it is recommended to use them instead of its v1 counterpart, as v1 will be deprecated, as we begin to update usage in-game and on our website.

**This application is being developed for internal network use. As such, no support will be provided for the usage of this software.**

## Requirements

- Go 1.22
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

Request handlers can be tested through [Postman](https://www.postman.com/00swan/workspace/quaver/collection/29785543-d09535f0-68bc-461d-920e-9d388c67f11b).

Some endpoints **require authentication.** To access them, you must [generate a JWT](https://jwt.io/), and set it in the `variables` section of the Postman collection.

#### Example JWT Payload

```json
{
  "user_id": 2,
  "username": "QuaverBot",
  "iat": 1516239022
}
```

## LICENSE

This software is licensed under the **GNU Affero General Public License v3.0.** Please see the LICENSE file for more information.
