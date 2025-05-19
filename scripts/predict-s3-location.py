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

    parser.add_argument("--shock-host", type=str , dest='host' , default=None )
    parser.add_argument("--token", type=str, dest='token' , default=None) 
    parser.add_argument("--output", type=str, dest='output' , default=None)                         
    parser.add_argument("--debug", action="store_true", default=False)
    parser.add_argument('node_id', nargs=1)

    
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







def main(config) :

    # Set header
    headers         = { 'Authorization' : " ".join(
                                                    [ 
                                                        config['shock']['bearer'] , 
                                                        config['shock']['token'] 
                                                    ]
                                                )  
                        }
    
    # Set url for node retrieval 
    node_id = None
    if len( args.node_id)  == 1  :
        node_id =  args.node_id[0]
    else:
        sys.stderr.write("Too many IDs, only one permitted")
        sys.exit(1)

    url    = "/".join( [ config['shock']['host'] , "node" , node_id ] )

    # print(url)
    # make request and parse json 
    data = None
    with requests.get(url , headers=headers ) as response :
        if response :
            try:
                data = response.json()
            except Exception as err :
                sys.stderr.write( "Error parsing response: " + str(err) + "\n")
        else : 
            sys.stderr.write( "Error retrieving data from MG-RAST (" + str(response.status_code) + ")\n")
            try:
                error = response.json()
                sys.stderr.write( "\n".join( error['error'] ) + " " + node_id + "\n" )
            except Exception as err :
                sys.stderr.write(err)
                sys.stderr.write("\n")
                sys.stderr.write( str(response.text) + "\n" )

    if data :

        owner       = None
        location    = None
        bucket      = None

        if 'data' in data and 'attributes' in data['data'] :  
            attributes = data['data']['attributes']
            if 'owner' in attributes :
                # print(attributes['owner'])
                if attributes['owner'] == 'ANL-SEQ-Core' :
                    owner = attributes['owner']
                    location    = 'anls3_anlseq'
                    bucket      = 'anlseq'
                elif re.match( "mgu" , attributes['owner']) :
                    owner       = 'mgrast'
                    location    = 'anls3_mgrast'
                    bucket      = 'mgrast'
                else :
                    sys.stderr.write("Unknown owner\t" + node_id + "\n")
                    sys.exit()
            elif 'job_id' in attributes and 'type' in attributes and 'project_id'in attributes :
                owner       = 'mgrast'
                location    = 'anls3_mgrast'
                bucket      = 'mgrast'
            else :
                owner       = ''
                location    = ''
                bucket      = ''
                sys.stderr.write("Unknown node type\t" + node_id + "\n")

            
            print( node_id + "\t" + ":".join([ owner , location , bucket ] ))   
        
    else :
        print(node_id + "\t" + "::")
        


if __name__== "__main__" :
    args = set_command_line_options()
    config = configure( args ) 
    main(config)