# Router Docker Compose

Production-ready docker-compose for Connext routers.

## Router Setup Using docker-compose

### Requirements

- [ Docker CE (Community Edition) ](https://docs.docker.com/install/) version 20.10.5 or higher
- [ Docker Compose ](https://docs.docker.com/compose/install/) version 1.27.4 or higher

### Run docker-compose Stack

1. Clone repo

```
cd ~
git clone https://github.com/connext/router-docker-compose.git
cd ~/router-docker-compose
```

2. Rename file `.env.example` to `.env` and modify it. You need to set the following environment variables:

- `ROUTER_VERSION` - version to use, get router version from either https://github.com/connext/nxtp/pkgs/container/router-publisher or `#amarok-routers` channel in [Discord](https://discord.gg/connext)
- `GRAFANA_PASSWORD` - your password for Grafana
- `LOGDNA_KEY` - LogDNA key, you need to register in [Mezmo App](https://app.mezmo.com/account/signin) and get it from there. You can optionally set `LOGDNA_TAG` as well
- `USERID` - if you use docker rootless, then it should be ID of your user e.g. 1001.
- `*_VERSION` - you can optionally set docker container versions of other apps

3. Modify `data/alertmanagerConfig/alertmanager.yml` file and set alert notifications to Mail, Slack, Discord, Telegram, Pagerduty, Opsgenie, etc. Additional configuration might be required.

4. Create NXTP configuration file `~/router-docker-compose/config.json`, it will be mounted into router container. See [Connext Configuration docs](https://docs.connext.network/routers/Reference/configuration) for configuration description. You can use `config.example.mainnet.json` or `config.example.testnet.json` as an example.

5. (Optional) Create external [Redis](https://redis.io/) instance and insert URL into `redisUrl` in config. (currently the docker-compose file includes redis container as well). If you want to use highly available RabbitMQ service - you can spin up it as well and update `config.json` as well.

6. (Optional) Follow security best practices from [Connext Security docs](https://docs.connext.network/routers/Guides/security)

7. (Optional) Use docker rootless configuration for better security. Just use `docker-compose-rootless.yml` instead of `docker-compose.yml` and follow [Docker Rootless Guide](https://docs.docker.com/engine/security/rootless/) to enable it.

8. (Optional) Edit `docker-compose.yml` to enable port forwarding for Router services and/or Prometheus/Alertmanager. It's disabled by default for security reasons. Do NOT expose these port to public networks, otherwise use proven authentication methods.

9. Rename file `key.example.yaml` to `key.yaml` and modify it. Web3Signer yaml key file `~/router-docker-compose/key.yaml` will be mounted into the signer container. This example file uses raw unencrypted files method. See [Web3Signer docs](https://docs.web3signer.consensys.net/en/latest/HowTo/Use-Signing-Keys/).
And for more custom commands of web3signer, edit `~/router-docker-compose/data/signerConfig/config.yaml`. Refer [Web3Signer Command docs](https://docs.web3signer.consensys.net/en/latest/Reference/CLI/CLI-Syntax/)



10. Create docker-compose services, volumes and network.

```
docker-compose create
```

11. Run docker-compose stack.

```
docker-compose up -d
```

12. Check the status.

```
docker-compose ps
OR
docker ps -a
```

13. Check logs to ensure router started successfully

```
docker-compose logs router-publisher | tail -n 200| grep 'Router publisher boot complete!'
docker-compose logs router-subscriber | tail -n 200| grep 'Router subscriber boot complete!'
```

14. Check the full logs if needed

```
docker-compose logs
OR
docker-compose logs router-publisher
docker-compose logs router-subscriber
```

You can also use these commands.

```
docker logs router-publisher
docker logs router-subscriber
```


## Other Tasks

### Stop and delete containers.

```
docker-compose down
```

### Delete data

```
docker-compose down -v
```

### Restart Stack

```
docker-compose restart
```

### Update Version

1. Modify `.env` to change `ROUTER_VERSION`
2. Update stack

```
docker-compose pull
docker-compose up -d
```


### Infrastructure model

![Infrastructure model](.infragenie/infrastructure_model.png)
