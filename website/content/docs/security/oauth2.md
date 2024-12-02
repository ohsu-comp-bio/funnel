---
title: OAuth2
menu:
  main:
    parent: Security
    weight: 10
---
# OAuth2

By default, a Funnel server allows open access to its API endpoints, but in
addition to Basic authentication it can also be configured to require a valid
JWT in the request.

Funnel itself does not redirect users to perform the login.
It just validates that the presented token is issued by a trusted service
(specified in the YAML configuration file) and the token has not expired.
In addition, if the OIDC provides a token introspection endpoint (in its
configuration JSON), Funnel server also calls that endpoint to make sure the
token is still active (i.e., no token invalidation before expiring).

Optionally, Funnel can also validate the scope and audience claims to contain
specific values.

It is not possible to configure administrative users for the OAuth2 authentication mode.

To enable JWT authentication, specify `OidcAuth` section in your config file:

```yaml
Server:
  OidcAuth:
    # URL of the OIDC service configuration:
    ServiceConfigURL: "https://my.oidc.service/.well-known/openid-configuration"

    # Client ID and secret are sent with the token introspection request
    # (Basic authentication):
    ClientId: your-client-id
    ClientSecret: your-client-secret

    # Optional: if specified, this scope value must be in the token:
    RequireScope: funnel-id

    # Optional: if specified, this audience value must be in the token:
    RequireAudience: tes-api

    # The URL where OIDC should redirect after login (keep the path '/login')
    RedirectURL: "http://localhost:8000/login"
```

Make sure to properly protect the configuration file so that it's not readable
by everyone:

```bash
$ chmod 600 funnel.config.yml
```

Note that the Funnel UI supports login through an OIDC service. However, OIDC
authentication is not supported at command-line.
