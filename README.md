# Serv

A server that easily lets you serve over HTTPS, clone and serve git
repositories, static files, setup redirects, execute commands, and create
reverse proxies, use subdomain. Once installed (`go install
github.com/minond/serv`), create a `Servfile` file that is made up of server
declaration with its paths, handlers, and any information needed to create the
handler. For example:

```
# Handle incoming requests to txtimg.*.*
case Host(txtimg, _, _) =>
  path /             proxy(http://localhost:3002)

# Handle incoming requests to dearme.*.*
case Host(dearme, _, _) =>
  path /             proxy(http://localhost:3003)

# Handle incoming requests to cp.*.*
case Host(cp, _, _) =>
  path /             proxy(http://localhost:3004)

# Handle any incoming request
case Host(_, _, _) =>
  path /             git(https://github.com/minond/minond.github.io.git)
  path /brainfuck    git(https://github.com/minond/brainfuck.git)
  path /brainloller  git(https://github.com/minond/brainloller.git)
  path /servies      git(https://github.com/minond/servies.git)
```

With this configuration, serv will checkout all repositories and serve them
along with serving or proxying anything else you tell it to. Run `serv` in a
directory with your `Servfile` and you're done. Additional options are:

```bash
-certCache string
      Path to Let's Encrypt cache file.
-certDomain string
      Domain whitelist.
-config string
      Path to Servfile file. (default "./Servfile")
-listen string
      Host and port to listen on. (default ":3002")
-pullInterval duration
      Interval git repos are pulled at. (default 15m0s)
```

### Listening on privileged ports

Instead of running server as root (in order to bind to a privileged port, like
80 and 443) one could give the `serv` binary permission to bind to those ports
using the [`setcap`](https://linux.die.net/man/8/setcap) command:

```bash
sudo setcap 'cap_net_bind_service=+ep' $(which serv)
```
