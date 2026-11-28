#!/usr/bin/env python3
"""Extract the open-world zone graph from ao-bin-dumps cluster/world.xml into zones.json.

usage: python3 -I scripts/zones.py world.xml > zones.json
Roads of Avalon and the Mists are left out: their connections change, so no fixed road goes through them.
"""
import json
import sys
import xml.etree.ElementTree as ET


def kind(t):
    if not t:
        return None
    if t.startswith('PLAYERCITY_SAFEAREA') or t in ('PLAYERCITY_BLACK_ROYAL', 'PLAYERCITY_BLACK'):
        return 'city'
    if t in ('SAFEAREA', 'STARTAREA', 'PASSAGE_SAFEAREA'):
        return 'blue'
    if t == 'OPENPVP_YELLOW':
        return 'yellow'
    if t in ('OPENPVP_RED', 'PASSAGE_RED'):
        return 'red'
    if t.startswith('OPENPVP_BLACK') or t == 'PASSAGE_BLACK':
        return 'black'
    return None


clusters = {c.get('id'): c for c in ET.parse(sys.argv[1]).getroot().find('clusters').findall('cluster')}
zones = {cid: {'id': cid, 'name': c.get('displayname'), 'kind': kind(c.get('type'))}
         for cid, c in clusters.items() if kind(c.get('type'))}
edges = set()
for cid in zones:
    for e in clusters[cid].iter('exit'):
        target = (e.get('targetid') or '').partition('@')[2]
        if e.get('targettype') == 'Cluster' and target in zones and target != cid:
            edges.add(tuple(sorted((cid, target))))
json.dump({'zones': sorted(zones.values(), key=lambda z: z['id']), 'edges': sorted(edges)},
          sys.stdout, ensure_ascii=False, separators=(',', ':'))
