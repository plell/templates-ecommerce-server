to deploy on server:

    docker-compose up -d

docker troubleshooting:

    docker-compose rm &&
    docker-compose pull &&
    docker-compose build --no-cache &&
    docker-compose up -d --force-recreate
