# Local Development

## Certificates and Domains

Traefik uses the `certs/prostack.crt` certificate and `certs/prostack.key` key for `prostack.io` and its subdomains. Create the directory and a self-signed certificate with SAN entries for local development:

```sh
mkdir -p certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certs/prostack.key \
  -out certs/prostack.crt \
  -subj "/CN=prostack.io" \
  -addext "subjectAltName=DNS:prostack.io,DNS:*.prostack.io"
```

Add the domains to your system `hosts` file so they resolve to your local machine:

```text
127.0.0.1 prostack.io
127.0.0.1 api.prostack.io
127.0.0.1 grafana.prostack.io
127.0.0.1 traefik.prostack.io
```

The file is located at `/etc/hosts` on macOS and Linux, and at `C:\Windows\System32\drivers\etc\hosts` on Windows. Then, from the `devops` directory, start the infrastructure with `docker compose up -d`. Your browser will warn that the self-signed certificate is not trusted.

To prevent the browser from warning that the certificate is not trusted, it is recommended to add `certs/prostack.crt` to your operating system's trusted certificates. Search for instructions for your operating system.

Kafka runs in single-node KRaft mode at `localhost:9092` with persistent storage in `kafka_data`. Applications run on the host, matching the image's default advertised listener. The user service publishes `user.created` to NATS and Kafka; select the email consumer with `EVENT_BROKER=nats` or `EVENT_BROKER=kafka`.
