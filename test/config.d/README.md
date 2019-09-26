# Shock Server Configuration



The server expects a config file `--conf shock-server.conf` and will use `/etc/shock.d/shock-server.conf` if it exists. 


We expext that most production scenarios will have a `/etc/shock.d/` directory with at least three files
- Locations.yaml -- External storage locations and credentials
- Types.yaml -- Definitions of data types and priorities
- shock-server.conf  -- main shock server config


The server expects `Locations.yaml` and `Types.yaml` in the same directory as the config file.


