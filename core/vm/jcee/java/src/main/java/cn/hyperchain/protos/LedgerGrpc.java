package cn.hyperchain.protos;

import static io.grpc.stub.ClientCalls.asyncUnaryCall;
import static io.grpc.stub.ClientCalls.asyncServerStreamingCall;
import static io.grpc.stub.ClientCalls.asyncClientStreamingCall;
import static io.grpc.stub.ClientCalls.asyncBidiStreamingCall;
import static io.grpc.stub.ClientCalls.blockingUnaryCall;
import static io.grpc.stub.ClientCalls.blockingServerStreamingCall;
import static io.grpc.stub.ClientCalls.futureUnaryCall;
import static io.grpc.MethodDescriptor.generateFullMethodName;
import static io.grpc.stub.ServerCalls.asyncUnaryCall;
import static io.grpc.stub.ServerCalls.asyncServerStreamingCall;
import static io.grpc.stub.ServerCalls.asyncClientStreamingCall;
import static io.grpc.stub.ServerCalls.asyncBidiStreamingCall;
import static io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall;
import static io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.0.3)",
    comments = "Source: contract.proto")
public class LedgerGrpc {

  private LedgerGrpc() {}

  public static final String SERVICE_NAME = "Ledger";

  // Static method descriptors that strictly reflect the proto.
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.Key,
      cn.hyperchain.protos.ContractProto.Value> METHOD_GET =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Ledger", "Get"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Key.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Value.getDefaultInstance()));
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.KeyValue,
      cn.hyperchain.protos.ContractProto.Response> METHOD_PUT =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Ledger", "Put"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.KeyValue.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Response.getDefaultInstance()));
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.Key,
      cn.hyperchain.protos.ContractProto.Response> METHOD_DELETE =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Ledger", "Delete"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Key.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Response.getDefaultInstance()));
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.BatchKey,
      cn.hyperchain.protos.ContractProto.BathValue> METHOD_BATCH_READ =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Ledger", "BatchRead"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.BatchKey.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.BathValue.getDefaultInstance()));
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.BatchKV,
      cn.hyperchain.protos.ContractProto.Response> METHOD_BATCH_WRITE =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Ledger", "BatchWrite"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.BatchKV.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Response.getDefaultInstance()));
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.Range,
      cn.hyperchain.protos.ContractProto.BathValue> METHOD_RANGE_QUERY =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING,
          generateFullMethodName(
              "Ledger", "RangeQuery"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Range.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.BathValue.getDefaultInstance()));
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.Event,
      cn.hyperchain.protos.ContractProto.Response> METHOD_POST =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Ledger", "Post"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Event.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Response.getDefaultInstance()));

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static LedgerStub newStub(io.grpc.Channel channel) {
    return new LedgerStub(channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static LedgerBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    return new LedgerBlockingStub(channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary and streaming output calls on the service
   */
  public static LedgerFutureStub newFutureStub(
      io.grpc.Channel channel) {
    return new LedgerFutureStub(channel);
  }

  /**
   */
  public static abstract class LedgerImplBase implements io.grpc.BindableService {

    /**
     */
    public void get(cn.hyperchain.protos.ContractProto.Key request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Value> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_GET, responseObserver);
    }

    /**
     */
    public void put(cn.hyperchain.protos.ContractProto.KeyValue request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_PUT, responseObserver);
    }

    /**
     */
    public void delete(cn.hyperchain.protos.ContractProto.Key request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_DELETE, responseObserver);
    }

    /**
     */
    public void batchRead(cn.hyperchain.protos.ContractProto.BatchKey request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.BathValue> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_BATCH_READ, responseObserver);
    }

    /**
     */
    public void batchWrite(cn.hyperchain.protos.ContractProto.BatchKV request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_BATCH_WRITE, responseObserver);
    }

    /**
     */
    public void rangeQuery(cn.hyperchain.protos.ContractProto.Range request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.BathValue> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_RANGE_QUERY, responseObserver);
    }

    /**
     */
    public void post(cn.hyperchain.protos.ContractProto.Event request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_POST, responseObserver);
    }

    @java.lang.Override public io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            METHOD_GET,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.Key,
                cn.hyperchain.protos.ContractProto.Value>(
                  this, METHODID_GET)))
          .addMethod(
            METHOD_PUT,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.KeyValue,
                cn.hyperchain.protos.ContractProto.Response>(
                  this, METHODID_PUT)))
          .addMethod(
            METHOD_DELETE,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.Key,
                cn.hyperchain.protos.ContractProto.Response>(
                  this, METHODID_DELETE)))
          .addMethod(
            METHOD_BATCH_READ,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.BatchKey,
                cn.hyperchain.protos.ContractProto.BathValue>(
                  this, METHODID_BATCH_READ)))
          .addMethod(
            METHOD_BATCH_WRITE,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.BatchKV,
                cn.hyperchain.protos.ContractProto.Response>(
                  this, METHODID_BATCH_WRITE)))
          .addMethod(
            METHOD_RANGE_QUERY,
            asyncServerStreamingCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.Range,
                cn.hyperchain.protos.ContractProto.BathValue>(
                  this, METHODID_RANGE_QUERY)))
          .addMethod(
            METHOD_POST,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.Event,
                cn.hyperchain.protos.ContractProto.Response>(
                  this, METHODID_POST)))
          .build();
    }
  }

  /**
   */
  public static final class LedgerStub extends io.grpc.stub.AbstractStub<LedgerStub> {
    private LedgerStub(io.grpc.Channel channel) {
      super(channel);
    }

    private LedgerStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LedgerStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new LedgerStub(channel, callOptions);
    }

    /**
     */
    public void get(cn.hyperchain.protos.ContractProto.Key request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Value> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_GET, getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void put(cn.hyperchain.protos.ContractProto.KeyValue request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_PUT, getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void delete(cn.hyperchain.protos.ContractProto.Key request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_DELETE, getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void batchRead(cn.hyperchain.protos.ContractProto.BatchKey request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.BathValue> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_BATCH_READ, getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void batchWrite(cn.hyperchain.protos.ContractProto.BatchKV request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_BATCH_WRITE, getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void rangeQuery(cn.hyperchain.protos.ContractProto.Range request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.BathValue> responseObserver) {
      asyncServerStreamingCall(
          getChannel().newCall(METHOD_RANGE_QUERY, getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void post(cn.hyperchain.protos.ContractProto.Event request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_POST, getCallOptions()), request, responseObserver);
    }
  }

  /**
   */
  public static final class LedgerBlockingStub extends io.grpc.stub.AbstractStub<LedgerBlockingStub> {
    private LedgerBlockingStub(io.grpc.Channel channel) {
      super(channel);
    }

    private LedgerBlockingStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LedgerBlockingStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new LedgerBlockingStub(channel, callOptions);
    }

    /**
     */
    public cn.hyperchain.protos.ContractProto.Value get(cn.hyperchain.protos.ContractProto.Key request) {
      return blockingUnaryCall(
          getChannel(), METHOD_GET, getCallOptions(), request);
    }

    /**
     */
    public cn.hyperchain.protos.ContractProto.Response put(cn.hyperchain.protos.ContractProto.KeyValue request) {
      return blockingUnaryCall(
          getChannel(), METHOD_PUT, getCallOptions(), request);
    }

    /**
     */
    public cn.hyperchain.protos.ContractProto.Response delete(cn.hyperchain.protos.ContractProto.Key request) {
      return blockingUnaryCall(
          getChannel(), METHOD_DELETE, getCallOptions(), request);
    }

    /**
     */
    public cn.hyperchain.protos.ContractProto.BathValue batchRead(cn.hyperchain.protos.ContractProto.BatchKey request) {
      return blockingUnaryCall(
          getChannel(), METHOD_BATCH_READ, getCallOptions(), request);
    }

    /**
     */
    public cn.hyperchain.protos.ContractProto.Response batchWrite(cn.hyperchain.protos.ContractProto.BatchKV request) {
      return blockingUnaryCall(
          getChannel(), METHOD_BATCH_WRITE, getCallOptions(), request);
    }

    /**
     */
    public java.util.Iterator<cn.hyperchain.protos.ContractProto.BathValue> rangeQuery(
        cn.hyperchain.protos.ContractProto.Range request) {
      return blockingServerStreamingCall(
          getChannel(), METHOD_RANGE_QUERY, getCallOptions(), request);
    }

    /**
     */
    public cn.hyperchain.protos.ContractProto.Response post(cn.hyperchain.protos.ContractProto.Event request) {
      return blockingUnaryCall(
          getChannel(), METHOD_POST, getCallOptions(), request);
    }
  }

  /**
   */
  public static final class LedgerFutureStub extends io.grpc.stub.AbstractStub<LedgerFutureStub> {
    private LedgerFutureStub(io.grpc.Channel channel) {
      super(channel);
    }

    private LedgerFutureStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LedgerFutureStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new LedgerFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.ContractProto.Value> get(
        cn.hyperchain.protos.ContractProto.Key request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_GET, getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.ContractProto.Response> put(
        cn.hyperchain.protos.ContractProto.KeyValue request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_PUT, getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.ContractProto.Response> delete(
        cn.hyperchain.protos.ContractProto.Key request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_DELETE, getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.ContractProto.BathValue> batchRead(
        cn.hyperchain.protos.ContractProto.BatchKey request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_BATCH_READ, getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.ContractProto.Response> batchWrite(
        cn.hyperchain.protos.ContractProto.BatchKV request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_BATCH_WRITE, getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.ContractProto.Response> post(
        cn.hyperchain.protos.ContractProto.Event request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_POST, getCallOptions()), request);
    }
  }

  private static final int METHODID_GET = 0;
  private static final int METHODID_PUT = 1;
  private static final int METHODID_DELETE = 2;
  private static final int METHODID_BATCH_READ = 3;
  private static final int METHODID_BATCH_WRITE = 4;
  private static final int METHODID_RANGE_QUERY = 5;
  private static final int METHODID_POST = 6;

  private static class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final LedgerImplBase serviceImpl;
    private final int methodId;

    public MethodHandlers(LedgerImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_GET:
          serviceImpl.get((cn.hyperchain.protos.ContractProto.Key) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Value>) responseObserver);
          break;
        case METHODID_PUT:
          serviceImpl.put((cn.hyperchain.protos.ContractProto.KeyValue) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response>) responseObserver);
          break;
        case METHODID_DELETE:
          serviceImpl.delete((cn.hyperchain.protos.ContractProto.Key) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response>) responseObserver);
          break;
        case METHODID_BATCH_READ:
          serviceImpl.batchRead((cn.hyperchain.protos.ContractProto.BatchKey) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.BathValue>) responseObserver);
          break;
        case METHODID_BATCH_WRITE:
          serviceImpl.batchWrite((cn.hyperchain.protos.ContractProto.BatchKV) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response>) responseObserver);
          break;
        case METHODID_RANGE_QUERY:
          serviceImpl.rangeQuery((cn.hyperchain.protos.ContractProto.Range) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.BathValue>) responseObserver);
          break;
        case METHODID_POST:
          serviceImpl.post((cn.hyperchain.protos.ContractProto.Event) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    return new io.grpc.ServiceDescriptor(SERVICE_NAME,
        METHOD_GET,
        METHOD_PUT,
        METHOD_DELETE,
        METHOD_BATCH_READ,
        METHOD_BATCH_WRITE,
        METHOD_RANGE_QUERY,
        METHOD_POST);
  }

}
