# -*- coding: utf-8 -*-
from common import MSSERVERS
import json

WORKERS = MSSERVERS[2:] # does not include the gateway

tmplt_cloud_json="""{
    "NLAMBDA": %d,
    "PEERS": %s,
    "MAXKEY": 100000,
    "REDISIP": "%s"
}
"""

tmplt_constant_go = """package common

type AnyJson = map[string]interface{}
const T = %d
const ADDR = "127.0.0.1:18080"

const REDIS = true
"""

with open("../ccmesh-server/config/cloud.json", "w") as f:
    f.write(tmplt_cloud_json % (len(WORKERS), json.dumps(WORKERS), MSSERVERS[0]))

with open("../flowerkz/pkg/common/constants.go", "w") as f:
    f.write(tmplt_constant_go % (len(WORKERS)))
