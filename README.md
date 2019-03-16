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
  "azimuth_range":[$1,$2],
  "elevation":$3,
  "num_scans":2,
  "start_time":"$(date -u --rfc-3339=seconds -d'2 minutes' | tr ' ' T)",
  "turnaround_time":30,
  "speed":0.8
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