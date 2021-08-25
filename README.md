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
git clone https://github.com/connext/vector-router-docker-compose.git
```

2. Rename file `.env.example` to `.env` and modify it. You need to set next environment variables:

- `ROUTER_VERSION` - version to use, images found here: https://github.com/connext/nxtp/pkgs/container/nxtp-router
- `LOGDNA_KEY` - set LogDNA Ignestion key
- `LOGDNA_TAG` - optionally set LogDNA tag
- `ROUTER_EXTERNAL_PORT`, `GRAFANA_EXTERNAL_PORT` - modify ports for external access

3. (Optional) Modify `.env` file and set alert notifications to Slack or Discord.

For Discord set:

- `DISCORD_WEBHOOK` - Discord webhook full url

Modify `docker-compose.yml` file and uncomment (remove #) for all `alertmanager-discord` section.

Note: for Discord notifications used two containers `alertmanager` and `alertmanager-discord`

4. Create NXTP configuration file `~/vector-router-docker-compose/config.json`, it will be mounted into router container. See [Connext docs](https://docs.connext.network/Routers/Reference/configuration/) for configuration description.

5. Create docker-compose services, volumes and network.

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

## CI

Update `latestVersion` file in the repo to automatically update production router.
