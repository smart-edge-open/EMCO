# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation
from pymongo import MongoClient
#Step 1: Connect to MongoDB - Note: Change connection string as needed
client = MongoClient(port=27017)
db=client.mco
#Step 2: remove the records
db.orchestrator.remove({})
db.controller.remove({})
db.customization.remove({})
db.resource.remove({})
db.cluster.remove({})
db.cloudconfig.remove({})
print("Cleared tables")