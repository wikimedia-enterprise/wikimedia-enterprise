# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: pipelines/protos/bulk.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1bpipelines/protos/bulk.proto\x12\x04\x62ulk\"\x11\n\x0fProjectsRequest\"1\n\x10ProjectsResponse\x12\r\n\x05total\x18\x01 \x01(\x05\x12\x0e\n\x06\x65rrors\x18\x02 \x01(\x05\"\x13\n\x11NamespacesRequest\"3\n\x12NamespacesResponse\x12\r\n\x05total\x18\x01 \x01(\x05\x12\x0e\n\x06\x65rrors\x18\x02 \x01(\x05\"5\n\x0f\x41rticlesRequest\x12\x11\n\tnamespace\x18\x01 \x01(\x05\x12\x0f\n\x07project\x18\x02 \x01(\t\"1\n\x10\x41rticlesResponse\x12\r\n\x05total\x18\x01 \x01(\x05\x12\x0e\n\x06\x65rrors\x18\x02 \x01(\x05\x32\xbd\x01\n\x04\x42ulk\x12\x39\n\x08Projects\x12\x15.bulk.ProjectsRequest\x1a\x16.bulk.ProjectsResponse\x12?\n\nNamespaces\x12\x17.bulk.NamespacesRequest\x1a\x18.bulk.NamespacesResponse\x12\x39\n\x08\x41rticles\x12\x15.bulk.ArticlesRequest\x1a\x16.bulk.ArticlesResponseB\x0bZ\t./;protosb\x06proto3')



_PROJECTSREQUEST = DESCRIPTOR.message_types_by_name['ProjectsRequest']
_PROJECTSRESPONSE = DESCRIPTOR.message_types_by_name['ProjectsResponse']
_NAMESPACESREQUEST = DESCRIPTOR.message_types_by_name['NamespacesRequest']
_NAMESPACESRESPONSE = DESCRIPTOR.message_types_by_name['NamespacesResponse']
_ARTICLESREQUEST = DESCRIPTOR.message_types_by_name['ArticlesRequest']
_ARTICLESRESPONSE = DESCRIPTOR.message_types_by_name['ArticlesResponse']
ProjectsRequest = _reflection.GeneratedProtocolMessageType('ProjectsRequest', (_message.Message,), {
  'DESCRIPTOR' : _PROJECTSREQUEST,
  '__module__' : 'pipelines.protos.bulk_pb2'
  # @@protoc_insertion_point(class_scope:bulk.ProjectsRequest)
  })
_sym_db.RegisterMessage(ProjectsRequest)

ProjectsResponse = _reflection.GeneratedProtocolMessageType('ProjectsResponse', (_message.Message,), {
  'DESCRIPTOR' : _PROJECTSRESPONSE,
  '__module__' : 'pipelines.protos.bulk_pb2'
  # @@protoc_insertion_point(class_scope:bulk.ProjectsResponse)
  })
_sym_db.RegisterMessage(ProjectsResponse)

NamespacesRequest = _reflection.GeneratedProtocolMessageType('NamespacesRequest', (_message.Message,), {
  'DESCRIPTOR' : _NAMESPACESREQUEST,
  '__module__' : 'pipelines.protos.bulk_pb2'
  # @@protoc_insertion_point(class_scope:bulk.NamespacesRequest)
  })
_sym_db.RegisterMessage(NamespacesRequest)

NamespacesResponse = _reflection.GeneratedProtocolMessageType('NamespacesResponse', (_message.Message,), {
  'DESCRIPTOR' : _NAMESPACESRESPONSE,
  '__module__' : 'pipelines.protos.bulk_pb2'
  # @@protoc_insertion_point(class_scope:bulk.NamespacesResponse)
  })
_sym_db.RegisterMessage(NamespacesResponse)

ArticlesRequest = _reflection.GeneratedProtocolMessageType('ArticlesRequest', (_message.Message,), {
  'DESCRIPTOR' : _ARTICLESREQUEST,
  '__module__' : 'pipelines.protos.bulk_pb2'
  # @@protoc_insertion_point(class_scope:bulk.ArticlesRequest)
  })
_sym_db.RegisterMessage(ArticlesRequest)

ArticlesResponse = _reflection.GeneratedProtocolMessageType('ArticlesResponse', (_message.Message,), {
  'DESCRIPTOR' : _ARTICLESRESPONSE,
  '__module__' : 'pipelines.protos.bulk_pb2'
  # @@protoc_insertion_point(class_scope:bulk.ArticlesResponse)
  })
_sym_db.RegisterMessage(ArticlesResponse)

_BULK = DESCRIPTOR.services_by_name['Bulk']
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z\t./;protos'
  _PROJECTSREQUEST._serialized_start=37
  _PROJECTSREQUEST._serialized_end=54
  _PROJECTSRESPONSE._serialized_start=56
  _PROJECTSRESPONSE._serialized_end=105
  _NAMESPACESREQUEST._serialized_start=107
  _NAMESPACESREQUEST._serialized_end=126
  _NAMESPACESRESPONSE._serialized_start=128
  _NAMESPACESRESPONSE._serialized_end=179
  _ARTICLESREQUEST._serialized_start=181
  _ARTICLESREQUEST._serialized_end=234
  _ARTICLESRESPONSE._serialized_start=236
  _ARTICLESRESPONSE._serialized_end=285
  _BULK._serialized_start=288
  _BULK._serialized_end=477
# @@protoc_insertion_point(module_scope)