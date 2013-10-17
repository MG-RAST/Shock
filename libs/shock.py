"""A basic Shock(https://github.com/MG-RAST/Shock) python access class. 
Uses shock-client for high performance uploads and download if it is in 
the users path.

Authors:

* Jared Wilkening
* Travis Harrison
"""

#-----------------------------------------------------------------------------
# Imports
#-----------------------------------------------------------------------------

import cStringIO
import json
import os
import requests
import subprocess
import urllib

#-----------------------------------------------------------------------------
# Classes
#-----------------------------------------------------------------------------

class Client:
    
    shock_url = ''
    transport_method = ''
    auth_header = {}
    token = ''
    template = "An exception of type {0} occured. Arguments:\n{1!r}"
    
    def __init__(self, shock_url, token=''):
        self.shock_url = shock_url
        if token != '':
            self.set_auth(token)
        if self._cmd_exists('shock-client'):
            self.transport_method = 'shock-client'
        else:
            self.transport_method = 'requests'
        
    def set_auth(self, token):
        self.auth_header = {'Authorization': 'OAuth %s'%token}
        if self.transport_method == 'shock-client':
            self._set_shockclient_auth(token)
    
    def _set_shockclient_auth(self, token):
        proc = subprocess.Popen("shock-client auth set-token \'{\"access_token\": \"%s\"}\'"%(token), shell=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
        return_code = proc.wait()
        if return_code > 0:
            err = ""
            for line in proc.stderr:
                err += line
            raise Exception(u'Error setting auth token in shock-client: %s'%err)
    
    def get_acl(self, node):
        return self._manage_acl(node, 'get')
    
    def add_acl(self, node, acl, email):
        return self._manage_acl(node, 'put', acl, email)
    
    def delete_acl(self, node, acl, email):
        return self._manage_acl(node, 'delete', acl, email)
    
    def _manage_acl(self, node, method, acl=None, email=None):
        url = self.shock_url+'/node/'+node+'/acl'
        if acl and email:
            url += '?'+acl+'='+urllib.quote(email)
        try:
            if method == 'get':
                req = requests.get(url, headers=self.auth_header)
            elif method == 'put':
                req = requests.put(url, headers=self.auth_header)
            elif method == 'delete':
                req = requests.delete(url, headers=self.auth_header)
        except Exception as ex:
            message = self.template.format(type(ex).__name__, ex.args)
            raise Exception(u'Unable to connect to Shock server %s\n%s' %(url, message))
        if not (req.ok and req.text):
            raise Exception(u'Unable to connect to Shock server %s: %s' %(url, req.raise_for_status()))
        rj = req.json()
        if not (rj and isinstance(rj, dict) and all([key in rj for key in ['status','data','error']])):
            raise Exception(u'Return data not valid Shock format')
        if rj['error']:
            raise Exception('Shock error: %d: %s'%(rj['status'], rj['error'][0]))
        return rj['data']
    
    def get_node(self, node):
        return self._get_node_data('/'+node)
    
    def query_node(self, query):
        query_string = '?query&'+urllib.urlencode(query)
        return self._get_node_data(query_string)
    
    def _get_node_data(self, path):
        url = self.shock_url+'/node'+path
        try:
            rget = requests.get(url, headers=self.auth_header, allow_redirects=True)
        except Exception as ex:
            message = self.template.format(type(ex).__name__, ex.args)
            raise Exception(u'Unable to connect to Shock server %s\n%s' %(url, message))
        if not (rget.ok and rget.text):
            raise Exception(u'Unable to connect to Shock server %s: %s' %(url, rget.raise_for_status()))
        rj = rget.json()
        if not (rj and isinstance(rj, dict) and all([key in rj for key in ['status','data','error']])):
            raise Exception(u'Return data not valid Shock format')
        if rj['error']:
            raise Exception('Shock error: %d: %s'%(rj['status'], rj['error'][0]))
        return rj['data']
    
    def download_to_path(self, node, path):
        if node == '' or path == '':
            raise Exception(u'download_to_path requires non-empty node & path parameters')
        if self.transport_method == 'shock-client':
            return self._download_shockclient(node, path)
        url = '%s/node/%s?download'%(self.shock_url, node)
        try:
            rget = requests.get(url, headers=self.auth_header, allow_redirects=True, stream=True)
        except Exception as ex:
            message = self.template.format(type(ex).__name__, ex.args)
            raise Exception(u'Unable to connect to Shock server %s\n%s' %(url, message))
        if not (rget.ok):
            raise Exception(u'Unable to connect to Shock server %s: %s' %(url, rget.raise_for_status()))
        with open(path, 'wb') as f:
            for chunk in rget.iter_content(chunk_size=8192): 
                if chunk:
                    f.write(chunk)
                    f.flush()
        return path

    def _download_shockclient(self, node, path):
        proc = subprocess.Popen("shock-client pdownload -threads=4 %s %s"%(node,path), shell=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
        return_code = proc.wait()
        if return_code > 0:
            err = ""
            for line in proc.stderr:
                err += line
            raise Exception(u'Error downloading via shock-client: %s => %s: error: %s' %(node, path, err))
        else: 
            return path
    
    def create_node(self, data='', attr='', data_name=''):
        return self.upload("", data, attr, data_name)
        
    def upload(self, node='', data='', attr='', data_name=''):
        try:
            if self.transport_method == 'shock-client' and node == '' and os.path.exists(data):
                res = self._upload_shockclient(data)
                if attr == '':
                    return res
                else:
                    node = res['id']
                    data = ''
        except TypeError:
            # data was file handle, we skip shock-client
            pass
        method = 'post'
        files = {}
        url = self.shock_url+'/node'
        if node != '':
            url = '%s/%s'%(url, node)
            method = 'put'
        if data != '':
            files['upload'] = self._get_handle(data, data_name)
        if attr != '':
            files['attributes'] = self._get_handle(attr)
        try:
            if method == 'put':
                req = requests.put(url, headers=self.auth_header, files=files, allow_redirects=True)
            else:
                req = requests.post(url, headers=self.auth_header, files=files, allow_redirects=True)
            rj = req.json()
        except Exception as ex:
            message = self.template.format(type(ex).__name__, ex.args)
            raise Exception(u'Unable to connect to Shock server %s\n%s' %(url, message))
        if not (req.ok):
            raise Exception(u'Unable to connect to Shock server %s: %s' %(url, req.raise_for_status()))
        if rj['error']:
            raise Exception(u'Shock error %s : %s'%(rj['status'], rj['error'][0]))
        else:
            return rj['data']  

    def _upload_shockclient(self, path):       
        proc = subprocess.Popen("shock-client pcreate -threads=4 -full %s"%(path), shell=True, stderr=subprocess.PIPE, stdout=subprocess.PIPE)
        return_code = proc.wait()
        if return_code > 0:
            err = ""
            for line in proc.stderr:
                err += line
            raise Exception(u'Error uploading via shock-client: %s: error: %s' %(path, err))
        else:
            res = ""
            for line in proc.stdout:
                if 'Uploading' not in line: 
                    res += line             
            return json.loads(res)
    
    # handles 3 cases
    # 1. file path
    # 2. file object (handle)
    # 3. file content (string)
    def _get_handle(self, d, n=''):
        try:
            if os.path.exists(d):
                name = n if n else os.path.basename(d)
                return (name, open(d))            
            else:
                name = n if n else "n/a"
                return (name, cStringIO.StringIO(d))
        except TypeError:
            try:
                name = n if n else d.name
                return (name, d)
            except:
                raise Exception(u'Error opening file handle for upload')

    def _cmd_exists(self, cmd):
        return subprocess.call("type " + cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE) == 0
