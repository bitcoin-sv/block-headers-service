# Block Headers Service

-----------------------------------------------------

## How to use this image

-----------------------------------------------------

### starting new instance

`docker run ${DOCKER_IMAGE}:${VERSION}`

### run with volume

`docker run -v bhs-data:/app/data ${DOCKER_IMAGE}:${VERSION}`

### run with preloaded database

You can load prepared database containing 750k headers already imported.
To use it run the docker with `--preloaded` argument:

`docker run -v bhs-data:/app/data ${DOCKER_IMAGE}:${VERSION} --preloaded`

### clean start

If you use docker volume the data is persisted between runs.
If you would like to run application in a clean environment with the same volume as previously,
you can start it with `--clean` argument:

`docker run -v bhs-data:/app/data ${DOCKER_IMAGE}:${VERSION} --clean`

### clean preloaded start

If you use docker volume the data is persisted between runs.
If you have already some data in database, but would like to run in preloaded mode
and don't want to recreate a volume you can run the image providing both `--clean` and `--preloaded` arguments

`docker run -v bhs-data:/app/data ${DOCKER_IMAGE}:${VERSION} --clean --preloaded`
