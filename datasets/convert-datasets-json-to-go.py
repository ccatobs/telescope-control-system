#!/usr/bin/env python3

import json
import re
import sys

import bs4
import requests

def print_struct(name, fields):
    print()
    print(f"type {name} struct {{")

    seen = {}
    enums = {}
    for field in fields:
        name = field['name']
        type_ = field['type']

        gname = go_name(name)
        gtype = go_type(type_)

        if gname in seen:
            seen[gname] += 1
            print(f"*** DUPLICATE FIELD {gname} ***", file=sys.stderr)
            gname += str(seen[gname])
        else:
            seen[gname] = 1

        if 'values' in field:
            enum = {}
            for key,val in field['values'].items():
                enum[gname + go_name(key)] = val
            enums[gname] = enum

        extras = f'`json:"{name}"`'
        if 'unit' in field:
            extras += f" // Unit: [{field['unit']}]"
        if 'comment' in field:
            extras += f" // {field['comment']}"

        print(gname, gtype, extras)

    print("}")

    return enums

def go_name(name):
    # special cases
    if name.lower() == '24v power failure':
        return 'PowerFailure24V'
    elif name.startswith('3rd '):
        name = 'Third ' + name[4:]
    goname = ''.join( w if w[0].isupper() else w.title() for w in name.split() )
    return re.sub(r'[^A-Za-z0-9]', '', goname)

def go_type(type_):
    gotypes = {
        'bool (1 byte)'          : 'bool',
        'LIST (1 byte)'          : 'uint8',
        'FAULT (1 byte)'         : 'uint8',
	'word (2 bytes)'         : 'uint16',
        'int (4 bytes)'          : 'int32',
        'unsigned int (4 bytes)' : 'uint32',
        'double (8 bytes)'       : 'float64',
        'std::string (32 bytes)' : '[32]byte',
        'std::string (48 bytes)' : '[48]byte',
        'std::string (70 bytes)' : '[70]byte',

	'Declination (8 bytes)'  : 'float64',
	'RightAscension (8 bytes)' : 'float64',
    }
    return gotypes[type_]

def new_go_file(filename):
    sys.stdout = open(filename, "w")
    print("package datasets")

if __name__ == '__main__':
    datasets = json.load(open(sys.argv[1]))

    all_enums = {}
    for dataset,fields in datasets.items():
        new_go_file(f"{dataset}.go")
        enums = print_struct(dataset, fields)
        # dedup
        for k,v in enums.items():
            if k in all_enums:
                assert all_enums[k] == v
            else:
                all_enums[k] = v

    new_go_file(f"constants.go")
    for enum in all_enums.values():
        print()
        print("const (")
        for k,v in enum.items():
            print(f"{k} = {v}")
        print(")")

