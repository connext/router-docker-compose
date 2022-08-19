# NXTP Router Docker Compose

Production-ready docker-compose for NXTP routers.

## Router Setup Using docker-compose

### Requirements

- [ Docker CE (Community Edition) ](https://docs.docker.com/install/) version 20.10.5 or higher
- [ Docker Compose ](https://docs.docker.com/compose/install/) version 1.27.4 or higher

### Run docker-compose Stack

1. Clone repo

```
cd ~
git clone https://github.com/connext/nxtp-router-docker-compose.git
```

2. Rename file `.env.example` to `.env` and modify it. You need to set next environment variables:

- `ROUTER_VERSION` - version to use, images found here: https://github.com/connext/nxtp/pkgs/container/nxtp-router
- `LOGDNA_KEY` - set LogDNA Ignestion key
- `LOGDNA_TAG` - optionally set LogDNA tag
- `ROUTER_EXTERNAL_PORT`, `GRAFANA_EXTERNAL_PORT`

3. (Optional) Modify `.env` file and set alert notifications to Slack or Discord.

4. Create NXTP configuration file `~/nxtp-router-docker-compose/config.json`, it will be mounted into router container. See [Connext docs](https://docs.connext.network/Routers/Reference/configuration/) for configuration description.

5. Create external [Redis](https://redis.io/) instance and insert URL into `redisUrl` in config.


6. Rename file `key.example.yaml` to `key.yaml` and modify it. Web3Signer yaml key file `~/nxtp-router-docker-compose/key.yaml` will be mounted into the signer container. Example file use raw unencrypted files method. See [Web3Signer docs](https://docs.web3signer.consensys.net/en/latest/HowTo/Use-Signing-Keys/). 
And for more custom commands of web3signer, edit `~/nxtp-router-docker-compose/data/signerConfig/config.yaml`. Refer [Web3Signer Command docs](https://docs.web3signer.consensys.net/en/latest/Reference/CLI/CLI-Syntax/)


7. Create docker-compose services, volumes and network.

```
cd ~/nxtp-router-docker-compose
docker-compose create
```

6. Run docker-compose stack.

```
docker-compose up -d
```

7. Check the status.

```
docker-compose ps
OR
docker ps -a
```

8. Check the logs.

```
docker-compose logs
OR
docker-compose logs router
```

You can also use these commands.

```
docker logs router
```

9. Stop and delete containers.

```
docker-compose down
```

## Other Tasks

### Restart Stack

```
docker-compose restart
```

### Update Version

1. Modify `.env` to change `NXTP_VERSION`
2. Update stack

```
docker-compose pull
docker-compose up -d
```


### Infrastructure model

![Infrastructure model](.infragenie/infrastructure_model.png)