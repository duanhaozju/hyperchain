// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: contract.proto

package cn.hyperchain.protos;

public interface RequestOrBuilder extends
    // @@protoc_insertion_point(interface_extends:Request)
    com.google.protobuf.MessageOrBuilder {

  /**
   * <code>string txid = 1;</code>
   */
  java.lang.String getTxid();
  /**
   * <code>string txid = 1;</code>
   */
  com.google.protobuf.ByteString
      getTxidBytes();

  /**
   * <code>string method = 2;</code>
   */
  java.lang.String getMethod();
  /**
   * <code>string method = 2;</code>
   */
  com.google.protobuf.ByteString
      getMethodBytes();

  /**
   * <code>string namespace = 3;</code>
   */
  java.lang.String getNamespace();
  /**
   * <code>string namespace = 3;</code>
   */
  com.google.protobuf.ByteString
      getNamespaceBytes();

  /**
   * <code>string cid = 4;</code>
   */
  java.lang.String getCid();
  /**
   * <code>string cid = 4;</code>
   */
  com.google.protobuf.ByteString
      getCidBytes();

  /**
   * <code>repeated bytes args = 5;</code>
   */
  java.util.List<com.google.protobuf.ByteString> getArgsList();
  /**
   * <code>repeated bytes args = 5;</code>
   */
  int getArgsCount();
  /**
   * <code>repeated bytes args = 5;</code>
   */
  com.google.protobuf.ByteString getArgs(int index);
}
