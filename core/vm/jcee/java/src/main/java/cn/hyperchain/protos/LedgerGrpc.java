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
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.Key,
      cn.hyperchain.protos.Value> METHOD_GET =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Ledger", "Get"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.Key.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.Value.getDefaultInstance()));
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.KeyValue,
      cn.hyperchain.protos.Response> METHOD_PUT =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Ledger", "Put"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.KeyValue.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.Response.getDefaultInstance()));

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
    public void get(cn.hyperchain.protos.Key request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.Value> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_GET, responseObserver);
    }

    /**
     */
    public void put(cn.hyperchain.protos.KeyValue request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.Response> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_PUT, responseObserver);
    }

    @java.lang.Override public io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            METHOD_GET,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.Key,
                cn.hyperchain.protos.Value>(
                  this, METHODID_GET)))
          .addMethod(
            METHOD_PUT,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.KeyValue,
                cn.hyperchain.protos.Response>(
                  this, METHODID_PUT)))
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
    public void get(cn.hyperchain.protos.Key request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.Value> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_GET, getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void put(cn.hyperchain.protos.KeyValue request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.Response> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_PUT, getCallOptions()), request, responseObserver);
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
    public cn.hyperchain.protos.Value get(cn.hyperchain.protos.Key request) {
      return blockingUnaryCall(
          getChannel(), METHOD_GET, getCallOptions(), request);
    }

    /**
     */
    public cn.hyperchain.protos.Response put(cn.hyperchain.protos.KeyValue request) {
      return blockingUnaryCall(
          getChannel(), METHOD_PUT, getCallOptions(), request);
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
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.Value> get(
        cn.hyperchain.protos.Key request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_GET, getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.Response> put(
        cn.hyperchain.protos.KeyValue request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_PUT, getCallOptions()), request);
    }
  }

  private static final int METHODID_GET = 0;
  private static final int METHODID_PUT = 1;

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
          serviceImpl.get((cn.hyperchain.protos.Key) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.Value>) responseObserver);
          break;
        case METHODID_PUT:
          serviceImpl.put((cn.hyperchain.protos.KeyValue) request,
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.Response>) responseObserver);
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
        METHOD_PUT);
  }

}
