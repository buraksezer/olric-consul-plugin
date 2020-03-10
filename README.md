# olric-consul-plugin

Consul integration for service discovery

## Build

Get the code:

```
go get -u github.com/buraksezer/olric-consul-plugin
```

With a properly configured Go environment:

```
go build -buildmode=plugin -o consul.so 
```

In order to strip debug symbols:

```
go build -ldflags="-s -w" -buildmode=plugin -o consul.so 
```

## Configuration

If you prefer to deploy Olric in client-server scenario, add a `serviceDiscovery` block to your `olricd.yaml`. A sample:

```yaml

serviceDiscovery:
  provider: "consul"
  path: "YOUR_PLUGIN_PATH/consul.so"
  url: "http://127.0.0.1:8500"
  passingOnly: true
  replaceExistingChecks: false
  insecureSkipVerify: true
  payload: '
      {
          "Name": "olric-cluster",
          "ID": "olric-node-1",
          "Tags": [
            "primary",
            "v1"
          ],
          "Address": "localhost",
          "Port": MEMBERLIST_PORT,
          "EnableTagOverride": false,
          "check": {
            "name": "Olric node on MEMBERLIST_PORT",
            "tcp": "0.0.0.0:MEMBERLIST_PORT",
            "interval": "10s",
            "timeout": "1s"
          }
      }
'
```

In embedded member deployment scenario:

```go
// import config package
"github.com/buraksezer/olric/config"

// Get a new Olric config for local environment
c := config.New("local")

// Set service discovery definition
sd := make(map[string]interface{})
sd["provider"] = "consul"
sd["path"] = "YOUR_PLUGIN_PATH/consul.so"
sd["address"] = "http://127.0.0.1:8500"
sd["passingOnly"] = true
sd["payload"] = `{
  "Name": "olric-cluster",
  "ID": "olric-node-1",
  "Tags": [
    "primary",
    "v1"
  ],
  "Address": "localhost",
  "Port": MEMBERLIST_PORT,
  "EnableTagOverride": false,
  "check": {
    "name": "Check Olric node on MEMBERLIST_PORT",
    "tcp": "0.0.0.0:MEMBERLIST_PORT",
    "interval": "1s",
    "timeout": "1s"
  }
}`

c.ServiceDiscovery = sd
```

Or you can load the plugin directly as a library:

```go
import (
    olricconsul "github.com/buraksezer/olric-consul-plugin/lib"
)

//...
sd := make(map[string]interface{})
sd["plugin"] = &olricconsul.ConsulDiscovery{}
//...
```

## Contributions

Please don't hesitate to fork the project and send a pull request or just e-mail me to ask questions and share ideas.

## License

The Apache License, Version 2.0 - see LICENSE for more details.
