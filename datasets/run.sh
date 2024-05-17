#!/bin/bash
set -efu -o pipefail

datasets=(
    CmdPointingCorrection
    CmdSectorScanParameters
    CmdSPEMParameter
    CmdStarTrackTransfer
    CmdTimePositionOffsetTransfer
    CmdTimePositionTransfer
    CmdTwoLineTransfer
    CmdWeatherStation
    Status3rdAxis
    StatusDetailedFaults
    StatusExtra8100
    StatusGeneral8100
#    StatusPointingCorrection # not available may 17th 2024
#    StatusSATPDetailed8100 # not available may 17th 2024
)

./get-datasets-json.py "${datasets[@]}" > datasets.json

echo "converting to go..." >&2
./convert-datasets-json-to-go.py datasets.json

echo "formatting..." >&2
go fmt

echo "done" >&2
