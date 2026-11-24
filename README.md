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

Your ssh-agent needs to hold a key with read access to the
`antenna-control-unit` repo (check with `ssh-add -l`). The key is
forwarded into the build, and never stored in the image:

```sh
docker build --ssh default -t tcs .
```

### Running

```sh
docker run --network host -e FYST_ACU_ADDR=127.0.0.1:8100 tcs
```

## Updating

- Edit `go.mod`, updating the `datasets` library tag version.

- Run `GOPRIVATE='github.com/ccatobs/*' go mod tidy`.

## Coordinate Systems

The `/path` and `/track` commands accept a `coordsys` field with one of three values:

| `coordsys` | Description |
|---|---|
| `"Horizon"` | Topocentric azimuth and elevation (degrees). Az is N=0, E=90. |
| `"ICRS"` | ICRS right ascension and declination (degrees) at J2000.0. |
| `"Galactic"` | Galactic longitude and latitude (degrees). |

ICRS and Galactic coordinates are converted to Az/El at runtime using [ERFA](https://github.com/liberfa/erfa).
For `/track`, the `ra` and `dec` fields hold the first and second coordinate regardless of system (e.g. Galactic l and b).

## Commands

Command responses are JSON, and include a TCS-generated command `id`:

```json
{"status":"ok","id":14478539199}
{"status":"error","id":14478539200,"message":"commanded azimuth (1000) out of range [-180,360]"}
```

IDs are assigned to every recognized `POST` command, including rejected ones,
and also appear in the TCS log. They increase monotonically, and are seeded
from the startup time (in milliseconds) so they stay unique across restarts.

Commands which take parameters may include an optional `tags` object of string
key/value pairs, which are logged along with the command ID. There may be at
most 16 tags, with non-empty keys of at most 64 bytes, and values of at most 256
bytes.

```sh
curl 'localhost:5600/move-to' -d '{"azimuth": 120, "elevation": 45, "tags": {"scan": "s42"}}'
```

### `/abort`

Abort the current command.

```sh
curl -X POST 'http://localhost:5600/abort'
```

### `/reset`

Abort the current command (if any), stop the telescope, and reset ACU failures.
Use after an E-stop, or to get the telescope back to a known state.

```sh
curl -X POST 'localhost:5600/reset'
```

### `/azimuth-scan`

Scan repeatedly in azimuth, at constant elevation.
The telescope first moves to the start of the scan, (`azimuth_range[0]`,`elevation`),
and waits there until shortly before `start_time`.
`start_time` is optional; if omitted (or 0) the scan starts ~10 seconds after
the move finishes. If given, it must leave enough time for the move, plus 5 seconds.

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


### `/holography-scan`

Scan `num_lines` straight lines through (`azimuth`,`elevation`), each from
`-radius` to `+radius` at constant `speed`. Lines come in perpendicular pairs,
and each pair is rotated `180/num_lines` degrees from the last, e.g. for
`num_lines=50` the lines are at 0°, 90°, 3.6°, 93.6°, ..., 86.4°, 176.4°
(measured from the +azimuth direction). `num_lines` must be even.
The ACU generates the turnarounds between lines, which take `turnaround_time` seconds.
The telescope first moves to the center, then starts the scan there, at rest, at
`start_time`; it reaches the first line `turnaround_time` seconds later.
At the end it returns to the center, also over `turnaround_time` seconds.
The TCS samples this lead-in & lead-out every 0.1 seconds, using the same
constant-jerk profile as the turnarounds in `scan_50_turns_scans.dat`.
`start_time` is optional; if omitted (or 0) the scan starts ~10 seconds after
the move finishes. If given, it must leave enough time for the move, plus 5 seconds.
At the end it turns back to the center and stops.
Offsets are in raw azimuth & elevation, i.e. not scaled by cos(elevation).

```sh
curl 'localhost:5600/holography-scan' -d@- <<___
{
  "start_time": 1615586380,
  "azimuth": 120,
  "elevation": 45,
  "radius": 1,
  "num_lines": 50,
  "speed": 0.285714,
  "turnaround_time": 3
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

### `/path/clear`

Clear the current program track from the telescope.

```sh
curl -X POST 'localhost:5600/path/clear'
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

### `/status`

Get the currently running command, or `null` if idle.
`started_at` is the unixtime when the TCS began executing the command.
For `/path` commands, `params` has `num_points` in place of the full `points` list.

```sh
curl 'localhost:5600/status'
```

```json
{"status":"ok","command":{"id":14478539201,"endpoint":"/move-to","params":{"azimuth":120,"elevation":45},"tags":{"scan":"s42"},"started_at":1790171340.12}}
{"status":"ok","command":null}
```

### `/telescope-position`

Get details of telescope position (lat, long, elevation)

## ACU Commands

Lower-level commands which talk to the ACU directly.

### `/acu/datasets`

List the names of the available ACU datasets.

```sh
curl 'localhost:5600/acu/datasets'
```

```json
["BusStations","CmdPointingCorrection",...,"SunAvoidance"]
```

### `/acu/datasets/<name>`

Get an ACU dataset as JSON, with field names and coded values as the ACU reports them.
Undefined (NaN) values are `null`.

```sh
curl 'localhost:5600/acu/datasets/Shutter'
```

### `/acu/values/<identifier>`

Get the actual values of an ACU module as JSON, as returned by the ACU's
`/Values?identifier=<identifier>&type=Actual&format=JSON`.

```sh
curl 'localhost:5600/acu/values/Antenna.Corrections.Tiltmeter+Az'
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

### `/acu/reboot`

Reboot the ACU.

```sh
curl -X POST 'http://localhost:5600/acu/reboot'
```

### `/acu/spem`

Set parameters of the ACU's systematic pointing error model (SPEM).
Any subset of the parameters may be given; the rest are left unchanged.
Units are mdeg, except `LEL_velocity` which is deg/s.

Parameters: `IA`, `IE`, `TFC`, `TF`, `TFS`, `TFEC`, `TFE`, `TFES`, `AN`, `AW`,
`AN2`, `AW2`, `CA`, `NPAE`, `AES`, `AEC`, `AES2`, `AEC2`, `EES`, `EEC`, `EES2`,
`EEC2`, `NRX`, `NRY`, `LELA`, `LEVA`, `LELE`, `LEVE`, `LEL_velocity`.

```sh
curl 'localhost:5600/acu/spem' -d@- <<___
{
    "IA": 1.5,
    "IE": -0.3,
    "CA": 0.02
}
___
```

The current values can be read with `/acu/datasets/CmdSPEMParameter`.

### `/acu/status`

Get the raw status of the ACU.

```sh
curl 'localhost:5600/acu/status'
```

### `/acu/stow`

Insert the stow pins, without moving the telescope.
Aborts the running command, if any.
Done once the elevation axis is in its stow position, the pins are inserted, and both axes have stopped.

```sh
curl -X POST 'localhost:5600/acu/stow'
```

### `/acu/survival-mode`

Move to the survival position and insert the stow pins.
Aborts the running command, if any.
Done once both axes have reached their commanded positions, the pins are inserted, and both axes have stopped.

```sh
curl -X POST 'localhost:5600/acu/survival-mode'
```

### `/acu/unstow`

Retract the stow pins.
Done once the pins are retracted and both axes have stopped.

```sh
curl -X POST 'localhost:5600/acu/unstow'
```
