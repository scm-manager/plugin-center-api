#!/bin/bash

TOKEN=$1
shift
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/api/v1/plugins/2.26.0 $@
