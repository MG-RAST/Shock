
from __future__ import print_function
from os.path import dirname, abspath
from subprocess import check_output
import json
import os
import requests
import subprocess

DATADIR = dirname(abspath(__file__)) + "/testdata/"
DEBUG = 1
PORT = os.environ.get('SHOCK_PORT', "7445")
URL  = os.environ.get('SHOCK_HOST', "http://localhost") 
SHOCK_URL = URL + ":" + PORT
TOKEN = os.environ.get("MGRKEY")
if URL == "http://localhost":
    TOKEN = "1234"
    AUTH = "OAuth {}".format(TOKEN)
else:
    AUTH = "mgrast {}".format(TOKEN)

FILELIST = ["AAA.txt", "BBB.txt", "CCC.txt"] 
TESTHEADERS = {"Authorization": AUTH}
DONTDELETE = 1

def create_nodes(FILELIST):
    '''Takes a list of filenames, uploads to shock, returns list of shock ids.'''
    NODES = []
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
# to get multipart-form correctly, data has to be specified in this strange way
# and passed as the files= parameter to requests.put
    FORMDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT"}')}
    for FILE in FILELIST:
        FILES = {'upload': open(DATADIR + FILE, 'rb')}
        if DEBUG:
            print("POST", TESTURL, TESTHEADERS, FILES)
        response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES, data=FORMDATA)
        data = json.loads(response.content.decode("utf-8"))
        assert data["status"] == 200, data["error"]
        assert data["data"]["attributes"]["project_id"] == "TESTPROJECT"
        NODES += [data["data"]["id"]]
        if DEBUG:
            print("PUT", SHOCK_URL + "/node/" + NODES[-1], FORMDATA)
        r = requests.put(SHOCK_URL + "/node/" +
                         NODES[-1], files=FORMDATA, headers=TESTHEADERS)
        if DEBUG:
            print("RESPONSE:", r.content.decode("utf-8"))
        data = json.loads(r.content.decode("utf-8"))
        assert data["data"]["attributes"]["project_id"] == "TESTPROJECT"
    return(NODES)


def confirm_nodes_project(NODES, PROJECT):
    '''Tests a list of nodes to makes sure that attributes->project_id is the same as PROJECT'''
    for NODEID in NODES:
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        if DEBUG:
            print("curl '{}' -H 'Authorization: Oauth {}'".format(TESTURL, TOKEN))
        response = requests.get(TESTURL, headers=TESTHEADERS)
        data = json.loads(response.content.decode("utf-8"))
        assert data["status"] == 200, data["error"]
        assert PROJECT in data["data"]["attributes"]["project_id"]


def delete_nodes(NODELIST):
    '''Delete nodes, confirm http response only'''
    for NODEID in NODELIST:
        NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
        if DEBUG:
            print("DELETE", NODEURL, TESTHEADERS)
        if not DONTDELETE:
            response = requests.delete(NODEURL, headers=TESTHEADERS)
            assert json.loads(response.content.decode("utf-8"))["status"] == 200
    return

def test_delete_nodes():
    NODEID = create_nodes(["AAA.txt"])[0]
    NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
    if DEBUG:
        print("GET", NODEURL, TESTHEADERS)
    predeleteresponse = requests.get(NODEURL, headers=TESTHEADERS)
    assert predeleteresponse.status_code == 200
    assert "Node not found" not in predeleteresponse.content.decode("utf-8")
    delete_nodes([NODEID])
    if DEBUG:
        print("GET", NODEURL, TESTHEADERS)
    postdeleteresponse = requests.get(NODEURL, headers=TESTHEADERS)
    assert postdeleteresponse.status_code == 404
    assert "Node not found" in postdeleteresponse.content.decode("utf-8")


def test_nodelist_noauth():
    TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {}
    TESTHEADERS = {}
    if DEBUG:
        print("GET", TESTURL, TESTDATA, TESTHEADERS)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    assert data["total_count"] >= 0


def test_nodelist_auth():
    TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {}
    if DEBUG:
        print("GET", TESTURL, TESTDATA, TESTHEADERS)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    assert data["total_count"] >= 0


def test_nodelist_badauth():
    TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {}
    TESTHEADERS = {"Authorization": "OAuth BADTOKENREJECTME"}
    if DEBUG:
        print("GET", TESTURL, TESTDATA, TESTHEADERS)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    # 403 unauthorized 400 bad query
    assert data["status"] == 403 or data["status"] == 400


def test_upload_emptyfile():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    FILES = {'upload': open(DATADIR + 'emptyfile', 'rb')}
    if DEBUG:
        print("POST", TESTURL, TESTHEADERS, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES)
    data = json.loads(response.content.decode("utf-8"))
    if DEBUG:
        print("RESPONSE", response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    assert data["data"]["file"]["checksum"]["md5"] == "d41d8cd98f00b204e9800998ecf8427e"
    # cleanup
    NODEID = data["data"]["id"]
    delete_nodes([NODEID])


def test_upload_threefiles():
    NODES = create_nodes(FILELIST)
    TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {}
    if DEBUG:
        print("GET", TESTURL, TESTHEADERS, TESTDATA)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["total_count"] >= 3
    assert NODES[0] in response.content.decode("utf-8")
    assert b"AAA.txt" in response.content
    assert b"BBB.txt" in response.content
    assert b"CCC.txt" in response.content
    # cleanup
    delete_nodes(NODES)


def test_upload_and_delete_node():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
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
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
   # delete my node
    if DEBUG:
        print("DELETE", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL+"/node/{}".format(NODEID)
    response = requests.delete(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    # test my node is gone
    if DEBUG:
        print("GET", TESTURL, TESTHEADERS)
    TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 404


def test_upload_and_download_node_GET():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
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
    FILES = {}
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    DLURL = SHOCK_URL + "/node/{}?download".format(NODEID)
    response = requests.get(DLURL, headers=TESTHEADERS)
    assert response.content[0:3] == b"CCC"
    # cleanup
    delete_nodes([NODEID])


def test_upload_and_download_node_GET_gzip():
    # download file in compressed format, works with all the above options
    # curl -X GET http://<host>[:<port>]/node/<node_id>?download&compression=<zip|gzip>
    # upload node
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
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
    assert data["status"] == 200, data["error"]
    # Download node
    DLURL = SHOCK_URL + "/node/{}?download&compression=gzip".format(NODEID)
    if DEBUG:
        print("GET", DLURL, TESTHEADERS)
    response = requests.get(DLURL, headers=TESTHEADERS)
    assert response.content[0:3] != b"CCC"
    # cleanup
    delete_nodes([NODEID])


def test_upload_and_download_node_GET_zip():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
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
    FILES = {}
    response = requests.get(TESTURL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    DLURL = SHOCK_URL + "/node/{}?download&compression=zip".format(NODEID)
    response = requests.get(DLURL, headers=TESTHEADERS)
    assert response.content[0:3] != b"CCC"
    # cleanup
    delete_nodes([NODEID])

def test_upload_and_download_node_gzip():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
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
    assert data["status"] == 200, data["error"]
    DLURL = SHOCK_URL + "/node/{}?download&compression=gzip".format(NODEID)
    response = requests.get(DLURL, headers=TESTHEADERS)
    assert response.content[0:3] != b"CCC"
    # cleanup
    delete_nodes([NODEID])


def test_download_zip_GET():
    NODES = create_nodes(FILELIST) 
    # confirm nodes exist
    confirm_nodes_project(NODES, "TESTPROJECT")
    # query for TESTDATA
    TESTURL = "{SHOCK_URL}/node?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {"project_id": "TESTPROJECT"}
    if DEBUG:
        print("GET", TESTURL, TESTDATA)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    if DEBUG: 
        print("RESPONSE", response.content)
    data = json.loads(response.content.decode("utf-8"))
    assert data["total_count"] >= 3, "Missing or incorrect total_count" + " ".join([str(response.status_code), str(response.content)])
    assert NODES[0] in response.content.decode("utf-8"), NODES[0] + " not in " + response.content.decode("utf-8")
    # issue query for TESTPROJECT FILES downloaded as ZIP
    TESTURL = SHOCK_URL+"/node?query&download_url&archive=zip".format()
    if DEBUG:
        print("curl '{}' -H 'Authorization: Oauth {}' -G -d {}".format(TESTURL, TOKEN, TESTDATA))
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
    print(" ".join([ "Debugging ZIP Download", str(response.status_code), str(response.content)]))
    data = json.loads(response.content.decode("utf-8"))
    # extract preauth uri from response

    PREAUTH = data["data"]["url"]
    if DEBUG:  
        print("GET", PREAUTH, TESTHEADERS);
    with requests.get(PREAUTH, headers=TESTHEADERS, stream=True) as response:
        # write it to file and test ZIP
        print("Debugging status code: " + str(response.status_code))
        if response.encoding is None:
            response.encoding = 'utf-8'
        # subprocess.run(["ls", "-l"], shell=True) 
        with open("TEST.zip", "wb") as F:
            subprocess.run("ls -l TEST.zip", shell=True)
            for chunk in response.iter_content(chunk_size=512):
                if chunk:
                    F.write(chunk)
            subprocess.run("ls -l TEST.zip", shell=True)
    out = check_output("unzip -l TEST.zip", shell=True)
    assert b'TEST.zip' in out
    assert b'CCC.txt' in out
    assert b'     4 ' in out  # This fails if there are no 4-byte-files
    # cleanup
    delete_nodes(NODES)

def test_download_tar_GET():
    # Per test invokation on https://github.com/MG-RAST/Shock/wiki/API
    # download multiple files in a single archive format (zip or tar), returns 1-time use download url for archive
    # use download_url with a standard query
    # curl -X GET http://<host>[:<port>]/node?query&download_url&archive=zip&<key>=<value>

    NODES = create_nodes(FILELIST)
    # confirm nodes exist
    confirm_nodes_project(NODES, "TESTPROJECT")
    # query for TESTDATA
    TESTURL = "{SHOCK_URL}/node?query".format(SHOCK_URL=SHOCK_URL)
    TESTDATA = {"project_id": "TESTPROJECT"}
    if DEBUG:
        print("GET", TESTURL, TESTDATA)
    response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA)
#    if DEBUG: print("RESPONSE", response.content)
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
    if DEBUG:
        print("GET", PREAUTH, TESTHEADERS)
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
    NODES = create_nodes(FILELIST)
    # confirm nodes exist
    confirm_nodes_project(NODES, "TESTPROJECT")
    # query for TESTDATA
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    # Remember, multipart-forms that are not files have format {key: (None, value)}
    FORMDATA = {"ids": (None, ",".join(NODES)),
                "download_url": (None, 1),
                "archive_format": (None, "tar")}
    # issue query for TESTPROJECT FILES downloaded as TAR
    if DEBUG:
        print("POST", TESTURL, FORMDATA)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FORMDATA)
    data = json.loads(response.content.decode("utf-8"))
    # extract preauth uri from response
    if DEBUG:
        print("RESPONSE", response)
    PREAUTH = data["data"]["url"]
    if DEBUG:
        print("GET", PREAUTH, TESTHEADERS)
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
    NODES = create_nodes(FILELIST)
    # confirm nodes exist
    confirm_nodes_project(NODES, "TESTPROJECT")
    # query for TESTDATA
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    # Remember, multipart-forms that are not files have format {key: (None, value)}
    FORMDATA = {"ids": (None, ",".join(NODES)),
                "download_url": (None, 1),
                "archive_format": (None, "zip")}
    # issue query for TESTPROJECT FILES downloaded as TAR
    if DEBUG:
        print("POST", TESTURL, FORMDATA)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FORMDATA)
    data = json.loads(response.content.decode("utf-8"))
    # extract preauth uri from response
    if DEBUG:
        print("RESPONSE", response)
    PREAUTH = data["data"]["url"]
    if DEBUG:
        print("Debugging receiving : " + PREAUTH)
    response = requests.get(PREAUTH, headers=TESTHEADERS)
    if DEBUG:
        print("Debugging status code: " + str(response.status_code))
    # write it to file and test ZIP
    with open("TESTP.zip", "wb") as f:
        f.write(response.content)
    out = check_output("unzip -l TESTP.zip", shell=True)
    assert b'CCC.txt' in out
    assert b'     4 ' in out  # This fails if there are no 4-byte-files
    # cleanup
    delete_nodes(NODES)

def test_put_attributesstr():
    '''Test PUT request containing attributes_str populates attributes'''
    NODE = create_nodes(["AAA.txt"])[0]
    FORMDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT2"}')}
    if DEBUG:
        print("PUT", SHOCK_URL + "/node/" + NODE, FORMDATA)
    r = requests.put(SHOCK_URL + "/node/" +
                     NODE, files=FORMDATA, headers=TESTHEADERS)
    if DEBUG:
        print("RESPONSE", r.content.decode("utf-8"))
    data = json.loads(r.content.decode("utf-8"))
    if DEBUG:
        print("DATA", data)
    assert data["data"]["attributes"]["project_id"] == "TESTPROJECT2"
    FORMDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT"}')}
    if DEBUG:
        print("PUT", SHOCK_URL + "/node/" + NODE, FORMDATA)
    r = requests.put(SHOCK_URL + "/node/" +
                     NODE, files=FORMDATA, headers=TESTHEADERS)
    if DEBUG:
        print("RESPONSE", r.content.decode("utf-8"))
    data = json.loads(r.content.decode("utf-8"))
    assert data["data"]["attributes"]["project_id"] == "TESTPROJECT"
    delete_nodes([NODE])

def test_post_attributes():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
# to get multipart-form correctly, data has to be specified in this strange way
# and passed as the files= parameter to requests.put
    TESTDATA = {}
    FILES = {'attributes': open(DATADIR + "attr.json", 'rb'),
             'upload': open(DATADIR + "AAA.txt", 'rb')}
    if DEBUG:
            print("POST", TESTURL, TESTDATA, TESTHEADERS, FILES)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES, data=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    NODE = data["data"]["id"]
    assert data["data"]["file"]["name"] == "AAA.txt"
    assert data["data"]["attributes"]["format"] == "replace_format"
    delete_nodes([NODE])

def test_post_gzip():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
# to get multipart-form correctly, data has to be specified in this strange way
# and passed as the files= parameter to requests.put
    TESTDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT"}')}
    FILES = {'gzip': open(DATADIR + "10kb.fna.gz", 'rb')}
    if DEBUG:
            print("POST", TESTURL, TESTDATA, TESTHEADERS)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES, data=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    NODE = data["data"]["id"]
    assert data["data"]["file"]["name"] == "10kb.fna"
    assert data["data"]["file"]["checksum"]["md5"] == "730c276ea1510e2b7ef6b682094dd889"
    delete_nodes([NODE])

def test_post_bzip():
    TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
# to get multipart-form correctly, data has to be specified in this strange way
# and passed as the files= parameter to requests.put
    TESTDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT"}')}
    FILES = {'bzip2': open(DATADIR + "10kb.fna.bz2", 'rb')}
    if DEBUG:
            print("POST", TESTURL, TESTDATA, TESTHEADERS)
    response = requests.post(TESTURL, headers=TESTHEADERS, files=FILES, data=TESTDATA)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    NODE = data["data"]["id"]
    assert data["data"]["file"]["name"] == "10kb.fna"
    assert data["data"]["file"]["checksum"]["md5"] == "730c276ea1510e2b7ef6b682094dd889"
    delete_nodes([NODE])

def test_copynode():
    NODE = create_nodes(["AAA.txt"])[0]
    NODEURL = "{SHOCK_URL}/node/".format(SHOCK_URL=SHOCK_URL)
# to get multipart-form correctly, data has to be specified in this strange way
# and passed as the files= parameter to requests.put
    TESTDATA = {"copy_data": (None, NODE)}
    if DEBUG:
            print("POST", NODEURL, TESTDATA, TESTHEADERS)
    response = requests.post(NODEURL, headers=TESTHEADERS, files=TESTDATA)
    assert response.status_code == 200
    data = json.loads(response.content.decode("utf-8")) 
    print(data)
    NODE2 = data["data"]["id"]
    NODE2URL = "{SHOCK_URL}/node/{NODE2}".format(SHOCK_URL=SHOCK_URL, NODE2=NODE2)
    if DEBUG:
        print("GET", NODE2URL, TESTHEADERS)
    response = requests.get(NODE2URL, headers=TESTHEADERS)
    data = json.loads(response.content.decode("utf-8"))
    assert data["status"] == 200, data["error"]
    assert data["data"]["file"]["checksum"]["md5"] == "8880cd8c1fb402585779766f681b868b" # AAA.txt
    delete_nodes([NODE, NODE2])
