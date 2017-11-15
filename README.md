# Serv

A little server that easily lets you serve git repositories, static files,
redirects, execute commands, and create reverse proxies. Once installed (`go
install github.com/minond/serv`), create a `Servfile` file that is made up of a
path, handler type, and any information needed to create the handler. For
example:

```
/             git        https://github.com/minond/minond.github.io.git
/servies      git        https://github.com/minond/servies.git
/static       dir        .
/github       redirect   https://github.com/minond
/ps           cmd        ps aux
/imdb         proxy      http://www.imdb.com:80
```

With this configuration, serv will checkout all repo and serve them using the
configured paths. Run `serv` in a directory with your `Servfile` and you're
done.
