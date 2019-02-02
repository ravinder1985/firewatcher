import json
import sys
import httplib

# Following tag would be added to all the metrics belongs to this script
print("file: {}".format("config.xml"))
print("backend: {}".format("appserver"))

# Following would get added to only the configs added after Tag_
print("Tag_app.name_contain: {}".format("true"))
print("Tag_app.dbconnection_contain: {}".format("true"))

# Following would export result per properties.
print("app.name_Result: {}".format("0"))
print("app.dbconnection_Result: {}".format("1"))
print("app.dbconnection.url_Result: {}".format("1"))
print("app.dbconnection.username_Result: {}".format("1"))
