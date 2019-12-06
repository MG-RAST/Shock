#!/usr/bin/env cwl-runner

cwlVersion: v1.0
class: CommandLineTool
baseCommand: env

hints:
  DockerRequirement:
    dockerPull: mgrast/shock-tools:latest

requirements:
  EnvVarRequirement:
    envDef:
      # Shock
      SHOCK_URL : $(inputs.shock_url)
      SHOCK_TOKEN : $(inputs.token)


inputs:
  shock_url     : string
  shock_token   : string
  node_id       : string
outputs:
  location:
    type: stdout
  node_id:
    type: string
stdout: output.txt