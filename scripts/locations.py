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
    parser.add_argument('location', nargs=1)
    parser.add_argument('status', nargs='?' , choices=['info', 'missing', 'present' , 'inflight' ] , help='Set behavior for call')
    
    if len(sys.argv)==1:
        parser.print_help(sys.stderr)
        sys.exit(1)

    return parser.parse_args()

def configure(args) :
    # SHOCK
    config['shock'] = {}
    if args.host :
        config['shock']['host'] = args.host
    else :
        config['shock']['host'] = environ.get('SHOCK_TOKEN') if environ.get('SHOCK_TOKEN') else 'http://shock.mg-rast.org'

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
    if action is not 'info' :
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
    # location_url    = "/".join( [ config['shock']['host'] , "location"  , location , action ] )
    location_url    = "/".join( [ config['shock']['host'] , "node?limit=100" ] )
    
    print(location_url)
    # make request and parse json 
    with requests.get(location_url , headers=headers, stream=True) as response :
        if response :
            try:
                data = response.json()
            except Exception as err :
                sys.stderr.write( "Error parsing response: " + str(err) + "\n")
        else : 
            sys.stderr.write( "Error retrieving data from MG-RAST (" + str(response.status_code) + ")\n")
            try:
                error = response.json()
                sys.stderr.write( "\n".join( error['error'] ) + "\n" )
            except Exception as err :
                sys.stderr.write(err)
                sys.stderr.write("\n")
                sys.stderr.write( str(response.text) + "\n" )
             
    return data['data'] 

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
    # location_url    = "/".join( [ config['shock']['host'] , "location"  , location , action ] )
    location_url    = "/".join( [ config['shock']['host'] , "node?limit=100" ] )
    
    print(location_url)
    # make request and parse json 
    with requests.get(location_url , headers=headers, stream=True) as response :
        if response :
            try:
                data = response.json()
            except Exception as err :
                sys.stderr.write( "Error parsing response: " + str(err) + "\n")
        else : 
            sys.stderr.write( "Error retrieving data from MG-RAST (" + str(response.status_code) + ")\n")
            try:
                error = response.json()
                sys.stderr.write( "\n".join( error['error'] ) + "\n" )
            except Exception as err :
                sys.stderr.write(err)
                sys.stderr.write("\n")
                sys.stderr.write( str(response.text) + "\n" )
             
    return data 

def main(config) :
    info    = None
    results = None
    ofile   = None

    print(args.location , args.status)
    if  args.location and len(args.location) > 0 and args.status != 'info' :
        info = get_location(config=config , location=args.location[0] , action='info')
    if args.status and len(args.status) > 0 :
        results = info = get_location(config=config , location=args.location[0] , action=args.status)

        if args.output and os.path.isfile( args.output ) :
            ofile = open( args.output , 'w')
        if results :

            if 'data' in results :
                for node in results['data'] :
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

if __name__== "__main__" :
    args = set_command_line_options()
    config = configure( args ) 
    main(config)