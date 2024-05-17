#!/usr/bin/env python3

import json
import re
import sys

import bs4
import requests

ACU_ADDR = "192.168.1.113:8110"

def get_dataset(name):
    print(f"getting {name}...", file=sys.stderr)
    url = f"http://{ACU_ADDR}/Documentation?identifier=Datasets.{name}"
    r = requests.get(url)
    assert r.status_code == 200

    html = r.text.replace('&nbsp;', ' ')
    soup = bs4.BeautifulSoup(html, features='html.parser')

    fields = []
    for row in soup.table('tr', recursive=False):
        cols = row('td', recursive=False)
        if len(cols) == 0:
            continue

        name = cols[1].p.string
        type_ = cols[2].p.string
        values = cols[3].contents[0]

        field = {'name':name, 'type':type_}

        if values.name == 'p':
            comment = values.string.strip()
            if comment != '':
                match = re.fullmatch(r'Unit: \[(.+)\]', comment)
                if match:
                    field['unit'] = match.group(1)
                else:
                    field['comment'] = comment

        elif values.name == 'table':
            enum = {}
            for vrows in values('tr')[1:]:
                vcols = vrows('td')
                num = int(vcols[0].p.string)
                txt = vcols[1].p.string
                enum[txt] = num
            field['values'] = enum

        else:
            assert False

        fields.append(field)

    return fields

if __name__ == '__main__':
    datasets = { dataset: get_dataset(dataset) for dataset in sys.argv[1:] }
    json.dump(datasets, sys.stdout, indent=2)

