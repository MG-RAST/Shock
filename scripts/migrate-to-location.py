#!/usr/bin/env python3

import argparse , boto3 , hashlib , json , locations , logging , os , re , requests , shutil , sys , time
from os import environ as environ

import locations

config = {
    'shock' : {
        'host'      : None ,
        'token'     : None ,
        'bearer'    : None ,
    },
    'S3' : {
        's3_service_name'           : 's3' ,
        's3_aws_access_key_id'      : None ,
        's3_aws_secret_access_key'  : None ,
        's3_endpoint_url'           : None ,
        's3_bucket_name'            : None ,
    }
}



def set_command_line_options() :
    
    # parser = argparse.ArgumentParser()

    # # CLI
    
    # parser.add_argument("--location", dest="location",  required=True , help='set location; get node IDs for location if action is provided, otherwise get IDs from file or command line option',  default=None)
    # parser.add_argument("--action", dest="action", choices=['info', 'missing', 'inflight' , 'present'] ,  default=None)
    # parser.add_argument("--node-ids",  dest="node_id",  help='migrate nodes with id',  nargs="*" , default=[])
    # parser.add_argument("--input",  dest="file",  help='input file with node IDs, one ID per line',  nargs="*" , default=[])
    # parser.add_argument("--output", type=str, dest='output' , default=None)                         
    # parser.add_argument("--debug", action="store_true", default=False)
    # parser.add_argument("--config", action="store_true", default=False )


  
    # # Shock
    # parser.add_argument("--shock-host", type=str , dest='host' , default="http://shock.mg-rast.org" )
    # parser.add_argument("--token", type=str, dest='token' , default=None) 
      
    # # S3
    # parser.add_argument("--s3-access-key", dest="s3_access_key", default=None)
    # parser.add_argument("--s3-secret-access-key",  dest="s3_secret_access_key",  default=None)  
    # parser.add_argument("--s3-endpoint-url",  dest="s3_endpoint_url",  default=None) 
    # parser.add_argument("--s3-bucket-name",   dest="s3_bucket_name",    default=None)                      


    # if len(sys.argv)==1:
    #     parser.print_help(sys.stderr)
    #     sys.exit(1)

    # return parser.parse_args()

   
    # CLI
    
    command_parser = argparse.ArgumentParser()
    command_parser.add_argument("--shock-host", type=str , dest='host' , default="http://shock.mg-rast.org" )
    command_parser.add_argument("--token", type=str, dest='token' , default=None) 
    command_parser.add_argument("--output", type=str, dest='output' , default=None)                         
    command_parser.add_argument("--debug", action="store_true", default=False)
    command_parser.add_argument("--location", dest="location",  required=True , help='set location; get node IDs for location if action is provided, otherwise get IDs from file or command line option',  default=None)
    command_parser.add_argument("--node-ids",  dest="node_id",  help='migrate nodes with id',  nargs="*" , default=[])
    command_parser.add_argument("--input",  dest="file",  help='input file with node IDs, one ID per line',  nargs="*" , default=[])
    command_parser.add_argument("--config", action="store_true", default=False )

    subparsers = command_parser.add_subparsers(help='Choose a command', dest='subparser' )

    location_parser = subparsers.add_parser('s3' , help='Location cli')
    location_parser.add_argument('--bucket'  , type=str, dest='bucket' , default=None , help='Bucket name')
    location_parser.add_argument("--access-key", dest="s3_access_key", default=None)
    location_parser.add_argument("--secret-access-key",  dest="s3_secret_access_key",  default=None)  
    location_parser.add_argument("--endpoint-url",  dest="s3_endpoint_url",  default=None) 
    location_parser.add_argument("--bucket-name",   dest="s3_bucket_name",    default=None) 
    location_parser.set_defaults(action=lambda: 's3')

    shock_parser = subparsers.add_parser('shock', help='"get" location info for node')
    shock_parser.add_argument("--token", type=str, dest='shock_token' , default=None)
    shock_parser.set_defaults(action=lambda: 'shock')

    tsm_parser = subparsers.add_parser('tsm', help='"set" location for node')
    tsm_parser.set_defaults(action=lambda: 'set')


    command_parser.parse_args()
    
    if len(sys.argv)==1:
        command_parser.print_help(sys.stderr)
        sys.exit(1)

    return command_parser.parse_args()

def get_shock_node(node_id , config=None) :
    '''Retrieve node data from shock'''

    headers         = { 'Authorization' : " ".join(
                                                    [ 
                                                        config['shock']['bearer'] , 
                                                        config['shock']['token'] 
                                                    ]
                                                )  
                        }

    sleep = 60
    max   = 3
    tries = 0

    response = None
    while (not response) and tries < max :
        try:
            response = requests.get(config['shock']['host'] + "/node/" + node_id , headers=headers)
        except Exception as e: 
            print(str(e))
            logging.error(e)

        tries = tries + 1         
        time.sleep( sleep * tries )


    node = None
    if response.status_code == 200 :
        envelope = response.json()

        if "data" in envelope :
            node = envelope['data']
        else:
            print("Wrong object")
            sys.exit(1)

    elif response.status_code == 401 :
        print("Error 401\tcheck token passed on command line or in SHOCK_TOKEN")     
        sys.exit(401)  
    elif response.status_code == 404 :
        print('Error: Node not found (' + node_id + ')') 
    else :
        print(response.status_code)

    return node

def get_file_from_shock(node , max=1 , current = 0 ):

    # counter for tries
    current = current + 1

    print("Downloading " + node['id'] )
    file_name = node['id'] + ".data"

    headers         = { 'Authorization' : " ".join(
                                                    [ 
                                                        config['shock']['bearer'] , 
                                                        config['shock']['token'] 
                                                    ]
                                                )  
                        }
    download_url    = config['shock']['host'] + "/node/"  + node['id'] + "?download"
    

    with requests.get(download_url , headers=headers, stream=True) as response :
        with open( file_name , mode='wb') as localfile:   
            shutil.copyfileobj(response.raw, localfile)  
            # localfile.write(response.content)

    local_md5 = get_digest(file_name)
    if local_md5 == node['file']['checksum']['md5'] :
        print("Download ok")
    else:
        print("Error, download failed (%s)" , download_url  )
        print(local_md5 , node['file']['checksum']['md5'])
        if max > current :
            # wait before trying again
            time.sleep(sleep * current)
            file_name , tmp = get_file_from_shock(node=node , max=max , current=current)
        else:
            file_name = None

    return file_name , node['file']['checksum']['md5']

def push_to_s3(s3resource=None , file_name=None , md5=None , bucket=None , object_name=None , node=None) :

    """Upload a file to an S3 bucket

    :param file_name: File to upload
    :param bucket: Bucket to upload to
    :param object_name: S3 object name. If not specified then file_name is used
    :return: True if file was uploaded, else False
    """

    s3client = s3resource.meta.client

    # s3uri = 's3://anlseq/' + file_name

    # If S3 object_name was not specified, use file_name
    if object_name is None:
        object_name = file_name

    # Upload the file

    response = None 
    try:
        # response = s3client.upload_file(file_name, bucket, object_name , ExtraArgs={'Metadata': {'shock-md5': md5 }})
        response = s3resource.Object( bucket , file_name ).upload_file(file_name , ExtraArgs={'Metadata': {'shock-md5': md5 }} )
    except Exception as e: 
        print(str(e))
        logging.error(e)
        print(response)
        return False

    if not response :
        
        summary = s3resource.meta.client.head_object(Bucket = bucket, Key = file_name)
        print(summary['ETag'] , summary['ContentLength'])
        if ( re.search( md5 , summary['ETag']) ) :
            print('Same md5')
        else :
            print( 'MD5s:' , md5 , summary['ETag'])

        if node :
            if summary['ContentLength'] == node['file']['size'] :
                print('Same length')
            else :
                print('Length:' , node['file']['size'] , summary['ContentLength'])
            
            if ( re.search( md5 , summary['ETag']) ) or ( summary['ContentLength'] == node['file']['size'] ) :
                print( 'MD5 and/or length matching for ' + file_name )
                os.remove(file_name)
                pass
            else :
                return False
        
    else:
        return False

    return True

def configure(args) :

    # Set config, check env and command line options
    if args.host :
        config['shock']['host'] = args.host
    else :
        config['shock']['host'] = environ.get('SHOCK_URL') if environ.get('SHOCK_URL') else 'http://shock.mg-rast.org'

    # check url 
    if not re.match("http://" , config['shock']['host']) :
        sys.stderr.write("Missing http:// prefix for shock url")
        sys.exit()

    # set token
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
    

    info = locations.get_location_info(config=config , location=args.location)
    if not info :
        sys.stderr.write('Missing or unknown location '  + str(args.location) )
        sys.exit()
    else :
        if args.debug :
            print(info)
        if info['data']['type'] == 'S3' :
            config['S3']['s3_endpoint_url'] = info['data']['url']
            config['S3']['s3_bucket_name']  = info['data']['bucket']

            config['S3']['s3_service_name']             = 's3' 
            config['S3']['s3_aws_access_key_id']        = args.s3_access_key if args.s3_access_key else environ['S3_ACCESS_KEY']  
            config['S3']['s3_aws_secret_access_key']    = args.s3_secret_access_key if args.s3_secret_access_key else environ['S3_SECRET_ACCESS_KEY'] 
            # config['S3']['s3_endpoint_url'] = args.s3_endpoint_url if args.s3_endpoint_url else environ['S3_ENDPOINT_URL'] 
            # config['S3']['s3_bucket_name']  = args.s3_bucket_name if args.s3_bucket_name else environ['S3_BUCKET_NAME'] 

        elif info['data']['type'] == 'tsm' :
            sys.stderr.write('Unsupported type:\t' + info['data']['type'] )
            sys.exit()
        else:
            sys.stderr.write('Unsupported type:\t' + info['data']['type'] )
            sys.exit()
        
        
    
    # S3
    # config['S3']['s3_service_name']             = 's3' 
    # config['S3']['s3_aws_access_key_id']        = args.s3_access_key if args.s3_access_key else environ['S3_ACCESS_KEY']  
    # config['S3']['s3_aws_secret_access_key']    = args.s3_secret_access_key if args.s3_secret_access_key else environ['S3_SECRET_ACCESS_KEY'] 
    # config['S3']['s3_endpoint_url'] = args.s3_endpoint_url if args.s3_endpoint_url else environ['S3_ENDPOINT_URL'] 
    # config['S3']['s3_bucket_name']  = args.s3_bucket_name if args.s3_bucket_name else environ['S3_BUCKET_NAME'] 

    config['location']  = args.location
    

    if not config['location'] :
        sys.stderr.write('Missing location')
        sys.exit(404)
    

    return config

def get_digest(file_path):
    h = hashlib.md5()

    with open(file_path, 'rb') as file:
        while True:
            # Reading is buffered, so we can read smaller chunks.
            chunk = file.read(h.block_size)
            if not chunk:
                break
            h.update(chunk)

    return h.hexdigest()

def populate_list(node_ids=[] , file=None):

    if file and os.path.isfile(file) :
        f = open(file, "r") 
        for id in f :
            # print(id.strip("\n"))
            node_ids.append( id.strip("\n") )

    if args.location  :
        for i in locations.get_ids_for_location(config=config , location=args.location , action='missing' ):  
            node_ids.append(i)

    return node_ids

def main(config) :

    if args.debug and args.config :
        print(config)

    ids = populate_list( node_ids = args.node_id , file=args.file )  # list of shock node ids

    # remove in production
    if args.debug :
        print(ids)
        
    s3session   = None
    s3resource  = None
    s3client    = None    
    if 's3' :
        s3session   = boto3.session.Session()
        s3resource  = s3session.resource(
                    service_name            = config['S3']['s3_service_name'] ,
                    aws_access_key_id       = config['S3']['s3_aws_access_key_id'],
                    aws_secret_access_key   = config['S3']['s3_aws_secret_access_key'],
                    endpoint_url            = config['S3']['s3_endpoint_url'],
            )
        s3client = s3resource.meta.client

    #######################

    for node_id in ids :
        # check if location is set 
        # if not download file
        
        print("Get node info for " + node_id)
        node = get_shock_node(node_id , config=config)
        print(node)
        found = False
        if "locations" in node:
            for l in node['locations'] :
                print(l)
                if l['id'] == args.location :
                    found = True
           
        # skip loop
        if found :
            sys.stderr.write("Skipping " + node_id + "\n")
            print("Skipping " + node_id  )
            continue

        file_name , md5  = get_file_from_shock(node , max=max)

        print(file_name)
        
        # move to S3 storage
        success = False
        if file_name :
            success = push_to_s3(s3resource=s3resource , file_name=file_name , md5=md5 , bucket=config['S3']['s3_bucket_name'] , object_name=None , node=node)
        else : 
            print('No file downloaded, or misising file name: ' + str(file_name))

        if success :
            response = locations.set_node_location(config=config , node_id=node_id , location=args.location)
            if response :
                print("Set location " + args.location + " to " + node_id )
            else :
                print( "Error:\tCan not set location for " + node_id )
        if not success :
            print("Error pushing " + file_name )


if __name__== "__main__" :
    args = set_command_line_options()
    config = configure( args ) 
    main(config)        