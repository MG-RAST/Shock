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
  shock_url : string
  token     : string
  location  : string
  node_id   : string
outputs:
  migrated:
    type: stdout
stdout: output.txt