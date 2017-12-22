# Serv

A server that easily lets you serve over HTTPS, clone and serve git
repositories, static files, setup redirects, execute commands, and create
reverse proxies. Once installed (`go install github.com/minond/serv`), create a
`Servfile` file that is made up of a path, handler type, and any information
needed to create the handler. For example:

```
# Path        Type       Endpoint information
/             git        https://github.com/minond/minond.github.io.git
/servies      git        https://github.com/minond/servies.git
/static       dir        .
/github       redirect   https://github.com/minond
/ps           cmd        ps aux
/imdb         proxy      http://www.imdb.com:80
/unibrow      proxy      http://localhost:3001
```

With this configuration, serv will checkout all repositories and serve them
along with serving or proxying anything else you tell it to. Run `serv` in a
directory with your `Servfile` and you're done. Additional options are:

```bash
-config string
      Path to Servfile file. (default "./Servfile")
-listen string
      Host and port to listen on. (default ":3002")
-listenHttps string
      Path to Let's Encript cache file instead of host/port.
-pullInterval duration
      Interval git repos are pulled at. (default 15m0s)
```
