eok.vin
=======

Private, single-user, self-hosted URL shortening service.

> "These are words." -- anonymous

Usage
----- 

eokvin enforces TLS, relying on LetsEncrypt to issue certificates
automatically by default. Short URLs are created using a basic POST API
and expire after a set time.

```
Usage of ./eokvin:
  -cert-file string
        TLS certificate chain file, blank for autocert
  -hash-token string
        If given, the sha256 of the given value will be printed
  -host string
        Listen hostname (default "eok.vin")
  -http-port int
        HTTP listen port (default 80)
  -key-file string
        TLS private key file, blank for autocert
  -port int
        HTTPS listen port (default 443)
  -token string
        SHA256 of the secret token, used to authenticate
```

### Important notes

* A secret token is required to create new short URLs, given in the `-token`
  command-line option. This must be the hexadecimal string representation of
  the token (ie. 64 characters long) in question.
* Use the `-hash-token` command-line option to print the SHA256 result 
  (formatted as the `-hash` option expects) of a given phrase.
  ```
  $ ./eokvin -hash-token "this is a really long and arbitrary input"
  d1e6d468c926d9167693c190688d964fec0258c4ef4a4e1ed9cd87ea9c682156
  ```
* To avoid using LetsEncrypt (for example, when running locally), use a
  standard cert and key by using the `-cert-file` and `-key-file` 
  command-line options.


### Create a short URL

Submit a POST request with Content-Type `application/x-www-form-encoded`
to `/new` with the request body containing:

1. The original, plaintext secret token as `token=<plain token>`.
2. The real URL to redirect to for the new short URL as `url=<redirect-url>`.

Upon success, the server will respond with a JSON blob. For example:

```json
{"short-url":"https://localhost:3000/yuzahnt5"}
```

Visiting the generated URL in your browser will redirect you to the specified
redirect URL. After the configured expiry duration passes (5 minutes at the
time of writing), short URLs are no longer accessible and are removed from the
in-memory data store.