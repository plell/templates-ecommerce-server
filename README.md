create the network and volumes:

    docker network create web

create certificate file:

    touch acme.json
    chmod 600 acme.json

create admin password for traefik monitor with htpasswd, then put that password into traefik_dynamic line 3:

    htpasswd -c .htpasswd admin

create the traefik container

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

start docker-compose services:

    docker-compose up -d

To reset docker cache:

    docker-compose down &&
    docker-compose rm &&
    docker-compose pull &&
    docker-compose build --no-cache &&
    docker-compose up --force-recreate

traefik how to:

    https://www.digitalocean.com/community/tutorials/how-to-use-traefik-v2-as-a-reverse-proxy-for-docker-containers-on-ubuntu-20-04

monitoring:

    https://monitor.[domain name]/dashboard/

webhook development:

    stripe listen --forward-to localhost:8000/webhook
