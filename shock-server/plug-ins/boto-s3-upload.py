#!/usr/bin/python

# boto3 python client to upload files to S3 and verify
# AWS_ACCESS_KEY_ID .. The access key for your AWS account.
# AWS_SECRET_ACCESS_KEY .. The secret key for your AWS account.

# folker@anl.gov

import sys, getopt, boto3, hashlib, io, os
import argparse

def md5sum(src, length=io.DEFAULT_BUFFER_SIZE):
    md5 = hashlib.md5()
    with io.open(src, mode="rb") as fd:
        for chunk in iter(lambda: fd.read(length), b''):
            md5.update(chunk)
    return md5.hexdigest()


def main():

   parser = argparse.ArgumentParser()
   parser.add_argument("-a","--keyid", default=None, help=" aws_access_key_id")
   parser.add_argument("-b","--bucket", default=None, help="AWS bucket")
   parser.add_argument("-f","--filepath",  default=None,help="local file to upload")
   parser.add_argument("-o","--objectname",  default=None,help="object name (key) in S3")
   parser.add_argument("-k","--accesskey",  default=None, help="aws_secret_access_key")
   parser.add_argument("-v", "--verbose", action="count", default=0, help="increase output verbosity")
   parser.add_argument("-r","--region", default=None, help="AWS region")
   parser.add_argument("-s","--s3endpoint",  default="https://s3.it.anl.gov:18082")
   parser.add_argument("--verify-only", action="store_true", help="only verify object exists with HEAD request")
   args = parser.parse_args()

   if args.verbose:
      print ('keyId  is =', args.keyid)
      print ('accessKey is =', args.accesskey)
      print ('bucket is =', args.bucket)
      print ('filepath is =', args.filepath)
      print ('region is=', args.region)
      print ('objectname is =', args.objectname)
      print ('verify_only is =', args.verify_only)

   if args.objectname is None:
      print ('we need an object name')
      sys.exit(2)

   if not args.verify_only and args.filepath is None:
      print ('we need a filepath for upload')
      sys.exit(2)

   # if passed use credentials to establish connection
   if args.accesskey is None:
      if args.verbose:
         print ('using existing credentials from ENV vars or files')
      s3 = boto3.client('s3',
            endpoint_url=args.s3endpoint,
            region_name=args.region
            )
   else:
      # use env. default for connection details --> see  https://boto3.amazonaws.com/v1/documentation/api/latest/guide/configuration.html
      if args.verbose:
         print ('using credentials from cmd-line')
      s3 = boto3.client('s3',
         endpoint_url=args.s3endpoint,
         region_name=args.region,
         aws_access_key_id=args.keyid,
         aws_secret_access_key=args.accesskey
      )

   # verify-only mode: check if object exists and return size
   if args.verify_only:
      try:
         response = s3.head_object(Bucket=args.bucket, Key=args.objectname)
         size = response['ContentLength']
         print(size)
         sys.exit(0)
      except Exception as e:
         if args.verbose:
            print('Object not found or error:', str(e))
         sys.exit(1)

   # upload mode
   if not os.path.exists(args.filepath):
      print('File not found:', args.filepath)
      sys.exit(2)

   file_size = os.path.getsize(args.filepath)

   with open(args.filepath, 'rb') as f:
      s3.upload_fileobj(f, args.bucket, args.objectname)

   # verify upload by checking size
   try:
      response = s3.head_object(Bucket=args.bucket, Key=args.objectname)
      uploaded_size = response['ContentLength']
      if uploaded_size == file_size:
         print(uploaded_size)
         sys.exit(0)
      else:
         if args.verbose:
            print('Size mismatch: local=%d, remote=%d' % (file_size, uploaded_size))
         sys.exit(1)
   except Exception as e:
      if args.verbose:
         print('Verification failed:', str(e))
      sys.exit(1)

main()
