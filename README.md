# CCAT-prime Telescope Control System (TCS)

## Building

```sh
./build-deps
go build
```
### Note to Ubuntu 18.04 users

Install go from the [golang stable release site](https://golang.org/dl/) rather than apt-get,
as the latter is too out-of-date.

If you don't have your git credentials set correctly,
`go build` will fail because it needs access to a private github repository.
Try setting the environment variable `GIT_TERMINAL_PROMPT=1`.

## Commands

### `/abort`

Abort the current command.

```sh
curl -X POST 'http://localhost:5600/abort'
```

### `/acu-status`

Get the raw status of the ACU.

```sh
curl 'localhost:5600/acu-status'
```

### `/azimuth-scan`

Scan repeatedly in azimuth, at constant elevation.

```sh
curl 'localhost:5600/azimuth-scan' -d@- <<___
{
  "azimuth_range": [110,130],
  "elevation": 60,
  "num_scans": 20,
  "start_time": "2019-03-16T20:56:30Z",
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
    "coordsys": "ICRS",
    "points": [
        [1555190103, 103, -33],
        [1555190166, 106, -35],
        [1555190228, 109, -37],
        [1555190288, 111, -39]
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

### `/telescope-position`

Get details of telescope position (lat, long, elevation)
