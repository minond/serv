# Serv

A little server that easily lets you serve git repositories right now and with
more options to come. Once installed (`go install github.com/minond/serv`),
create a `Servfile` file that is made up of a path, the type of resource, and
information about the resource. For example:

```
/             git        https://github.com/minond/minond.github.io.git
/servies      git        https://github.com/minond/servies.git
/brainfuck    git        https://github.com/minond/brainfuck.git
/brainloller  git        https://github.com/minond/brainloller.git
/static       directory  .
```

With this configuration, serv will checkout all repo and serve them using the
configured paths. Run `serv` in a directory with your `Servfile` and you're
done.
