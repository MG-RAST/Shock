#!/usr/bin/env cwl-runner

cwlVersion: v1.0
class: Workflow

requirements:
 ScatterFeatureRequirement: {}

inputs:
  node_ids      : string[]
  shock_url     : string
  shock_token   : string

outputs:
  migrated_nodes:
    type: File
    outputSource: migrate/migrated

steps:

  predict-location:
    run: predict-location.tool.cwl
    scatter: node_id
    in:
      node_id       : node_ids
      shock_url     : shock_url
      shock_token   : shock_token
    out: [ node_id , location ]

  migrate:
    run: migrate-to-s3.tool.cwl
    scatter: [ node_id , location ] 
    scatterMethod: dotproduct
    in:
      node_id       : predict-location/node_id
      location      : predict-location/location
      shock_url     : shock_url
      token         : shock_token
    out: [migrated]

