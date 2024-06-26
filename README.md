# FYST Telescope Control System (TCS)

## Building

First, build `liberfa`:
```sh
./build-deps
```

Then:
```sh
go build
```

## Running

```
export FYST_ACU_HOST=10.1.1.1
export FYST_ACU_PORT=8100
./telescope-control-system
```


## Docker

### Building

```sh
docker build --build-arg SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" -t tcs .
```

### Running

```sh
docker run --network host -e FYST_ACU_ADDR=127.0.0.1:8100 tcs
```

## Updating

- Edit `go.mod`, updating the `datasets` library tag version.

- Run `GOPRIVATE='github.com/ccatobs/*' go mod tidy`.

## Commands

### `/abort`

Abort the current command.

```sh
curl -X POST 'http://localhost:5600/abort'
```

### `/acu/failure-reset`

Reset failures. Needed after E-stops.

```sh
curl -X POST 'http://localhost:5600/acu/failure-reset'
```

### `/acu/position-broadcast`

Enable the position broadcast UDP stream, or change where it's sent to.

```sh
curl -X POST 'localhost:5600/acu/position-broadcast' -d@- <<___
{
    "destination_host": "host.name",
    "destination_port": 100000
}
___
```

### `/acu/status`

Get the raw status of the ACU.

```sh
curl 'localhost:5600/acu/status'
```

### `/azimuth-scan`

Scan repeatedly in azimuth, at constant elevation.

```sh
curl 'localhost:5600/azimuth-scan' -d@- <<___
{
  "azimuth_range": [110,130],
  "elevation": 60,
  "num_scans": 20,
  "start_time": 1615586380,
  "turnaround_time": 30,
  "speed": 0.8
}
___
```


### `/move-to`

Move to the specified position.

```sh
curl 'localhost:5600/move-to' -d@- <<___
{
    "azimuth": 120,
    "elevation": 45
}
___
```

### `/path`

Follow a path of points.

```sh
curl 'localhost:5600/path' -d@- <<___
{
    "start_time": 1615586629,
    "coordsys": "ICRS",
    "points": [
        [0,   103, -33, 0.05, -0.05],
        [60,  106, -36, 0.05, -0.05],
        [120, 109, -39, 0.05, -0.05],
        [180, 112, -42, 0.05, -0.05]
    ]
}
___
```

### `/track`

Track a point on the sky.

```sh
curl 'localhost:5600/track' -d@- <<___
{
    "start_time": 1555190103,
    "stop_time": 1555190166,
    "ra": 120,
    "dec": 45,
    "coordsys": "ICRS"
}
___
```
### `/clear-track`

Clear the current program track from telescope

```sh
curl -X POST 'localhost:5600/clear-track"
```


### `/telescope-position`

Get details of telescope position (lat, long, elevation)
