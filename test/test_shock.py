
from __future__ import print_function
from os.path import dirname, abspath
import json
import requests

DATADIR = dirname(abspath(__file__)) + "/testdata/"
DEBUG = 1
PORT = "7445"
SHOCK_URL = "http://localhost:"+PORT
TOKEN = "1234"

def test_nodelist_noauth():
    TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {}
    TESTHEADERS = {}
    if DEBUG:
        print(TESTURL, TESTDATA, TESTHEADERS)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200

def test_nodelist_auth():
    TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {}
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    if DEBUG:
        print(TESTURL, TESTDATA, TESTHEADERS)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200

def test_nodelist_badauth():
    TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {}
    TESTHEADERS = {"Authorization": "OAuth BADTOKENREJECTME"}
    if DEBUG:
        print(TESTURL, TESTDATA, TESTHEADERS)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 403 or data["status"] == 400  # 403 unauthorized 400 bad query

def test_upload_emptyfile():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {'upload': open(DATADIR + 'emptyfile', 'rb')}
    if DEBUG:
        print(TESTURL, TESTHEADERS)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    print(response.content.decode("utf-8"))
    assert data["status"] == 200
    assert data["data"]["file"]["checksum"]["md5"] == "d41d8cd98f00b204e9800998ecf8427e"
    # cleanup
    NODEID = data["data"]["id"]
    NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
    response = requests.delete(NODEURL, headers=TESTHEADERS)

def test_upload_threefiles():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {'upload': open(DATADIR + 'AAA.txt', 'rb')}
    if DEBUG:
        print(TESTURL, TESTHEADERS)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    print(response.content.decode("utf-8"))
    assert data["status"] == 200
    NODES = []
    NODES += [data["data"]["id"]]
    FILES = {'upload': open(DATADIR + 'BBB.txt', 'rb')}
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODES += [data["data"]["id"]]
    FILES = {'upload': open(DATADIR + 'CCC.txt', 'rb')}
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODES += [data["data"]["id"]]
    # get node list
    TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
    print(TESTURL)
    TESTDATA = {}
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["total_count"] >= 3
    assert NODES[0] in response.content.decode("utf-8")
    assert b"AAA.txt" in response.content
    assert b"BBB.txt" in response.content
    assert b"CCC.txt" in response.content
    # cleanup
    for NODEID in NODES:
        NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
        response = requests.delete(NODEURL, headers=TESTHEADERS)

def test_upload_and_delete_node():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {'upload': open(DATADIR + 'CCC.txt', 'rb')}
    if DEBUG: print("POST", TESTURL, TESTHEADERS, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODEID = data["data"]["id"]
    # test my node exists
    if DEBUG: print("GET", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {}
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200
   # delete my node
    if DEBUG: print("DELETE", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL+"/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    response = requests.delete(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    # test my node is gone
    if DEBUG: print("GET", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 404

def test_download_zip():
    # upload three files, collect NODEIDS
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {'upload': open(DATADIR + 'AAA.txt', 'rb')}
    if DEBUG:
        print(TESTURL, TESTHEADERS)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODES = []
    NODES += [data["data"]["id"]]
    FILES = {'upload': open(DATADIR + 'BBB.txt', 'rb')}
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODES += [data["data"]["id"]]
    FILES = {'upload': open(DATADIR + 'CCC.txt', 'rb')}
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODES += [data["data"]["id"]]
    #confirm nodes exist
    for NODEID in NODES:
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
        response = requests.get(TESTURL, headers=TESTHEADERS)
        data = json.loads(response.content.decode("utf-8"))
        assert data["status"] == 200
    # query ZIP
    TESTURL = "{SHOCK_URL}/node/querynode".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {"ids": ",".join(NODES)}
    print("GET", TESTURL, TESTDATA)
    response = requests.get(TESTURL, headers=TESTHEADERS, files=FILES, params=TESTDATA)
    print(response.content)
    assert False
    assert data["total_count"] >= 3
    assert NODES[0] in response.content.decode("utf-8")
    # cleanup
    for NODEID in NODES:
        NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
        response = requests.delete(NODEURL, headers=TESTHEADERS)
