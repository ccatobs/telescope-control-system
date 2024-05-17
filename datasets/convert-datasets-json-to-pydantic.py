#!/usr/bin/env python3

import json
import re
import sys

def clean_dataset(raw_fields):
    cleaned = []
    seen = {}
    enums = {}
    for raw_field in raw_fields:
        raw_name = raw_field['name']
        raw_type = raw_field['type']

        # check for duplicate fields
        if raw_name in seen:
            seen[raw_name] += 1
            print(f"*** DUPLICATE FIELD {raw_name} ***", file=sys.stderr)
            raw_name += str(seen[raw_name])
        else:
            seen[raw_name] = 1

        field = {
            'raw_name': raw_name,
            'raw_type': raw_type,
        }

        if 'values' in raw_field:
            field['is_enum'] = True
            enums[raw_name] = raw_field['values']

        if 'unit' in raw_field:
            field['unit'] = raw_field['unit']

        if 'comment' in field:
            field['comment'] = raw_field['comment']

        cleaned.append(field)

    return cleaned, enums

def struct_field_name(name):
    # special cases
    if name.lower() == '24v power failure':
        return 'PowerFailure24V'
    elif name.startswith('3rd '):
        name = 'Third ' + name[4:]
    goname = ''.join( w if w[0].isupper() else w.title() for w in name.split() )
    return re.sub(r'[^A-Za-z0-9_]', '', goname)

class Go:

    def convert_name(name):
        return struct_field_name(name)

    def convert_type(type_):
        gotypes = {
            'bool (1 byte)'          : 'bool',
            'LIST (1 byte)'          : 'uint8',
            'FAULT (1 byte)'         : 'uint8',
            'int (4 bytes)'          : 'int32',
            'unsigned int (4 bytes)' : 'uint32',
            'double (8 bytes)'       : 'float64',
            'std::string (32 bytes)' : '[32]byte',
            'std::string (48 bytes)' : '[48]byte',
        }
        return gotypes[type_]

class Pydantic:

    def convert_name(self, name):
        return struct_field_name(name)

    def convert_type(self, type_):
        pydtypes = {
            'bool (1 byte)'          : 'bool',
            'LIST (1 byte)'          : 'int',
            'FAULT (1 byte)'         : 'int',
            'word (2 bytes)'         : 'int',
            'int (4 bytes)'          : 'int',
            'unsigned int (4 bytes)' : 'int',
            'double (8 bytes)'       : 'float',
            'std::string (32 bytes)' : 'str',
            'std::string (48 bytes)' : 'str',
            'std::string (70 bytes)' : 'str',

            'Declination (8 bytes)'  : 'float',
            'RightAscension (8 bytes)' : 'float',
        }
        return pydtypes[type_]

    def make_enum(self, field_name, values):
        enum_name = self.convert_name(field_name) + "Enum"
        result = f"class {enum_name}(enum.IntEnum):\n"
        for k,v in values.items():
            field_name = self.convert_name(k)
            result += f"    {field_name} = {v}\n"
        return result

    def make_struct(self, dataset_name, fields):
        class_name = self.convert_name(dataset_name)
        result = f"class {class_name}(pydantic.BaseModel):\n"
        for field in fields:
            raw_name = field['raw_name']
            name = self.convert_name(raw_name)
            typ_ = self.convert_type(field['raw_type'])
            if 'is_enum' in field:
                typ_ = f"{name}Enum"
            if name != raw_name:
                typ_ = f"pydantic.typing.Annotated[{typ_}, pydantic.Field(alias='{raw_name}')]"
            result += f"    {name}: {typ_}\n"
        result += "    class Config:\n"
        result += "        populate_by_name = True\n"
        return result

def main():
    datasets = json.load(open("datasets.json"))
    lang = Pydantic()
    print("import enum")
    print("import pydantic")
    for dataset_name,dataset_fields in datasets.items():
        cleaned, enums = clean_dataset(dataset_fields)
        for k, v in enums.items():
            print(lang.make_enum(k, v))
        print(lang.make_struct(dataset_name, cleaned))


if __name__ == '__main__':
    main()
