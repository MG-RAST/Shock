
from __future__ import print_function
from os.path import dirname, abspath
from subprocess import check_output
import json
import os
import requests

DATADIR = dirname(abspath(__file__)) + "/testdata/"
DEBUG = 1
PORT = os.environ.get('SHOCK_PORT' , "7445")
URL  = os.environ.get('SHOCK_HOST' , "http://localhost") 
SHOCK_URL = URL + ":" + PORT
TOKEN = "1234"


def create_three_nodes():
    NODES = []
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
# to get multipart-form correctly, data has to be specified in this strange way
# and passed as the files= parameter to requests.put
    TESTDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT"}')}
    FILES = {'upload': open(DATADIR + 'AAA.txt', 'rb')}
    if DEBUG:
        print("POST", TESTURL, TESTDATA, TESTHEADERS)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    if DEBUG:
        print(response.status_code)
        print(response.text)
        print(response.error)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200
    NODES += [data["data"]["id"]]
    if DEBUG:
        print("PUT", SHOCK_URL + "/node/" + NODES[-1], TESTDATA)
    r = requests.put(SHOCK_URL + "/node/" +
                     NODES[-1], files=TESTDATA, headers=TESTHEADERS)
    FILES = {'upload': open(DATADIR + 'BBB.txt', 'rb')}
    if DEBUG:
        print("POST", TESTURL, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODES += [data["data"]["id"]]
    if DEBUG:
        print("PUT", SHOCK_URL + "/node/" + NODES[-1], TESTDATA)
    r = requests.put(SHOCK_URL + "/node/" +
                     NODES[-1], files=TESTDATA, headers=TESTHEADERS)
    FILES = {'upload': open(DATADIR + 'CCC.txt', 'rb')}
    if DEBUG:
        print("POST", TESTURL, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODES += [data["data"]["id"]]
    if DEBUG:
        print("PUT", SHOCK_URL + "/node/" + NODES[-1], TESTDATA)
    r = requests.put(SHOCK_URL + "/node/" +
                     NODES[-1], files=TESTDATA, headers=TESTHEADERS)
    print(r.content.decode("utf-8"))
    data = json.loads(r.content.decode("utf-8"))
    assert data["data"]["attributes"]["project_id"] == "TESTPROJECT"
    return(NODES)


def confirm_nodes_project(NODES, PROJECT):
    for NODEID in NODES:
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
        if DEBUG:
            print("curl '{}' -H 'Authorization: Oauth {}'".format(TESTURL, TOKEN))
        response = requests.get(TESTURL, headers=TESTHEADERS)
        data = json.loads(response.content.decode("utf-8"))
        assert data["status"] == 200
        assert PROJECT in data["data"]["attributes"]["project_id"]


def delete_nodes(NODELIST):
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    for NODEID in NODELIST:
        NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
        response = requests.delete(NODEURL, headers=TESTHEADERS)
        assert json.loads(response.content.decode("utf-8"))["status"] == 200
    return


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
    # 403 unauthorized 400 bad query
    assert data["status"] == 403 or data["status"] == 400


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
    NODES = create_three_nodes()
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
    if DEBUG:
        print("POST", TESTURL, TESTHEADERS, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODEID = data["data"]["id"]
    # test my node exists
    if DEBUG:
        print("GET", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {}
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200
   # delete my node
    if DEBUG:
        print("DELETE", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL+"/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    response = requests.delete(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    # test my node is gone
    if DEBUG:
        print("GET", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 404


def test_upload_and_download_node_GET():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {'upload': open(DATADIR + 'CCC.txt', 'rb')}
    if DEBUG:
        print("POST", TESTURL, TESTHEADERS, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODEID = data["data"]["id"]
    # test my node exists
    if DEBUG:
        print("GET", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {}
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200
    DLURL = SHOCK_URL + "/node/{}?download".format(NODEID)
    response = requests.get(DLURL, headers=TESTHEADERS)
    assert response.content[0:3] == b"CCC"
    # cleanup
    NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
    requests.delete(NODEURL, headers=TESTHEADERS)


def test_upload_and_download_node_GET_gzip():
    # download file in compressed format, works with all the above options
    # curl -X GET http://<host>[:<port>]/node/<node_id>?download&compression=<zip|gzip>
    # upload node
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {'upload': open(DATADIR + 'CCC.txt', 'rb')}
    if DEBUG:
        print("POST", TESTURL, TESTHEADERS, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODEID = data["data"]["id"]
    # test my node exists
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {}
    if DEBUG:
        print("GET", TESTURL, TESTHEADERS)
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200
    # Download node
    DLURL = SHOCK_URL + "/node/{}?download&compression=gzip".format(NODEID)
    response = requests.get(DLURL, headers=TESTHEADERS)
    assert response.content[0:3] != b"CCC"
    # cleanup
    NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
    requests.delete(NODEURL, headers=TESTHEADERS)


def test_upload_and_download_node_GET_zip():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {'upload': open(DATADIR + 'CCC.txt', 'rb')}
    if DEBUG:
        print("POST", TESTURL, TESTHEADERS, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODEID = data["data"]["id"]
    # test my node exists
    if DEBUG:
        print("GET", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {}
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200
    DLURL = SHOCK_URL + "/node/{}?download&compression=zip".format(NODEID)
    response = requests.get(DLURL, headers=TESTHEADERS)
    assert response.content[0:3] != b"CCC"
    # cleanup
    NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
    requests.delete(NODEURL, headers=TESTHEADERS)


def test_upload_and_download_node_gzip():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    FILES = {'upload': open(DATADIR + 'CCC.txt', 'rb')}
    if DEBUG:
        print("POST", TESTURL, TESTHEADERS, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    NODEID = data["data"]["id"]
    # test my node exists
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    FILES = {}
    if DEBUG:
        print("GET", TESTURL, TESTHEADERS)
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200
    DLURL = SHOCK_URL + "/node/{}?download&compression=gzip".format(NODEID)
    response = requests.get(DLURL, headers=TESTHEADERS)
    assert response.content[0:3] != b"CCC"
    # cleanup
    NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
    requests.delete(NODEURL, headers=TESTHEADERS)


def test_download_zip_GET():
    NODES = create_three_nodes()
    # confirm nodes exist
    confirm_nodes_project(NODES, "TESTPROJECT")
    # query for TESTDATA
    TESTURL = "{SHOCK_URL}/node?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {"project_id": "TESTPROJECT"}
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    if DEBUG:
        print("GET", TESTURL, TESTDATA)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
#    if DEBUG: print(response.content)
    data = json.loads(response.content.decode("utf-8"))
    assert data["total_count"] >= 3
    assert NODES[0] in response.content.decode("utf-8")
    # issue query for TESTPROJECT FILES downloaded as ZIP
    TESTURL = SHOCK_URL+"/node?query&download_url&archive=zip".format()
    if DEBUG:
        print("curl '{}' -H 'Authorization: Oauth {}' -G -d {}".format(TESTURL, TOKEN, TESTDATA))
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    # extract preauth uri from response
    PREAUTH = data["data"]["url"]
    response = requests.get(PREAUTH, headers=TESTHEADERS)
    # write it to file and test ZIP
    with open("TEST.zip", "wb") as F:
        F.write(response.content)
    out = check_output("unzip -l TEST.zip", shell=True)
    assert b'TEST.zip' in out
    assert b'CCC.txt' in out
    assert b'     4 ' in out  # This fails if there are no 4-byte-files
    # cleanup
    for NODEID in NODES:
        NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
        response = requests.delete(NODEURL, headers=TESTHEADERS)


def test_download_tar_GET():
    # Per test invokation on https://github.com/MG-RAST/Shock/wiki/API
    # download multiple files in a single archive format (zip or tar), returns 1-time use download url for archive
    # use download_url with a standard query
    # curl -X GET http://<host>[:<port>]/node?query&download_url&archive=zip&<key>=<value>

    NODES = create_three_nodes()
    # confirm nodes exist
    confirm_nodes_project(NODES, "TESTPROJECT")
    # query for TESTDATA
    TESTURL = "{SHOCK_URL}/node?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {"project_id": "TESTPROJECT"}
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    if DEBUG:
        print("GET", TESTURL, TESTDATA)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
#    if DEBUG: print(response.content)
    data = json.loads(response.content.decode("utf-8"))
    assert data["total_count"] >= 3
    assert NODES[0] in response.content.decode("utf-8")
    # issue query for TESTPROJECT FILES downloaded as ZIP
    TESTURL = SHOCK_URL+"/node?query&download_url&archive=tar".format()
    if DEBUG:
        print("curl '{}' -H 'Authorization: Oauth {}' -G -d {}".format(TESTURL, TOKEN, TESTDATA))
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    # extract preauth uri from response
    PREAUTH = data["data"]["url"]
    response = requests.get(PREAUTH, headers=TESTHEADERS)
    # write it to file and test ZIP
    with open("TEST.tar", "wb") as f:
        f.write(response.content)
    out = check_output("tar tvf TEST.tar", shell=True)
    assert b'CCC.txt' in out
    assert b'     4 ' in out  # This fails if there are no 4-byte-files
    # cleanup
    delete_nodes(NODES)


def test_download_tar_POST():
    NODES = create_three_nodes()
    # confirm nodes exist
    confirm_nodes_project(NODES, "TESTPROJECT")
    # query for TESTDATA
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    # Remember, multipart-forms that are not files have format { key: (None, value) }
    TESTDATA = {"ids": (None, ",".join(NODES)),
                "download_url": (None, 1),
                "archive_format": (None, "tar")}
    # issue query for TESTPROJECT FILES downloaded as TAR
    if DEBUG:
        print("POST", TESTURL, TESTDATA)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    # extract preauth uri from response
    if DEBUG:
        print(response)
    PREAUTH = data["data"]["url"]
    response = requests.get(PREAUTH, headers=TESTHEADERS)
    # write it to file and test ZIP
    with open("TESTP.tar", "wb") as f:
        f.write(response.content)
    out = check_output("tar tvf TESTP.tar", shell=True)
    assert b'CCC.txt' in out
    assert b'     4 ' in out  # This fails if there are no 4-byte-files
    # cleanup
    delete_nodes(NODES)


def test_download_zip_POST():
    # Per test invokation on https://github.com/MG-RAST/Shock/wiki/API
    # use download_url with a POST and list of node ids
    # curl -X POST -F "download_url=1" -F "archive_format=zip" -F "ids=<node_id_1>,<node_id_2>,<...>" http://<host>[:<port>]/node
    NODES = create_three_nodes()
    # confirm nodes exist
    confirm_nodes_project(NODES, "TESTPROJECT")
    # query for TESTDATA
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    TESTHEADERS = {"Authorization": "OAuth {}".format(TOKEN)}
    # Remember, multipart-forms that are not files have format { key: (None, value) }
    TESTDATA = {"ids": (None, ",".join(NODES)),
                "download_url": (None, 1),
                "archive_format": (None, "zip")}
    # issue query for TESTPROJECT FILES downloaded as TAR
    if DEBUG:
        print("POST", TESTURL, TESTDATA)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    # extract preauth uri from response
    if DEBUG:
        print(response)
    PREAUTH = data["data"]["url"]
    response = requests.get(PREAUTH, headers=TESTHEADERS)
    # write it to file and test ZIP
    with open("TESTP.zip", "wb") as f:
        f.write(response.content)
    out = check_output("unzip -l TESTP.zip", shell=True)
    assert b'CCC.txt' in out
    assert b'     4 ' in out  # This fails if there are no 4-byte-files
    # cleanup
    delete_nodes(NODES)
