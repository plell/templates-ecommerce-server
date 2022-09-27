to deploy on server:

    docker-compose up -d

To reset docker cache:

    docker-compose down &&
    docker-compose rm &&
    docker-compose pull &&
    docker-compose build --no-cache &&
    docker-compose up --force-recreate
