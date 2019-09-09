# Command Line Client

Make sure the shock-client command is in your path. You will also need access to a running shock server. For demo purposes, I've installed the shock server on localhost. See installing.

**1. Set up your .shock-client.cfg file in your home directory.**  The only changes that I had to make to the template was set the server url. The design looks for -conf, then SHOCK_CLIENT_CONF environment variable, then the default location which is the current user's home directory.

The url requires the protocol (http) and port (7044).

[server]
url=http://localhost:7044

**2. Authenticate with shock server.** For this I'm using the Globus Online.

shock-client auth show

shock-client set

**3. Upload some data.**

For this example, I have ChIP-seq data. I have two files, Input_tags.bed and Treatement_tags.bed. The first is the control and the second represents a treatment. These files are used as in put into the MACS program which computes transcription factor binding sites.

In total, we will create three shock nodes. One node holding the meta-data and data for the control data represented in the Input_tags.bed file. The second node holding the meta-data and data for the treatment data in the Treatment_tags.bed file. The third node is a virtual node used to link the two preceding nodes.

shock-client create -attribute=input_tags.json -full=Input_tags.bed
shock-client create -attributes=treatment_tags.json -full=Treatment_tags.bed

The input parameters must have an = sign. Omitting the equal sign can cause unexpected and uninformative error messages.

These commands return the node id in a json document. Here is the output json document for each of the commands run above:

{
    "id": "546955d9-d8a0-484c-9dd6-639be14ff4ea",
    "version": "b6b51c4b26704c9544a870d079fb435f",
    "file": {
        "name": "Input_tags.bed",
        "size": 156714470,
        "checksum": {
            "md5": "74394f5fcb0b2b5ff73d5fc1f3cd599b",
            "sha1": "2385e57e8edec916902d29f654ca05f5abcc301e"
        },
        "format": "",
        "virtual": false,
        "virtual_parts": null
    },
    "attributes": {
        "type": "control"
    },
    "indexes": {
        "size": {
            "total_units": 150,
            "average_unit_size": 1048576
        }
    },
    "tags": null,
    "linkages": null
}

and

{
    "id": "123b625b-cab7-4464-8ae3-347bd17301cd",
    "version": "3da0e2d9202ea5f6d44a2af59f8ce4c5",
    "file": {
        "name": "Treatment_tags.bed",
        "size": 117087092,
        "checksum": {
            "md5": "dda3014ed8cbd5ed7dd39dfc1ee8311b",
            "sha1": "cbce35a86781d0ec510f357ba92473810b848bf8"
        },
        "format": "",
        "virtual": false,
        "virtual_parts": null
    },
    "attributes": {
        "type": "treatment"
    },
    "indexes": {
        "size": {
            "total_units": 112,
            "average_unit_size": 1048576
        }
    },
    "tags": null,
    "linkages": null
}

Now we can link these two nodes together using a virtual node like this:

shock-client create -attributes=foxa1.json -virtual_file=546955d9-d8a0-484c-9dd6-639be14ff4ea,123b625b-cab7-4464-8ae3-347bd17301cd

And we see this as the output:

{
    "id": "01e8a8f1-4bd6-4a26-8a3d-ada4ed226348",
    "version": "1f9752807344dfb4ce908d6611099c46",
    "file": {
        "name": "",
        "size": 273801562,
        "checksum": {
            "md5": "78c1eb8ce40abc5ad0814dd8b7523fb6",
            "sha1": "7a46b852d1730efe9e2fd0775bec97a39eb6d281"
        },
        "format": "",
        "virtual": true,
        "virtual_parts": [
            "123b625b-cab7-4464-8ae3-347bd17301cd",
            "546955d9-d8a0-484c-9dd6-639be14ff4ea"
        ]
    },
    "attributes": {
        "members": [
            "Input_tags.bed",
            "Treatment_tags.bed"
        ],
        "type": "set"
    },
    "indexes": {},
    "tags": null,
    "linkages": null
}



**4. Download that data.**

Now, we can download this data. We can download each file individually. To download both files requires each to be downloaded individually.

shock-client download 546955d9-d8a0-484c-9dd6-639be14ff4ea > Input_tags.bed.download
shock-client download 123b625b-cab7-4464-8ae3-347bd17301cd > Treatment_tags.bed.download

And to download the meta-data (attributes) for each file, each json document will be retrieved individually. You will notice that you downloaded more meta-data than you uploaded. That's to be expected. The meta-data you uploaded is returned as the "attributes" field in the json document.


shock-client get 546955d9-d8a0-484c-9dd6-639be14ff4ea > input_tags.json.download
shock-client get 123b625b-cab7-4464-8ae3-347bd17301cd > treatment_tags.json.download

more input_tags.json.download 
{
    "id": "546955d9-d8a0-484c-9dd6-639be14ff4ea",
    "version": "b6b51c4b26704c9544a870d079fb435f",
    "file": {
        "name": "Input_tags.bed",
        "size": 156714470,
        "checksum": {
            "md5": "74394f5fcb0b2b5ff73d5fc1f3cd599b",
            "sha1": "2385e57e8edec916902d29f654ca05f5abcc301e"
        },
        "format": "",
        "virtual": false,
        "virtual_parts": []
    },
    "attributes": {
        "type": "control"
    },
    "indexes": {
        "size": {
            "total_units": 150,
            "average_unit_size": 1048576
        }
    },
    "tags": [],
    "linkages": []
}

more treatment_tags.json.download 
{
    "id": "123b625b-cab7-4464-8ae3-347bd17301cd",
    "version": "3da0e2d9202ea5f6d44a2af59f8ce4c5",
    "file": {
        "name": "Treatment_tags.bed",
        "size": 117087092,
        "checksum": {
            "md5": "dda3014ed8cbd5ed7dd39dfc1ee8311b",
            "sha1": "cbce35a86781d0ec510f357ba92473810b848bf8"
        },
        "format": "",
        "virtual": false,
        "virtual_parts": []
    },
    "attributes": {
        "type": "treatment"
    },
    "indexes": {
        "size": {
            "total_units": 112,
            "average_unit_size": 1048576
        }
    },
    "tags": [],
    "linkages": []
}



