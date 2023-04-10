# -*- coding: utf-8 -*-
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: kv.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x08kv.proto\x12\x05proto\"\x19\n\nGetRequest\x12\x0b\n\x03key\x18\x01 \x01(\t\"\x1c\n\x0bGetResponse\x12\r\n\x05value\x18\x01 \x01(\x0c\"(\n\nPutRequest\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\x0c\"\x07\n\x05\x45mpty2Z\n\x02KV\x12,\n\x03Get\x12\x11.proto.GetRequest\x1a\x12.proto.GetResponse\x12&\n\x03Put\x12\x11.proto.PutRequest\x1a\x0c.proto.EmptyB4Z2github.com/hashicorp/go-plugin/examples/grpc/protob\x06proto3')

_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, globals())
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'kv_pb2', globals())
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z2github.com/hashicorp/go-plugin/examples/grpc/proto'
  _GETREQUEST._serialized_start=19
  _GETREQUEST._serialized_end=44
  _GETRESPONSE._serialized_start=46
  _GETRESPONSE._serialized_end=74
  _PUTREQUEST._serialized_start=76
  _PUTREQUEST._serialized_end=116
  _EMPTY._serialized_start=118
  _EMPTY._serialized_end=125
  _KV._serialized_start=127
  _KV._serialized_end=217
# @@protoc_insertion_point(module_scope)
