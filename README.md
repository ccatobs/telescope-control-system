# CCAT-prime Telescope Control System (TCS)

## Commands

### `/abort`

Abort the current command.

```
curl -X POST 'http://localhost:5600/abort'
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