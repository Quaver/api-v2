# api-v2

> 🌎 The new API server for Quaver.

**api-v2** is the most up-to-date (v2) and open source version of the Quaver web API.

As endpoints are made available in v2, it is recommended to use them instead of its v1 counterpart, as v1 will be deprecated, as we begin to update usage in-game and on our website.

**This application is being developed for internal network use. As such, no support will be provided for the usage of this software.**

## Requirements

- Go 1.22
- MariaDB / MySQL
- Redis
- ElasticSearch 8.14.1
- Steam Publisher API Key
- OpenAI API Key
- FFmpeg
- Compiled [Quaver.Tools Executable](https://github.com/Quaver/Quaver.API)
- Stripe CLI (for donations/store item development/testing)
- Postman (for testing)
  
## Setup

- Install `Go 1.22` or later.
- Clone the repository.
- Copy `config.example.json` and make a file named `config.json`
- Fill out the config file with the appropriate details.
  - **Note:** `quaver_tools_path` is the path to the Quaver.Tools executable, not the directory.
  - `oauth_access_token_secret` must be a long, random, server-only value. Never provide it to an OAuth client.
- Navigate to the `/cmd/api/` directory.
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

## OAuth 2.0

OAuth applications are managed through the `/v2/developers/applications` endpoints.
Applications used only for the client-credentials grant may have an empty redirect
URL. Authorization-code applications must use an absolute HTTP(S) redirect URL.

The v2 OAuth routes are:

- `GET /v2/oauth2/authorize` — validates an authorization request and returns safe application metadata for the consent UI.
- `POST /v2/oauth2/authorize` — requires an authenticated user, creates a one-time code, and redirects to the registered URL with `code` and optional `state`.
- `POST /v2/oauth2/token` — accepts `authorization_code` and `client_credentials` grants.
- `POST /v2/oauth2/me` — verifies a bearer access token and returns the current serialized user.

Authorization-code token requests must include `client_id`, `client_secret`,
`grant_type=authorization_code`, `code`, and the exact registered `redirect_uri`.
Client-credentials requests use `client_id`, `client_secret`, and
`grant_type=client_credentials`. Access tokens are bearer tokens valid for one hour;
refresh tokens are not issued.

OAuth access tokens are signed with the server-only `oauth_access_token_secret`;
the application `client_secret` authenticates the client but cannot mint access
tokens. Deploying this change invalidates access tokens issued by the previous
client-secret signing scheme, so clients must obtain new tokens.

The `/v2/oauth2/me` endpoint accepts `Authorization: Bearer <access_token>`. For
compatibility with the existing OAuth documentation, the same token may also be
sent in a `code` form or JSON body field.

## LICENSE

This software is licensed under the **GNU Affero General Public License v3.0.** Please see the LICENSE file for more information.
