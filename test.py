#! /usr/bin/env python
import json
import urllib
import sys

# picking example project
func= "Thr operon leader peptide"
params = urllib.urlencode({'annotation': func, "type" : "function"})
u = urllib.urlopen("http://api.metagenomics.anl.gov/query/?%s" % params)

# u is a file-like object
data = u.read()

# parse json object
response = json.loads(data)

# iterate though response
for mgid in response:
	params = urllib.urlencode({"type" : "functional", "seq" : "dna", "function" : func })
	u = urllib.urlopen("http://api.metagenomics.anl.gov/sequences/%s?%s" % (mgid, params))
	print mgid
	print u.read()
	
