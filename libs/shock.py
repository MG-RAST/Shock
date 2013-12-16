"""A basic Shock (https://github.com/MG-RAST/Shock) python access class.

Authors:

* Jared Wilkening
* Travis Harrison
"""

#-----------------------------------------------------------------------------
# Imports
#-----------------------------------------------------------------------------

import cStringIO
import os
import requests
import urllib

#-----------------------------------------------------------------------------
# Classes
#-----------------------------------------------------------------------------

class Client:
    
    shock_url = ''
    auth_header = {}
    token = ''
    template = "An exception of type {0} occured. Arguments:\n{1!r}"
    
    def __init__(self, shock_url, token=''):
        self.shock_url = shock_url
        if token != '':
            self.set_auth(token)
        
    def set_auth(self, token):
        self.auth_header = {'Authorization': 'OAuth %s'%token}
    
    def get_acl(self, node):
        return self._manage_acl(node, 'get')
    
    def add_acl(self, node, acl, user):
        return self._manage_acl(node, 'put', acl, user)
    
    def delete_acl(self, node, acl, user):
        return self._manage_acl(node, 'delete', acl, user)
    
    def _manage_acl(self, node, method, acl=None, user=None):
        url = self.shock_url+'/node/'+node+'/acl'
        if acl and user:
            url += '/'+acl+'?users='+urllib.quote(user)
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
    
    def download_to_path(self, node, path, index=None, part=None, chunk=None):
        if node == '' or path == '':
            raise Exception(u'download_to_path requires non-empty node & path parameters')
        url = '%s/node/%s?download'%(self.shock_url, node)
        if index and part:
            url += '&index='+index+'&part='+str(part)
            if chunk:
                url += '&chunk_size='+str(chunk)
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
    
    def delete_node(self, node):
        url = self.shock_url+'/node/'+node
        try:
            req = requests.delete(url, headers=self.auth_header)
            rj  = req.json()
        except Exception as ex:
            message = self.template.format(type(ex).__name__, ex.args)
            raise Exception(u'Unable to connect to Shock server %s\n%s' %(url, message))
        if rj['error']:
            raise Exception(u'Shock error %s : %s'%(rj['status'], rj['error'][0]))
        return rj
    
    def index_node(self, node, index):
        url = "%s/node/%s/index/%s"%(self.shock_url, node, index)
        try:
            req = requests.put(url, headers=self.auth_header)
            rj  = req.json()
        except Exception as ex:
            message = self.template.format(type(ex).__name__, ex.args)
            raise Exception(u'Unable to connect to Shock server %s\n%s' %(url, message))
        if rj['error']:
            raise Exception(u'Shock error %s : %s'%(rj['status'], rj['error'][0]))
        return rj
    
    def create_node(self, data='', attr='', data_name=''):
        return self.upload("", data, attr, data_name)

    # file_name is name of data file
    # form == True for multi-part form
    # form == False for data POST of file
    def upload(self, node='', data='', attr='', file_name='', form=True):
        method = 'POST'
        files = {}
        url = self.shock_url+'/node'
        if node != '':
            url = '%s/%s'%(url, node)
            method = 'PUT'
        if data != '':
            files['upload'] = self._get_handle(data, file_name)
        if attr != '':
            files['attributes'] = self._get_handle(attr)
        if form:
            try:
                if method == 'PUT':
                    req = requests.put(url, headers=self.auth_header, files=files, allow_redirects=True)
                else:
                    req = requests.post(url, headers=self.auth_header, files=files, allow_redirects=True)
                rj = req.json()
            except Exception as ex:
                message = self.template.format(type(ex).__name__, ex.args)
                raise Exception(u'Unable to connect to Shock server %s\n%s' %(url, message))
        elif (not form) and data:
            try:
                if method == 'PUT':
                    req = requests.put(url, headers=self.auth_header, data=files['upload'][1], allow_redirects=True)
                else:
                    req = requests.post(url, headers=self.auth_header, data=files['upload'][1], allow_redirects=True)
                rj = req.json()
            except Exception as ex:
                message = self.template.format(type(ex).__name__, ex.args)
                raise Exception(u'Unable to connect to Shock server %s\n%s' %(url, message))
        else:
            raise Exception(u'No data specificed for %s body'%method)
        if not (req.ok):
            raise Exception(u'Unable to connect to Shock server %s: %s' %(url, req.raise_for_status()))
        if rj['error']:
            raise Exception(u'Shock error %s : %s'%(rj['status'], rj['error'][0]))
        else:
            return rj['data']
    
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
                name = n if n else "unknown"
                return (name, cStringIO.StringIO(d))
        except TypeError:
            try:
                name = n if n else d.name
                return (name, d)
            except:
                raise Exception(u'Error opening file handle for upload')
