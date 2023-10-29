# Proxy Support

SGPT supports proxying requests through SOCKS and HTTP proxies. This is useful for environments where direct internet
access is not allowed, but proxying is allowed.

You can configure SGPT to use a proxy using the `http_proxy` and `https_proxy` environment variables.

```bash
$ export http_proxy=http://proxy.example.com:8080
$ export https_proxy=http://proxy.example.com:8080
$ sgpt "say hello"
Hello! How can I assist you today?
```

Exceptions can be made for specific hosts by setting the `no_proxy` environment variable.

```bash
$ export no_proxy=example.com
$ sgpt "say hello"
Hello! How can I assist you today?
```
