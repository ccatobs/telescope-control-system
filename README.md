# CCAT-prime Telescope Control System (TCS)

## Building

```sh
./build-deps
go build
```
**note to ubuntu users**

Install go from [golang stable release site](https://golang.org/dl/)  rather than the apt-get repo. The latter is out of data and doesn't work

in 1804 If you don't have your git creditial set correctly, go build will fail when connecting to github.com private repos (antenna-control-unit for example), before running go build see the envirnmemt variable GIT_TERMINAL_PROMPT=1 as follows:
```sh
export GIT_TERMINAL_PROMPT=1
```
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
    "start_time": "2019-04-01T20:00:00Z",
    "stop_time": "2019-04-01T21:00:00Z",
    "ra": 120,
    "dec": 45
}
___
```

