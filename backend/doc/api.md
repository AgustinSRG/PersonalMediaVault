# Backend API

This document specifies the HTTP API used to interact with the PersonalMediaVault backend.

## Authentication

The PersonalMediaVault backend API uses an API key authentication system, based on a session token you generate by calling the login API.

### Obtaining a session token

First, you must call the Login API:
 - Method: **POST**
 - Path: `/api/auth/login`
 - Body: **JSON** with the login credentials:

```json
{
    "username": "{Username}",
    "password": "{Password}"
}
```

The response, if the link credentials are correct, will be a **200 OK** response with a **JSON** body, containing the session token, along with the vault fingerprint (a string to uniquely identify the vault):

```json
{
    "session_id": "{Session Token}",
    "vault_fingerprint": "{Vault fingerprint}"
}
```

### Authenticating with the session token

In order to authenticate with the session token, pass it in the request, with a custom header named `x-session-token`.

```sh
curl -X 'GET' \
  'http://{VAULT_HOST}/api/account' \
  -H 'accept: application/json' \
  -H 'x-session-token: {SESSION_TOKEN}'
```

### Alternative authentication for assets

Due to most players and some tools not supporting sending headers in the request, two alternative authentication methods are supported:

 - Passing the session token in the URL, as a query parameter with the name `session_token`.

```
http://{VAULT_HOST}/assets/b/0/0/video.mp4?session_token={SESSION_TOKEN}
```

 - Passing the session token in a cookie with a name made by concatenating `st-` and the vault fingerprint.

```
Set-Cookie: st-{VAULT_FINGERPRINT}={SESSION_TOKEN}
```

Note: these methods are only available for assets URLs, they won't work for any API request to prevent any CSRF security issues.

## API specification

The API is documented in YAML format, using the Swagger specification. You can use tools like [Swagger Editor](https://editor.swagger.io/) to generate clients for different programming languages and frameworks.

[Go to API specification](./api-docs.yml)

