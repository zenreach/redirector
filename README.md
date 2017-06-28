Redirector
==========
This thing is used to redirect HTTP to HTTPS, redirect `example.com` to `www.example.com`, or both.

Running
-------
The `redirector` uses environment variables for configuration. These are:

- `LISTEN` - The `host:port` to listen on.
- `STATUS` - The path to listen for status checks on. Defaults to `_status`.
- `SSL` - Enable/Disable HTTP to HTTPs redirection. Either `true` or `false`.
- `WWW` - Enable/Disable WWW redirection. Either `true` or `false`.

Example: `LISTEN=0.0.0.0:80 SSL=false WWW=true ./redirector`

Response
--------
The `redirector` responds with a 307 if a redirection is necessary. Otherwise a 404 response is returned.
