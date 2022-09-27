create the traefik container:

    docker network create web
    touch acme.json
    chmod 600 acme.json

    docker run -d \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $PWD/traefik.toml:/traefik.toml \
    -v $PWD/traefik_dynamic.toml:/traefik_dynamic.toml \
    -v $PWD/acme.json:/acme.json \
    -p 80:80 \
    -p 443:443 \
    --network web \
    --name traefik \
    traefik:v2.2

to deploy on server:

    docker-compose up -d

To reset docker cache:

    docker-compose down &&
    docker-compose rm &&
    docker-compose pull &&
    docker-compose build --no-cache &&
    docker-compose up --force-recreate
