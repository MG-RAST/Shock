

### API Routes for /types (default port 7445):

##### OPTIONS

- Permitted by everyone:
  - all options requests respond with CORS headers and 200 OK

##### POST
- Permitted by everyone:
- N/A

- Permitted by admis 
  - `/types/<type_id>/info`  provide info on a data type
  
##### GET

- Permitted by everyone:
 - N/A

- Permitted by admis 
   


##### PUT

- N/A

Note: Configurations is via the Types.yaml file at server start time.


##### DELETE

- N/A

<br>


## An overview of /types

~~~~
>curl -s -X GET "localhost:7445/types/mixs/info" | jq .
{
  "status": 200,
  "data": {
    "id": "mixs",
    "description": "GSC MIxS Metadata file XLS format",
    "priority": 9
  },
  "error": null
}
~~~~





