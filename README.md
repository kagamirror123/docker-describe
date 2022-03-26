# docker-psx

This Docker CLI Plugin displays containers by Compose Project.

# Installation

You can download docker-psx binaries from the
[release page](https://github.com/kagamirror123/docker-psx/releases) on this repository.

```bash
mkdir -p ~/.docker/cli-plugins/
wget -q https://github.com/kagamirror123/docker-psx/releases/download/v0.0.2/docker-psx
chmod +x docker-psx
mv docker-psx ~/.docker/cli-plugins/
```

Rename the relevant binary for your OS to `docker-psx` and copy it to `$HOME/.docker/cli-plugins` 


# Using

```bash
docker psx
```

![image](https://user-images.githubusercontent.com/56927897/160223384-d9b9de29-5f71-429c-be0c-bad48bf12534.png)