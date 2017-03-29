# Onbuild version of the alpine go base image

Adapted from the [golang base image](https://github.com/docker-library/golang/blob/master/1.8/Dockerfile) and the [golang onbuild image](https://github.com/docker-library/golang/blob/master/1.8/onbuild/Dockerfile)

Build with:

```
docker build -t quay.io/coreos/golang:onbuildalpine .
```

Usage:

```
FROM quay.io/coreos/golang:onbuildalpine
ENTRYPOINT ["binary-name"]
CMD []
```

The `ENTRYPOINT` and `CMD` lines are only required if your program doesn't build to a binary called `app`.
