# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: pipelines/protos/snapshots.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n pipelines/protos/snapshots.proto\x12\tsnapshots\"\x8d\x01\n\rExportRequest\x12\x11\n\tnamespace\x18\x01 \x01(\x05\x12\x0f\n\x07workers\x18\x02 \x01(\x05\x12\x0f\n\x07project\x18\x03 \x01(\t\x12\x0e\n\x06prefix\x18\x04 \x01(\t\x12\x10\n\x08language\x18\x05 \x01(\t\x12\r\n\x05since\x18\x06 \x01(\x03\x12\x16\n\x0e\x65xclude_events\x18\x07 \x03(\t\"/\n\x0e\x45xportResponse\x12\r\n\x05total\x18\x01 \x01(\x05\x12\x0e\n\x06\x65rrors\x18\x02 \x01(\x05\"C\n\x0b\x43opyRequest\x12\x0f\n\x07workers\x18\x01 \x01(\x05\x12\x10\n\x08projects\x18\x02 \x03(\t\x12\x11\n\tnamespace\x18\x03 \x01(\x05\"-\n\x0c\x43opyResponse\x12\r\n\x05total\x18\x01 \x01(\x05\x12\x0e\n\x06\x65rrors\x18\x02 \x01(\x05\"1\n\x10\x41ggregateRequest\x12\x0e\n\x06prefix\x18\x01 \x01(\t\x12\r\n\x05since\x18\x02 \x01(\x03\"2\n\x11\x41ggregateResponse\x12\r\n\x05total\x18\x01 \x01(\x05\x12\x0e\n\x06\x65rrors\x18\x02 \x01(\x05\"\x16\n\x14\x41ggregateCopyRequest\"6\n\x15\x41ggregateCopyResponse\x12\r\n\x05total\x18\x01 \x01(\x05\x12\x0e\n\x06\x65rrors\x18\x02 \x01(\x05\x32\x9f\x02\n\tSnapshots\x12=\n\x06\x45xport\x12\x18.snapshots.ExportRequest\x1a\x19.snapshots.ExportResponse\x12\x37\n\x04\x43opy\x12\x16.snapshots.CopyRequest\x1a\x17.snapshots.CopyResponse\x12\x46\n\tAggregate\x12\x1b.snapshots.AggregateRequest\x1a\x1c.snapshots.AggregateResponse\x12R\n\rAggregateCopy\x12\x1f.snapshots.AggregateCopyRequest\x1a .snapshots.AggregateCopyResponseB\x0bZ\t./;protosb\x06proto3')



_EXPORTREQUEST = DESCRIPTOR.message_types_by_name['ExportRequest']
_EXPORTRESPONSE = DESCRIPTOR.message_types_by_name['ExportResponse']
_COPYREQUEST = DESCRIPTOR.message_types_by_name['CopyRequest']
_COPYRESPONSE = DESCRIPTOR.message_types_by_name['CopyResponse']
_AGGREGATEREQUEST = DESCRIPTOR.message_types_by_name['AggregateRequest']
_AGGREGATERESPONSE = DESCRIPTOR.message_types_by_name['AggregateResponse']
_AGGREGATECOPYREQUEST = DESCRIPTOR.message_types_by_name['AggregateCopyRequest']
_AGGREGATECOPYRESPONSE = DESCRIPTOR.message_types_by_name['AggregateCopyResponse']
ExportRequest = _reflection.GeneratedProtocolMessageType('ExportRequest', (_message.Message,), {
  'DESCRIPTOR' : _EXPORTREQUEST,
  '__module__' : 'pipelines.protos.snapshots_pb2'
  # @@protoc_insertion_point(class_scope:snapshots.ExportRequest)
  })
_sym_db.RegisterMessage(ExportRequest)

ExportResponse = _reflection.GeneratedProtocolMessageType('ExportResponse', (_message.Message,), {
  'DESCRIPTOR' : _EXPORTRESPONSE,
  '__module__' : 'pipelines.protos.snapshots_pb2'
  # @@protoc_insertion_point(class_scope:snapshots.ExportResponse)
  })
_sym_db.RegisterMessage(ExportResponse)

CopyRequest = _reflection.GeneratedProtocolMessageType('CopyRequest', (_message.Message,), {
  'DESCRIPTOR' : _COPYREQUEST,
  '__module__' : 'pipelines.protos.snapshots_pb2'
  # @@protoc_insertion_point(class_scope:snapshots.CopyRequest)
  })
_sym_db.RegisterMessage(CopyRequest)

CopyResponse = _reflection.GeneratedProtocolMessageType('CopyResponse', (_message.Message,), {
  'DESCRIPTOR' : _COPYRESPONSE,
  '__module__' : 'pipelines.protos.snapshots_pb2'
  # @@protoc_insertion_point(class_scope:snapshots.CopyResponse)
  })
_sym_db.RegisterMessage(CopyResponse)

AggregateRequest = _reflection.GeneratedProtocolMessageType('AggregateRequest', (_message.Message,), {
  'DESCRIPTOR' : _AGGREGATEREQUEST,
  '__module__' : 'pipelines.protos.snapshots_pb2'
  # @@protoc_insertion_point(class_scope:snapshots.AggregateRequest)
  })
_sym_db.RegisterMessage(AggregateRequest)

AggregateResponse = _reflection.GeneratedProtocolMessageType('AggregateResponse', (_message.Message,), {
  'DESCRIPTOR' : _AGGREGATERESPONSE,
  '__module__' : 'pipelines.protos.snapshots_pb2'
  # @@protoc_insertion_point(class_scope:snapshots.AggregateResponse)
  })
_sym_db.RegisterMessage(AggregateResponse)

AggregateCopyRequest = _reflection.GeneratedProtocolMessageType('AggregateCopyRequest', (_message.Message,), {
  'DESCRIPTOR' : _AGGREGATECOPYREQUEST,
  '__module__' : 'pipelines.protos.snapshots_pb2'
  # @@protoc_insertion_point(class_scope:snapshots.AggregateCopyRequest)
  })
_sym_db.RegisterMessage(AggregateCopyRequest)

AggregateCopyResponse = _reflection.GeneratedProtocolMessageType('AggregateCopyResponse', (_message.Message,), {
  'DESCRIPTOR' : _AGGREGATECOPYRESPONSE,
  '__module__' : 'pipelines.protos.snapshots_pb2'
  # @@protoc_insertion_point(class_scope:snapshots.AggregateCopyResponse)
  })
_sym_db.RegisterMessage(AggregateCopyResponse)

_SNAPSHOTS = DESCRIPTOR.services_by_name['Snapshots']
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z\t./;protos'
  _EXPORTREQUEST._serialized_start=48
  _EXPORTREQUEST._serialized_end=189
  _EXPORTRESPONSE._serialized_start=191
  _EXPORTRESPONSE._serialized_end=238
  _COPYREQUEST._serialized_start=240
  _COPYREQUEST._serialized_end=307
  _COPYRESPONSE._serialized_start=309
  _COPYRESPONSE._serialized_end=354
  _AGGREGATEREQUEST._serialized_start=356
  _AGGREGATEREQUEST._serialized_end=405
  _AGGREGATERESPONSE._serialized_start=407
  _AGGREGATERESPONSE._serialized_end=457
  _AGGREGATECOPYREQUEST._serialized_start=459
  _AGGREGATECOPYREQUEST._serialized_end=481
  _AGGREGATECOPYRESPONSE._serialized_start=483
  _AGGREGATECOPYRESPONSE._serialized_end=537
  _SNAPSHOTS._serialized_start=540
  _SNAPSHOTS._serialized_end=827
# @@protoc_insertion_point(module_scope)
