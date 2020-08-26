#!/usr/bin/env python3

import argparse , json , logging , os , re , requests , sys
from os import environ as environ


config = {
    'shock' : {
        'host'      : None ,
        'token'     : None ,
        'bearer'    : None ,
    }
}

def set_command_line_options() :
    
    parser = argparse.ArgumentParser()

    parser.add_argument("--shock-host", type=str , dest='host' , default="http://shock.mg-rast.org" )
    parser.add_argument("--token", type=str, dest='token' , default=None) 
    parser.add_argument("--output", type=str, dest='output' , default=None)                         
    parser.add_argument("--debug", action="store_true", default=False)
  


    
    command_parser = argparse.ArgumentParser()
    command_parser.add_argument("--shock-host", type=str , dest='host' , default="http://shock.mg-rast.org" )
    command_parser.add_argument("--token", type=str, dest='token' , default=None) 
    command_parser.add_argument("--output", type=str, dest='output' , default=None)                         
    command_parser.add_argument("--debug", action="store_true", default=False)

    subparsers = command_parser.add_subparsers(help='Choose a command', dest='subparser' )

    location_parser = subparsers.add_parser('loc' , help='Location cli')
    location_parser.add_argument('location_option'  , choices=['info', 'missing', 'present' , 'inflight' , 'ls' ] , help='Location options')
    location_parser.add_argument('location_name',   nargs="?" , type=str , default=None  , help='Location name')
    location_parser.set_defaults(action=lambda: 'loc')

    get_parser = subparsers.add_parser('get', help='"get" location info for node')
    get_parser.add_argument('node_id' , help='Node ID')
    get_parser.add_argument('location_name',   nargs="?" , type=str , default=None  , help='Location name')
    get_parser.set_defaults(action=lambda: 'get')

    set_parser = subparsers.add_parser('set', help='"set" location for node')
    set_parser.add_argument('node_id' , help='Node ID')
    set_parser.add_argument('location_name' , type=str , default=None  , help='Location name')
    set_parser.set_defaults(action=lambda: 'set')

    set_parser = subparsers.add_parser('delete', help='"delete" location for node')
    set_parser.add_argument('node_id' , help='Node ID')
    set_parser.add_argument('location_name' , type=str , default=None  , help='Location name')
    set_parser.set_defaults(action=lambda: 'set')

    command_parser.parse_args()
    
    if len(sys.argv)==1:
        command_parser.print_help(sys.stderr)
        sys.exit(1)

    return command_parser.parse_args()

def configure(args) :
    # SHOCK
    config['shock'] = {}
    if args.host :
        config['shock']['host'] = args.host
    else :
        config['shock']['host'] = environ.get('SHOCK_URL') if environ.get('SHOCK_URL') else 'http://shock.mg-rast.org'

    if not re.match("http://" , config['shock']['host']) :
        sys.stderr.write("Missing http:// prefix for shock url")
        sys.exit()

    token = args.token if args.token else environ.get('SHOCK_TOKEN')

    if not token :
        sys.stderr.write('No auth token\n')
    else:
        # check for bearer
        m = re.match('([^\s]+)\s+(.+)' , token)
        if m :
            config['shock']['token'] = m[2]
            config['shock']['bearer'] = m[1] # default bearer
        else :
            config['shock']['token'] = token
            config['shock']['bearer'] = 'OAuth' # default bearer
    # print(config)
    return config

def get_ids_for_location(config=None , location=None , action=None ):
    ids = []
    if action != 'info' :
        for i in  get_nodes(config=config , location=location, action=action) :
            ids.append( i['id'] )    
    return ids

def get_nodes(config=None , location=None , action=None ) :
    # init
    data = None # return data from request as data structure

    if not config :
       sys.stderr.write('Missing config to talk to shock\n')
       sys.exit()

    # set header and request url
    headers         = { 'Authorization' : " ".join(
                                                    [ 
                                                        config['shock']['bearer'] , 
                                                        config['shock']['token'] 
                                                    ]
                                                )  
                        }
    location_url    = "/".join( [ config['shock']['host'] , "location"  , location , action ] )
    # location_url    = "/".join( [ config['shock']['host'] , "node?limit=100" ] )
    
    print(location_url)
    # make request and parse json 
    with requests.get(location_url , headers=headers, stream=False) as response :
        if response :
            try:
                data = response.json()
            except Exception as err :
                sys.stderr.write( "Error parsing response: " + str(err) + "\n")
        else : 
            sys.stderr.write( "Error retrieving data from Shock (" + str(response.status_code) + ")\n")
            sys.stderr.write( "URL: " + location_url + "\n" )
            try:
                error = response.json()
                sys.stderr.write( "\n".join( error['error'] ) + "\n" )
            except Exception as err :
                sys.stderr.write(err)
                sys.stderr.write("\n")
                sys.stderr.write( str(response.text) + "\n" )
             
    return data['data'] if data else []

def get_location_info(config=None , location=None) :
    return get_location(config=config , location=location , action='info')

def get_location(config=None , location=None , action=None ):

    # init
    data = None # return data from request as data structure

    if not config :
       sys.stderr.write('Missing config to talk to shock')
       sys.exit()

    # set header and request url
    headers         = { 'Authorization' : " ".join(
                                                    [ 
                                                        config['shock']['bearer'] , 
                                                        config['shock']['token'] 
                                                    ]
                                                )  
                        }
    location_url    = "/".join( [ config['shock']['host'] , "location"  , location , action ] )
    # location_url    = "/".join( [ config['shock']['host'] , "node?limit=100" ] )
    
    print(location_url)
    # make request and parse json 
    with requests.get(location_url , headers=headers, stream=True) as response :
        if response :
            try:
                data = response.json()
            except Exception as err :
                sys.stderr.write( "Error parsing response: " + str(err) + "\n")
        else : 
            sys.stderr.write( "Error retrieving data from " + location_url + " (" + str(response.status_code) + ")\n")
            try:
                error = response.json()
                sys.stderr.write( "\n".join( error['error'] ) + "\n" )
            except Exception as err :
                sys.stderr.write(err)
                sys.stderr.write("\n")
                sys.stderr.write( str(response.text) + "\n" )
             
    return data 

def get_node_location(config=None , node_id=None ):
    url    = "/".join( [ config['shock']['host'] , "node"  , node_id  ] )
    return get_request(config=config, url=url) 

def set_node_location(config=None , node_id=None , location=None):

    if node_id and location :
        url = config['shock']['host'] + "/node/" + node_id + "/locations/" + location
        headers = { 
            "Authorization" : " ".join(
                                                    [ 
                                                        config['shock']['bearer'] , 
                                                        config['shock']['token'] 
                                                    ]
                                                ),
            "Content-Type" : "application/json"
            }
        response = requests.post( 
                                url, 
                                data='{\"id\":\"' + location + '\" , "stored" : true }' , 
                                headers=headers
                            )    
        
        data = None
        if response :
                try:
                    data = response.json()
                except Exception as err :
                    sys.stderr.write( "Error parsing response: " + str(err) + "\n")
        else : 
            sys.stderr.write( "Error retrieving data from Shock (" + str(response.status_code) + ")\n")
            try:
                error = response.json()
                sys.stderr.write( "\n".join( error['error'] ) + "\n" )
            except Exception as err :
                sys.stderr.write(err)
                sys.stderr.write("\n")
                sys.stderr.write( str(response.text) + "\n" )
    
    return data

def get_request(config=None , url=None , streaming=None) :
    
    # init
    data = None # return data from request as data structure

    if not config :
        sys.stderr.write('Missing config to talk to shock')
        sys.exit()

    # set header and request url
    headers         = { 'Authorization' : " ".join(
                                                    [ 
                                                        config['shock']['bearer'] , 
                                                        config['shock']['token'] 
                                                    ]
                                                )  
                        }
    print(url)
    # make request and parse json 
    with requests.get(url , headers=headers, stream=streaming) as response :
        if response :
            try:
                data = response.json()
            except Exception as err :
                sys.stderr.write( "Error parsing response: " + str(err) + "\n")
        else : 
            sys.stderr.write( "Error retrieving data from Shock (" + str(response.status_code) + ")\n" + "URL: " + url + "\n")
            try:
                error = response.json()
                sys.stderr.write( "\n".join( error['error'] ) + "\n" )
            except Exception as err :
                sys.stderr.write(err)
                sys.stderr.write("\n")
                sys.stderr.write( str(response.text) + "\n" )
                
    return data 

def post_request(url) :
    print("Setting location " + location )

    headers = { 'Authorizatioun' : shock_token ,
                "Content-Type" : "application/json" }
    response = requests.post(shock_url + "/node/" + node_id + "/locations/" + location , data='{\"id\":\"' + location + ' \" , "stored" : true }' , headers=headers)    

def print_location(data):

    if data :
        if 'data' in data :
            nodes = []
            if type(data['data']) == dict :
                nodes.append(data['data'])
            else:
                nodes =+ data['data']
            for n in nodes :
                if 'locations' in n :
                    locs = map(lambda x: x['id'] , n['locations'])
                    print( "\t".join( [ n['id'] ] + list(locs)) )
                else:
                    print( "\t".join( [ n['id'] , 'None' ] ))
        else :
            sys.stderr.write("Error: Not a shock node\n")
    else:
        print('None')


def main(config) :
    info    = None
    results = None
    ofile   = None

    # print(args.location , args.status)
    if args.subparser == 'loc' :
        if  args.location_name and len(args.location_name) > 0 and args.location_option == 'info' :
            info = get_location(config=config , location=args.location_name , action='info')
            print(info)
            print(info['data'])
        elif args.location_option and len(args.location_option) > 0 and args.location_option != 'info':
            results = info = get_location(config=config , location=args.location_name , action=args.location_option )

            print(results)
            if args.output and os.path.isfile( args.output ) :
                ofile = open( args.output , 'w')
            if results :

                if 'data' in results :
                    for node in results['data'] :
                        print(results)
                        if node['file']['size'] > 0 :
                            print(node['id'], node['locations'])
                            if ofile :
                                ofile.write(node['id'], node['locations'] , "\n")
                else :
                    for d in results :
                        print(d)
                        if ofile :
                            ofile.write(d)

            if ofile :
                ofile.close()

        else: 
            print(info)
    
    elif args.subparser == 'get' :
        data = get_node_location(config=config, node_id=args.node_id)
        print_location(data)
    elif args.subparser == 'set' :
        data = set_node_location(config=config, node_id=args.node_id , location=args.location_name)
        if data :
            locs = map(lambda x: x['id'] , data['data'])
            print( "\t".join( [ args.node_id ] + list(locs)) )
        else:
            print("\t".join([args.node_id , 'None']))
    elif args.subparser == 'delete' :
        sys.stder.write('Not implemented')

if __name__== "__main__" :
    args = set_command_line_options()
    config = configure( args ) 
    main(config)