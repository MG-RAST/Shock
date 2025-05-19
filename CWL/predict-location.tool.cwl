#!/usr/bin/env cwl-runner

cwlVersion: v1.1
class: CommandLineTool
baseCommand: predict-s3-location.py

hints:
    DockerRequirement:
        dockerPull: mgrast/shock-tools:latest

requirements:
  EnvVarRequirement:
    envDef:
      # Shock
      SHOCK_URL : $(inputs.shock_url)
      SHOCK_TOKEN : $(inputs.shock_token)

inputs:
  shock_url     : string
  shock_token   : string
  node_id       : 
    type: string
    inputBinding:
        position: 1

outputs:
  location:
    type: stdout
    
stdout: output.txt