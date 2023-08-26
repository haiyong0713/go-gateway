#!/usr/bin/env python3
import os

gofile = os.environ['GOFILE']
with open(gofile, 'rb') as fp:
    content = fp.read()
    content = content.decode('utf-8')
    if content.find('UpdateBaseInnerDescButton') > 0:
        raise Exception("Failed on `desc_button` overlap check: %s" % gofile)
