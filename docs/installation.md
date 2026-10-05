# Installation and operation

Contiwatch needs Docker Engine and access to its Docker socket. Docker Compose is required for the Compose installation below; source image builds also need Docker BuildKit. The published images support `linux/amd64`, `linux/arm/v7`, and `linux/arm64`.

## Controller with Docker Compose

From a repository checkout, edit [docker-compose.yml](../docker-compose.yml) and replace the `APP_PIN` placeholder with a unique, non-trivial PIN containing 4–8 digits. The controller refuses to start without a valid PIN. Keep the `/data` volume: it stores configuration and managed stack files.

The supplied Compose file publishes port 8080 on the host. Restrict access to a trusted private network or configure HTTPS through a reverse proxy. Docker socket access allows Contiwatch to manage containers on the host; give access only to trusted administrators. See [security](security.md).

```bash
docker compose up -d
```

Open `http://localhost:8080` and unlock the UI with your PIN. The Compose file uses the published image; it has no local `build` section.

To stop the controller while retaining the named data volume:

```bash
docker compose down
```

## Build and run a local image

Run these commands from the repository root. Replace the PIN placeholder before starting the container. Apply the same network restrictions as for Compose.

```bash
docker build -t contiwatch .
docker run -d \
  --name contiwatch \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v contiwatch-data:/data \
  -e APP_PIN="SET_A_UNIQUE_4_TO_8_DIGIT_PIN" \
  contiwatch
```

The Dockerfile builds the CodeMirror bundle and Go binary, then includes Docker CLI and Compose in the runtime image. For local development prerequisites and checks, see [testing](testing.md).

## Remote agent

On the remote Docker host, edit [compose_agent.yml](../compose_agent.yml). Set `CONTIWATCH_AGENT=true` and replace `CONTIWATCH_AGENT_TOKEN` with a unique token of at least 32 random characters. Do not configure `APP_PIN` on agents. Use HTTPS or restrict the agent endpoint to a trusted private network.

```bash
docker compose -f compose_agent.yml up -d
```

Alternatively, run the published image directly, after replacing the token placeholder:

```bash
docker run -d \
  --name contiwatch-agent \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v contiwatch-agent-data:/data \
  -e CONTIWATCH_AGENT=true \
  -e CONTIWATCH_AGENT_TOKEN="PUT_LONG_RANDOM_TOKEN_HERE" \
  ghcr.io/pbuzdygan/contiwatch:latest
```

Add the agent's name, URL, and matching token in the controller's Servers view. Agents expose an authenticated API rather than the interactive UI. See [configuration](configuration.md#servers) and [API authentication](api.md#authentication-and-agent-availability).

## Published images and upgrades

Stable images:

```bash
docker pull ghcr.io/pbuzdygan/contiwatch:latest
```

Development images:

```bash
docker pull ghcr.io/pbuzdygan/contiwatch:dev_latest
```

For a fixed version, use the exact release tag for stable images or `dev_<release-tag>` for development images. See [release checks and image tagging](release-check.md) for the mapping.

Before upgrading, back up configuration and stack files, and read the [changelog](../CHANGELOG.md). For the Compose controller, retain its volume and configured PIN, then run:

```bash
docker compose pull
docker compose up -d
```

For an agent, use the same commands with `-f compose_agent.yml`. Keep its token unchanged unless intentionally rotating it. Remote stack timeout extension needs agents running `1.3.4` or newer; older agents retain the synchronous compatibility path.

## Permissions and troubleshooting

The image entrypoint defaults to `PUID=1000` and `PGID=1000`, adjusts `/data` ownership, and runs the application as that user. If configuration writes fail, set these variables to your host user IDs, obtained with `id`:

```yaml
environment:
  PUID: "${PUID}"
  PGID: "${PGID}"
```

Define the interpolation values in your shell or deployment `.env` before using this example.

If Docker access fails with `permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock`, check that the socket is mounted and the application user belongs to its group. Contiwatch detects the socket group automatically; set `DOCKER_GID` explicitly when needed, using the host socket group ID from `stat -c '%g' /var/run/docker.sock`.

For a controller startup failure, check that `APP_PIN` contains 4–8 digits rather than the example placeholder. An agent requires both its mode flag and a non-empty token. See the [environment reference](configuration.md#environment-variables) for defaults and compatibility aliases.
