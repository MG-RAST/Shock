
## Authentication and Authorization:

Shock supports multiple forms of Authentication via plugin modules. Credentials are cached for 1 hour to speed up high transaction loads. Server restarts will clear the credential cache.

### Globus Online 
In this configuration Shock locally stores only uuids for users that it has already seen. The registration of new users is done exclusively with the external auth provider. The user api is disabled in this mode.

Examples:

    # globus online username & password
    curl --user username:password ...

    # globus online bearer token 
    curl -H "Authorization: OAuth $TOKEN" ...


<br>

# OAuth

TODO: add documentation



# Basic access authentication

While not recommended for production, basic auth can be useful for testing and development. Users are stored with their username and password in the mongo database.

Enable with basic auth using:

`shock-server --basic=true --users=admin`

Users listed under `--users` are only users with admin rights !

Add normal user to database:
```bash
docker exec -ti test_shock-mongo_1 mongo

use ShockDB;
db.Users.insert({ username: "user1", password: "secret"}) 
db.Users.findOne()

or update password
db.Users.findOne({ username: "user1"})
db.Users.update({ username: "user1"}, { $set : {password: "newsecret" }}) 

```

For each user create base64-encoded credentials that can be used to access Shock.
```bash
echo -n 'user1:secret' | base64
curl -H 'Authorization: dXNlcjE6c2VjcmV0'  <shock-api>/node...
```

## --debug-auth=true
For testing and debugging purposes, the server supports disabling all access control via `--debug-auth=true`.



