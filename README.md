# olric-consul-plugin

Consul integration for service discovery

## Build

With a properly configured Go environment:

```
go build -buildmode=plugin -o consul.so 
```

In order to strip debug symbols:

```
go build -ldflags="-s -w" -buildmode=plugin -o consul.so 
```

## Configuration

In `olricd.yaml`, add something like the following:

```yaml

serviceDiscovery:
  provider: "consul"
  path: "YOUR_PLUGIN_PATH/consul.so"
  url: "http://127.0.0.1:8500"
  passingOnly: true
  replaceExistingChecks: true
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
          "Port": 3322,
          "EnableTagOverride": false,
          "check": {
            "name": "Olric node on 3322",
            "tcp": "0.0.0.0:3322",
            "interval": "10s",
            "timeout": "1s"
          }
      }
'
```

```json
{
  "Name": "olric-cluster",
  "ID": "olric-node-1",
  "Tags": [
    "primary",
    "v1"
  ],
  "Address": "localhost",
  "Port": 3322,
  "EnableTagOverride": false,
  "check": {
    "name": "Check Olric node on 3322",
    "tcp": "0.0.0.0:3322",
    "interval": "10s",
    "timeout": "1s"
  }
}
```

## Contributions

Please don't hesitate to fork the project and send a pull request or just e-mail me to ask questions and share ideas.

## License

The Apache License, Version 2.0 - see LICENSE for more details.
