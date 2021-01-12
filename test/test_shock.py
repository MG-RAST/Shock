
from __future__ import print_function
from os.path import dirname, abspath
from subprocess import check_output
import json
import os
import requests
import subprocess
import pytest
from pprint import pprint




DATADIR = "testdata"
DEBUG = 0
SHOCK_URL = ""
SHOCK_USER_AUTH = ""
SHOCK_ADMIN_AUTH = ""

AUTH = ""
FILELIST = []
TESTHEADERS = {}
DONTDELETE = 0

def GETJSON(URL, headers={}, params={}, checkcode=True):
        response = requests.get(URL, headers=headers, params=params)
        if DEBUG:
            print ("GET", URL, headers, params)
        if DEBUG:
            print("curl '{}' -H '{}' -G -d {}".format(URL, headers, params))
        if checkcode:
            assert response.status_code == 200, contents
        contents =  response.content.decode("utf-8")
        data = json.loads(contents) 
        assert data is not None, contents
        return(data, response)

def POSTJSON(URL, headers={}, files={}, data={}):
        response = requests.post(URL, headers=headers, files=files, data=data)
        if DEBUG:
            print ("POST", URL, headers, files, data)
        assert response.status_code == 200, contents
        contents =  response.content.decode("utf-8")
        data = json.loads(contents) 
        assert data is not None, contents
        return(data, response)

class TestClass:
   
    
    
    @pytest.fixture(scope="session", autouse=True)
    def execute_before_any_test(self):
        """ setup any state specific to the execution of the given class (which
        usually contains tests).
        """
       
        
        
        
        
        print("execute_before_any_test started ----------------------------")
        # DATADIR = dirname(abspath(__file__)) + "/testdata/"
        global DEBUG
        DEBUG = 1
        #PORT = os.environ.get('SHOCK_PORT', "7445")
        #URL  = os.environ.get('SHOCK_HOST', "http://localhost")
        #SHOCK_URL = URL + ":" + PORT
        global SHOCK_URL
        SHOCK_URL  = os.environ.get('SHOCK_URL', "http://shock:7445")

        #TOKEN = os.environ.get("MGRKEY")

        # SHOCK_AUTH="bearer token"
        global SHOCK_AUTH

	# default AUTH is USER AUTH
        global AUTH
        global SHOCK_USER_AUTH

        
        global FILELIST
        FILELIST = ["AAA.txt", "BBB.txt", "CCC.txt"]
        global LONEFILEAAA
        LONEFILEAAA = os.path.join(DATADIR, "AAA.txt")
        global LONEFILECCC
        LONEFILECCC = os.path.join(DATADIR, "CCC.txt")
        # SHOCK_USER_AUTH="bearer token"
        
        SHOCK_USER_AUTH = os.environ.get("SHOCK_USER_AUTH", "basic dXNlcjE6c2VjcmV0")
        SHOCK_ADMIN_AUTH = os.environ.get("SHOCK_ADMIN_AUTH", "basic YWRtaW46c2VjcmV0")

        AUTH=SHOCK_USER_AUTH

        global TESTHEADERS
        TESTHEADERS = {"Authorization": SHOCK_USER_AUTH}
        global TESTAHEADERS
        TESTAHEADERS = {"Authorization": SHOCK_ADMIN_AUTH}


        #if URL == "https://sequencing.bio.anl.gov":
        #    TESTHEADERS= {"AUTH" : TOKEN}
        global DONTDELETE
        DONTDELETE = 0
        
        return


    def create_nodes(self, FILELIST):
        '''Takes a list of filenames, uploads to shock, returns list of shock ids.'''
        NODES = []
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    # to get multipart-form correctly, data has to be specified in this strange way
    # and passed as the files= parameter to requests.put
        FORMDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT"}')}
        for FILE in FILELIST:
            
            if not FILE.startswith(DATADIR):
                FILE=os.path.join(DATADIR, FILE)
            
            FILES = {'upload': open(FILE, 'rb')}
            data, response = POSTJSON(TESTURL, headers=TESTHEADERS, files=FILES, data=FORMDATA)
            assert "data" in data, response.content.decode("utf-8")
            assert data["data"] is not None, response.content.decode("utf-8")
            assert "attributes" in data["data"], data
            assert data["data"]["attributes"] is not None , data
            assert data["status"] == 200, data["error"]
            assert data["data"]["attributes"]["project_id"] == "TESTPROJECT"
            NODES += [data["data"]["id"]]
            if DEBUG:
                print("PUT", SHOCK_URL + "/node/" + NODES[-1], FORMDATA)
            r = requests.put(TESTURL + "/" + NODES[-1],
                                 files=FORMDATA,
                                 headers=TESTHEADERS)
            if DEBUG:
                print("RESPONSE:", r.content.decode("utf-8"))
            data = json.loads(r.content.decode("utf-8"))
            assert data is not None, response.content.decode("utf-8")
            assert data["data"] is not None, response.content.decode("utf-8")
            assert data["data"]["attributes"]["project_id"] == "TESTPROJECT"
        return(NODES)


    def confirm_nodes_project(self, NODES, PROJECT):
        '''Tests a list of nodes to makes sure that attributes->project_id is the same as PROJECT'''
        for NODEID in NODES:
            TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
            data, response = GETJSON(TESTURL, TESTHEADERS, {})
            assert PROJECT in data["data"]["attributes"]["project_id"]


    def delete_nodes(self, NODELIST):
        '''Delete nodes, confirm http response only'''
        for NODEID in NODELIST:
            NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
            if DEBUG:
                print("DELETE", NODEURL, TESTHEADERS)
            if not DONTDELETE:
                response = requests.delete(NODEURL, headers=TESTHEADERS)
                assert json.loads(response.content.decode("utf-8"))["status"] == 200
        return

    def test_delete_nodes(self):
        assert DONTDELETE is not 1, "This test fails unless deleting is enabled"
        NODEID = self.create_nodes([LONEFILEAAA])[0]
        NODEURL = SHOCK_URL + "/node/{}".format(NODEID)
        data, predeleteresponse = GETJSON(NODEURL, TESTHEADERS, {})
        assert "Node not found" not in predeleteresponse.content.decode("utf-8")
        self.delete_nodes([NODEID])
        data, postdeleteresponse = GETJSON(NODEURL, TESTHEADERS, {},  checkcode=False)
        assert postdeleteresponse.status_code == 404
        assert "Node not found" in postdeleteresponse.content.decode("utf-8")


    def test_nodelist_noauth(self):
        TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
        TESTDATA = {}
        TESTHEADERS = {}
        data, response = GETJSON(TESTURL, TESTHEADERS, TESTDATA, {})
        assert data["total_count"] >= 0


    def test_nodelist_auth(self):
        TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
        TESTDATA = {}
        data, response = GETJSON(TESTURL, TESTHEADERS, TESTDATA)
        assert data["total_count"] >= 0


    def test_nodelist_badauth(self):
        TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
        TESTDATA = {}
        TESTHEADERS = {"Authorization": "OAuth BADTOKENREJECTME"}
        data, response = GETJSON(TESTURL, TESTHEADERS, TESTDATA, checkcode=False)
        assert response.status_code == 403 or response.status_code == 400, response.content.decode("utf-8")

        # 403 unauthorized 400 bad query
        assert data["status"] == 403 or data["status"] == 400


    def test_upload_emptyfile(self):
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
        FILES = {'upload': open(os.path.join(DATADIR, 'emptyfile'), 'rb')}
        data, response = POSTJSON(TESTURL, headers=TESTHEADERS, files=FILES)
        assert data["data"]["file"]["checksum"]["md5"] == "d41d8cd98f00b204e9800998ecf8427e"
        # cleanup
        NODEID = data["data"]["id"]
        self.delete_nodes([NODEID])


    def test_upload_threefiles(self):
        NODES = self.create_nodes(FILELIST)
        TESTURL = "{SHOCK_URL}/node/?query".format(SHOCK_URL=SHOCK_URL)
        TESTDATA = {}
        data, response = GETJSON(TESTURL, TESTHEADERS, TESTDATA)
        assert data["total_count"] >= 3
        assert NODES[0] in response.content.decode("utf-8")
        assert b"AAA.txt" in response.content
        assert b"BBB.txt" in response.content
        assert b"CCC.txt" in response.content
        # cleanup
        self.delete_nodes(NODES)


    def test_upload_and_delete_node(self):
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
        NODEID = self.create_nodes([LONEFILECCC])[0]
        # test my node exists
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        data, response = GETJSON(TESTURL, TESTHEADERS, {})
       # delete my node
        if DEBUG:
            print("DELETE", TESTURL, TESTHEADERS)
        TESTURL = SHOCK_URL+"/node/{}".format(NODEID)
        response = requests.delete(TESTURL, headers=TESTHEADERS)
        data = json.loads(response.content.decode("utf-8"))
        assert data is not None, response.content.decode("utf-8")

        # test my node is gone
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        data, response = GETJSON(TESTURL, TESTHEADERS, checkcode=False)
        assert response.status_code == 404, response.content.decode("utf-8")
        assert data["status"] == 404


    def test_upload_and_download_node_GET(self):
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
        NODEID = self.create_nodes([LONEFILECCC])[0]

        # test my node exists
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        FILES = {}
        data, response = GETJSON(TESTURL, TESTHEADERS, {})
        assert response.status_code == 200, response.content.decode("utf-8")
        DLURL = SHOCK_URL + "/node/{}?download".format(NODEID)
        response = requests.get(DLURL, headers=TESTHEADERS)
        assert response.content[0:3] == b"CCC"
        # cleanup
        self.delete_nodes([NODEID])

    def test_upload_and_download_node_GET_file_name(self):
        NODEID = self.create_nodes([LONEFILECCC])[0]
        # test my node exists
        DLURL = SHOCK_URL + "/node/{}?download&file_name=TESTFILE-CCC.txt".format(NODEID)
        FILES = {}
        response = requests.get(DLURL, headers=TESTHEADERS)
        print("HEADERS", response.headers)
        assert 'Content-Disposition' in response.headers
        assert response.headers['Content-Disposition'] == 'attachment; filename=TESTFILE-CCC.txt'
        # cleanup
        self.delete_nodes([NODEID])


    def test_upload_and_download_node_GET_gzip(self):
        # download file in compressed format, works with all the above options
        # curl -X GET http://<host>[:<port>]/node/<node_id>?download&compression=<zip|gzip>
        # upload node
        NODEID = self.create_nodes([LONEFILECCC])[0]
        # test my node exists
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        FILES = {}
        data, response = GETJSON(TESTURL, TESTHEADERS)
        # Download node
        DLURL = SHOCK_URL + "/node/{}?download&compression=gzip".format(NODEID)
        response = requests.get(DLURL, headers=TESTHEADERS)
        assert response.content[0:3] != b"CCC"
        # cleanup
        self.delete_nodes([NODEID])


    def test_upload_and_download_node_GET_zip(self):
        NODEID = self.create_nodes([LONEFILECCC])[0]
        # test my node exists
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        FILES = {}
        data, response = GETJSON(TESTURL, TESTHEADERS)
        DLURL = SHOCK_URL + "/node/{}?download&compression=zip".format(NODEID)
        response = requests.get(DLURL, headers=TESTHEADERS)
        assert response.content[0:3] != b"CCC"
        # cleanup
        self.delete_nodes([NODEID])

    def test_upload_and_download_node_gzip(self):
        NODEID = self.create_nodes([LONEFILECCC])[0]
        # test my node exists
        TESTURL = SHOCK_URL + "/node/{}".format(NODEID)
        FILES = {}
        data, response = GETJSON(TESTURL, TESTHEADERS)
        DLURL = SHOCK_URL + "/node/{}?download&compression=gzip".format(NODEID)
        response = requests.get(DLURL, headers=TESTHEADERS)
        assert response.content[0:3] != b"CCC"
        # cleanup
        self.delete_nodes([NODEID])


    def test_download_url_zip_GET(self):
        NODES = self.create_nodes(FILELIST)
        # confirm nodes exist
        self.confirm_nodes_project(NODES, "TESTPROJECT")
        # query for TESTDATA
        assert SHOCK_URL != None
        print("SHOCK_URL:"+SHOCK_URL)
        TESTURL = "{SHOCK_URL}/node?query".format(SHOCK_URL=SHOCK_URL)
        TESTDATA = {"project_id": "TESTPROJECT"}
        data, response = GETJSON(TESTURL, TESTHEADERS, params=TESTDATA)
        assert "total_count" in data
        assert data["total_count"] >= 3, "Missing or incorrect total_count" + " ".join([str(response.status_code), str(response.content)])
        assert NODES[0] in response.content.decode("utf-8"), NODES[0] + " not in " + response.content.decode("utf-8")
        # issue query for TESTPROJECT FILES downloaded as ZIP
        TESTURL = SHOCK_URL+"/node?query&download_url&archive=zip"
        data, response = GETJSON(TESTURL, headers=TESTHEADERS, params=TESTDATA)
        print(" ".join([ "Debugging ZIP Download", str(response.status_code), str(response.content)]))
        assert "data" in data, response.content.decode("utf-8")
        # extract preauth uri from response
        PREAUTH_URL = data["data"]["url"] # example: http://localhost/preauth/TbqTUadG42vVf72LkWRg 
        TESTURL=PREAUTH_URL
        if DEBUG:
            print("GET", TESTURL, TESTHEADERS);
        with requests.get(TESTURL, headers=TESTHEADERS, stream=True) as response:
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
        os.unlink("TEST.zip")
        # cleanup
        self.delete_nodes(NODES)

    def test_download_url_GET(self):
        NODES = self.create_nodes(FILELIST)
        # confirm nodes exist
        self.confirm_nodes_project(NODES, "TESTPROJECT")
        # query for TESTDATA
        TESTDATA = {}
        assert SHOCK_URL != None
        # construct download_url query for only the first node 
        TESTURL = SHOCK_URL+"/node/{}?download_url".format(NODES[0])
        data, response = GETJSON(TESTURL, TESTHEADERS, params=TESTDATA)
        assert "data" in data, response.content.decode("utf-8")
        # extract preauth uri from response
        PREAUTH_URL = data["data"]["url"] # example: http://localhost/preauth/TbqTUadG42vVf72LkWRg 
        TESTURL=PREAUTH_URL
        if DEBUG:
            print("GET", TESTURL, TESTHEADERS);
        with requests.get(TESTURL, headers=TESTHEADERS, stream=True) as response:
            # write it to file and test ZIP
            print("Debugging status code: " + str(response.status_code))
            if response.encoding is None:
                response.encoding = 'utf-8'
            # subprocess.run(["ls", "-l"], shell=True)
            with open("TEST3.out", "wb") as F:
                for chunk in response.iter_content(chunk_size=512):
                    if chunk:
                        F.write(chunk)
        downloadedfile = open("TEST3.out").read()
        assert "AAA" in downloadedfile, downloadedfile        
        # NO TESTS 
        # cleanup
        os.unlink("TEST3.out")
        self.delete_nodes(NODES)

    def test_download_url_GET_file_name(self):
        NODES = self.create_nodes(FILELIST)
        # confirm nodes exist
        self.confirm_nodes_project(NODES, "TESTPROJECT")
        # query for TESTDATA
        TESTDATA = {}
        assert SHOCK_URL != None
        # construct download_url query for only the first node 
        TESTURL = SHOCK_URL+"/node/{}?download_url&file_name=TESTFILEAAA.txt".format(NODES[0])
        data, response = GETJSON(TESTURL, TESTHEADERS, params=TESTDATA)
        assert "data" in data, response.content.decode("utf-8")
        # extract preauth uri from response
        PREAUTH_URL = data["data"]["url"] # example: http://localhost/preauth/TbqTUadG42vVf72LkWRg 
        TESTURL=PREAUTH_URL
        if DEBUG:
            print("GET", TESTURL, TESTHEADERS);
        with requests.get(TESTURL, headers=TESTHEADERS, stream=True) as response:
            # write it to file and test ZIP
            print("Debugging status code: " + str(response.status_code))
            if response.encoding is None:
                response.encoding = 'utf-8'
            # Don't bother downloading file
        print("HEADERS", response.headers)
        assert 'Content-Disposition' in response.headers
        assert response.headers['Content-Disposition'] == 'attachment; filename=TESTFILEAAA.txt'
        # cleanup
        self.delete_nodes(NODES)

    def test_download_url_tar_GET(self):
        # Per test invokation on https://github.com/MG-RAST/Shock/wiki/API
        # download multiple files in a single archive format (zip or tar), returns 1-time use download url for archive
        # use download_url with a standard query
        # curl -X GET http://<host>[:<port>]/node?query&download_url&archive=zip&<key>=<value>

        NODES = self.create_nodes(FILELIST)
        # confirm nodes exist
        self.confirm_nodes_project(NODES, "TESTPROJECT")
        # query for TESTDATA
        TESTURL = "{SHOCK_URL}/node?query".format(SHOCK_URL=SHOCK_URL)
        TESTDATA = {"project_id": "TESTPROJECT"}
        data, response = GETJSON(TESTURL, TESTHEADERS, params=TESTDATA)
        if DEBUG:
            print("RESPONSE 1 :", response.content)
    #    if DEBUG: print("RESPONSE", response.content)
        assert "total_count" in data, response.content.decode("utf-8")
        assert data["total_count"] >= 3
        assert NODES[0] in response.content.decode("utf-8")
        # issue query for TESTPROJECT FILES downloaded as ZIP
        TESTURL = SHOCK_URL+"/node?query&download_url&archive=tar".format()
        response = requests.get(TESTURL, headers=TESTHEADERS, params=TESTDATA) # binary
        if DEBUG:
            print("RESPONSE 2 :", response.content)
        data = json.loads(response.content.decode("utf-8"))
        assert data is not None, response.content.decode("utf-8")
        # extract preauth uri from response
        assert "data" in data
        assert "url" in data["data"]
        PREAUTH = data["data"]["url"]
        if DEBUG:
            print("GET", PREAUTH, TESTHEADERS)
        response = requests.get(PREAUTH, headers=TESTHEADERS)
        if DEBUG:
            print("RESPONSE 3 :", response.content)
        # write it to file and test ZIP
        with open("TEST.tar", "wb") as f:
            f.write(response.content)
        out = check_output("tar tvf TEST.tar", shell=True)
        assert b'CCC.txt' in out
        assert b'     4 ' in out  # This fails if there are no 4-byte-files
        # cleanup
        os.unlink("TEST.tar")
        self.delete_nodes(NODES)


    def test_download_url_tar_POST(self):
        NODES = self.create_nodes(FILELIST)
        # confirm nodes exist
        self.confirm_nodes_project(NODES, "TESTPROJECT")
        # query for TESTDATA
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
        # Remember, multipart-forms that are not files have format {key: (None, value)}
        FORMDATA = {"ids": (None, ",".join(NODES)),
                    "download_url": (None, 1),
                    "archive_format": (None, "tar")}
        # issue query for TESTPROJECT FILES downloaded as TAR
        data, response = POSTJSON(TESTURL, headers=TESTHEADERS, files=FORMDATA)
        # extract preauth uri from response
        PREAUTH = data["data"]["url"]
        if DEBUG:
            print("GET", PREAUTH, TESTHEADERS)
        response = requests.get(PREAUTH, headers=TESTHEADERS)
        if DEBUG:
            print("RESPONSE 2 :", response.content)
        # write it to file and test ZIP
        with open("TESTP.tar", "wb") as f:
            f.write(response.content)
        out = check_output("tar tvf TESTP.tar", shell=True)
        assert b'CCC.txt' in out
        assert b'     4 ' in out  # This fails if there are no 4-byte-files
        # cleanup
        os.unlink("TESTP.tar")
        self.delete_nodes(NODES)


    def test_download_url_zip_POST(self):
        # Per test invokation on https://github.com/MG-RAST/Shock/wiki/API
        # use download_url with a POST and list of node ids
        # curl -X POST -F "download_url=1" -F "archive_format=zip" -F "ids=<node_id_1>,<node_id_2>,<...>" http://<host>[:<port>]/node
        NODES = self.create_nodes(FILELIST)
        # confirm nodes exist
        self.confirm_nodes_project(NODES, "TESTPROJECT")
        # query for TESTDATA
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
        # Remember, multipart-forms that are not files have format {key: (None, value)}
        FORMDATA = {"ids": (None, ",".join(NODES)),
                    "download_url": (None, 1),
                    "archive_format": (None, "zip")}
        # issue query for TESTPROJECT FILES downloaded as TAR
        data, response = POSTJSON(TESTURL, headers=TESTHEADERS, files=FORMDATA)

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
        os.unlink("TESTP.zip")
        self.delete_nodes(NODES)

    def test_put_attributesstr(self):
        '''Test PUT request containing attributes_str populates attributes'''
        NODE = self.create_nodes([LONEFILEAAA])[0]
        FORMDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT2"}')}
        if DEBUG:
            print("PUT", SHOCK_URL + "/node/" + NODE, FORMDATA)
        r = requests.put(SHOCK_URL + "/node/" +
                         NODE, files=FORMDATA, headers=TESTHEADERS)
        if DEBUG:
            print("RESPONSE", r.content.decode("utf-8"))
        data = json.loads(r.content.decode("utf-8"))
        assert data is not None, r.content.decode("utf-8")
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
        assert data is not None, r.content.decode("utf-8")
        assert data["data"]["attributes"]["project_id"] == "TESTPROJECT"
        self.delete_nodes([NODE])

    def test_post_attributes(self):
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    # to get multipart-form correctly, data has to be specified in this strange way
    # and passed as the files= parameter to requests.put
        TESTDATA = {}
        FILES = {'attributes': open(os.path.join(DATADIR, "attr.json"), 'rb'),
                 'upload': open(LONEFILEAAA, 'rb')}
        data, response = POSTJSON(TESTURL, headers=TESTHEADERS, files=FILES, data=TESTDATA)
        NODE = data["data"]["id"]
        assert data["data"]["file"]["name"] == "AAA.txt"
        assert data["data"]["attributes"]["format"] == "replace_format"
        self.delete_nodes([NODE])

    def test_post_gzip(self):
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    # to get multipart-form correctly, data has to be specified in this strange way
    # and passed as the files= parameter to requests.put
        TESTDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT"}')}
        FILES = {'gzip': open(os.path.join(DATADIR, "10kb.fna.gz"), 'rb')}
        data, response = POSTJSON(TESTURL, headers=TESTHEADERS, files=FILES, data=TESTDATA)
        NODE = data["data"]["id"]
        assert data["data"]["file"]["name"] == "10kb.fna"
        assert data["data"]["file"]["checksum"]["md5"] == "730c276ea1510e2b7ef6b682094dd889"
        self.delete_nodes([NODE])

    def test_post_bzip(self):
        TESTURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    # to get multipart-form correctly, data has to be specified in this strange way
    # and passed as the files= parameter to requests.put
        TESTDATA = {"attributes_str": (None, '{"project_id":"TESTPROJECT"}')}
        FILES = {'bzip2': open(os.path.join(DATADIR, "10kb.fna.bz2"), 'rb')}
        data, response = POSTJSON(TESTURL, headers=TESTHEADERS, files=FILES, data=TESTDATA)
        NODE = data["data"]["id"]
        assert data["data"]["file"]["name"] == "10kb.fna"
        assert data["data"]["file"]["checksum"]["md5"] == "730c276ea1510e2b7ef6b682094dd889"
        self.delete_nodes([NODE])

    def test_copynode(self):
        NODE = self.create_nodes([LONEFILEAAA])[0]
        NODEURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
    # to get multipart-form correctly, data has to be specified in this strange way
    # and passed as the files= parameter to requests.put
        TESTDATA = {"copy_data": (None, NODE)}
        data, response = POSTJSON(NODEURL, headers=TESTHEADERS, files=TESTDATA)
        NODE2 = data["data"]["id"]
        NODE2URL = "{SHOCK_URL}/node/{NODE2}".format(SHOCK_URL=SHOCK_URL, NODE2=NODE2)
        data, response = GETJSON(NODE2URL, headers=TESTHEADERS)
        assert data["status"] == 200, data["error"]
        assert data["data"]["file"]["checksum"]["md5"] == "8880cd8c1fb402585779766f681b868b" # AAA.txt
        self.delete_nodes([NODE, NODE2])

    def test_querynode_md5(self):
        NODE = self.create_nodes([LONEFILEAAA])[0]
        NODEURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
        PARAMS = {"querynode": "1", "file.checksum.md5": "8880cd8c1fb402585779766f681b868b"}
        data, response = GETJSON(NODEURL, headers=TESTHEADERS, params=PARAMS)
        self.delete_nodes([NODE])
        assert "total_count" in data.keys(), data
        assert data["total_count"] > 0, data
        assert data["data"][0]["file"]["checksum"]["md5"] == "8880cd8c1fb402585779766f681b868b"

    def test_querynode_name(self):
        NODE = self.create_nodes([LONEFILEAAA])[0]
        NODEURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)
        PARAMS = {"querynode": "1", "file.name": "AAA.txt"}
        data, response = GETJSON(NODEURL, headers=TESTHEADERS, params=PARAMS)
        self.delete_nodes([NODE])
        assert "total_count" in data.keys(), data
        assert data["total_count"] > 0, data
        assert data["data"][0]["file"]["name"] == "AAA.txt"

    
    def test_get_location_info(self):
        LOCATION = "S3" # this is defined in the Locations.yaml in {REPO}/test/config.d 
        TESTURL = "/".join( [SHOCK_URL , "location" , LOCATION , "info"  ] )

        data, response = GETJSON( TESTURL, headers=TESTAHEADERS)
    
    def test_get_location_missing(self):
        NODE = self.create_nodes([LONEFILEAAA])[0]
        NODEURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)

        LOCATION = "S3" # this is defined in the Locations.yaml in {REPO}/test/config.d 
        TESTURL = "/".join( [SHOCK_URL , "location" , LOCATION ,  "missing"  ] )

        response = requests.get( TESTURL, headers=TESTAHEADERS)
        if DEBUG:
            print ("URL", TESTURL)
            print("DATA", response.text)
        assert response.status_code == 200
 
    def test_types_get_info(self):
        LOCATION = "metadata"
        TESTURL = "/".join( [SHOCK_URL , "types" , LOCATION , "info"  ] )

        response = requests.get( TESTURL, headers=TESTAHEADERS)
        if DEBUG:
            print ("URL", TESTURL)
            print("DATA", response.text)
        assert response.status_code == 200
    
    def test_get_location_info(self):
        LOCATION = "S3" # this is defined in the Locations.yaml in {REPO}/test/config.d 
        TESTURL = "/".join( [SHOCK_URL , "location" , LOCATION , "info"  ] )

        response = requests.get( TESTURL, headers=TESTAHEADERS)
        if DEBUG:
            print ("URL", TESTURL)
            print("DATA", response.text)
        assert response.status_code == 200


    def test_get_location_missing1(self):
        NODE = self.create_nodes([LONEFILEAAA])[0]
        NODEURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)

        LOCATION = "S3" # this is defined in the Locations.yaml in {REPO}/test/config.d 
        TESTURL = "/".join( [SHOCK_URL , "location" , LOCATION ,  "missing"  ] )

        response = requests.get( TESTURL, headers=TESTAHEADERS)
        if DEBUG:
            print ("URL", TESTURL)
            print("DATA", response.text)
        assert response.status_code == 200

    def test_NODE_set_location(self):
        NODE = self.create_nodes([LONEFILEAAA])[0]
             
        PARAMS = {"id" : "S3", "stored" : "true"}
        FORMDATA = {"attributes_str": (None, PARAMS) } #'{"project_id":"TESTPROJECT"}')}
        NEWHEADER= TESTAHEADERS 
        NEWHEADER["Content-Type"]="application/json"

        TESTURL = "/".join( [SHOCK_URL , "node", NODE, "locations" ] )

        response = requests.post(TESTURL, headers=NEWHEADER, data=json.dumps(PARAMS))   
        
        if DEBUG:
            print ("URL", TESTURL)
            print("DATA", response.text)
        assert response.status_code == 200
 
    def test_get_location_missing2(self):
        NODE = self.create_nodes([LONEFILEAAA])[0]
        NODEURL = "{SHOCK_URL}/node".format(SHOCK_URL=SHOCK_URL)

        LOCATION = "S3" # this is defined in the Locations.yaml in {REPO}/test/config.d 
        TESTURL = "/".join( [SHOCK_URL , "location" , LOCATION ,  "missing"  ] )

        response = requests.get( TESTURL, headers=TESTAHEADERS)
        if DEBUG:
            print ("URL", TESTURL)
            print("DATA", response.text)
        assert response.status_code == 200

    def test_node_set_location(self) :
        pass

    def test_node_get_location(self) :
        pass
