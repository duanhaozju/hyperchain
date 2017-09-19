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
public class ContractGrpc {

  private ContractGrpc() {}

  public static final String SERVICE_NAME = "Contract";

  // Static method descriptors that strictly reflect the proto.
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.Request,
      cn.hyperchain.protos.ContractProto.Response> METHOD_HEART_BEAT =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.UNARY,
          generateFullMethodName(
              "Contract", "HeartBeat"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Request.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Response.getDefaultInstance()));
  @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/1901")
  public static final io.grpc.MethodDescriptor<cn.hyperchain.protos.ContractProto.Message,
      cn.hyperchain.protos.ContractProto.Message> METHOD_REGISTER =
      io.grpc.MethodDescriptor.create(
          io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING,
          generateFullMethodName(
              "Contract", "Register"),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Message.getDefaultInstance()),
          io.grpc.protobuf.ProtoUtils.marshaller(cn.hyperchain.protos.ContractProto.Message.getDefaultInstance()));

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ContractStub newStub(io.grpc.Channel channel) {
    return new ContractStub(channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ContractBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    return new ContractBlockingStub(channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary and streaming output calls on the service
   */
  public static ContractFutureStub newFutureStub(
      io.grpc.Channel channel) {
    return new ContractFutureStub(channel);
  }

  /**
   */
  public static abstract class ContractImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     *HeartBeat used to detect the health of server
     * </pre>
     */
    public void heartBeat(cn.hyperchain.protos.ContractProto.Request request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnimplementedUnaryCall(METHOD_HEART_BEAT, responseObserver);
    }

    /**
     * <pre>
     *Interface that provides support to jvm execution which will estabilish a stream between Server
     *and Client. Message type provide the context necessary for server to respond appropriately.
     * </pre>
     */
    public io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Message> register(
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Message> responseObserver) {
      return asyncUnimplementedStreamingCall(METHOD_REGISTER, responseObserver);
    }

    @java.lang.Override public io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            METHOD_HEART_BEAT,
            asyncUnaryCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.Request,
                cn.hyperchain.protos.ContractProto.Response>(
                  this, METHODID_HEART_BEAT)))
          .addMethod(
            METHOD_REGISTER,
            asyncBidiStreamingCall(
              new MethodHandlers<
                cn.hyperchain.protos.ContractProto.Message,
                cn.hyperchain.protos.ContractProto.Message>(
                  this, METHODID_REGISTER)))
          .build();
    }
  }

  /**
   */
  public static final class ContractStub extends io.grpc.stub.AbstractStub<ContractStub> {
    private ContractStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ContractStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ContractStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ContractStub(channel, callOptions);
    }

    /**
     * <pre>
     *HeartBeat used to detect the health of server
     * </pre>
     */
    public void heartBeat(cn.hyperchain.protos.ContractProto.Request request,
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Response> responseObserver) {
      asyncUnaryCall(
          getChannel().newCall(METHOD_HEART_BEAT, getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     *Interface that provides support to jvm execution which will estabilish a stream between Server
     *and Client. Message type provide the context necessary for server to respond appropriately.
     * </pre>
     */
    public io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Message> register(
        io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Message> responseObserver) {
      return asyncBidiStreamingCall(
          getChannel().newCall(METHOD_REGISTER, getCallOptions()), responseObserver);
    }
  }

  /**
   */
  public static final class ContractBlockingStub extends io.grpc.stub.AbstractStub<ContractBlockingStub> {
    private ContractBlockingStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ContractBlockingStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ContractBlockingStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ContractBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     *HeartBeat used to detect the health of server
     * </pre>
     */
    public cn.hyperchain.protos.ContractProto.Response heartBeat(cn.hyperchain.protos.ContractProto.Request request) {
      return blockingUnaryCall(
          getChannel(), METHOD_HEART_BEAT, getCallOptions(), request);
    }
  }

  /**
   */
  public static final class ContractFutureStub extends io.grpc.stub.AbstractStub<ContractFutureStub> {
    private ContractFutureStub(io.grpc.Channel channel) {
      super(channel);
    }

    private ContractFutureStub(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ContractFutureStub build(io.grpc.Channel channel,
        io.grpc.CallOptions callOptions) {
      return new ContractFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     *HeartBeat used to detect the health of server
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cn.hyperchain.protos.ContractProto.Response> heartBeat(
        cn.hyperchain.protos.ContractProto.Request request) {
      return futureUnaryCall(
          getChannel().newCall(METHOD_HEART_BEAT, getCallOptions()), request);
    }
  }

  private static final int METHODID_HEART_BEAT = 0;
  private static final int METHODID_REGISTER = 1;

  private static class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final ContractImplBase serviceImpl;
    private final int methodId;

    public MethodHandlers(ContractImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_HEART_BEAT:
          serviceImpl.heartBeat((cn.hyperchain.protos.ContractProto.Request) request,
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
        case METHODID_REGISTER:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.register(
              (io.grpc.stub.StreamObserver<cn.hyperchain.protos.ContractProto.Message>) responseObserver);
        default:
          throw new AssertionError();
      }
    }
  }

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    return new io.grpc.ServiceDescriptor(SERVICE_NAME,
        METHOD_HEART_BEAT,
        METHOD_REGISTER);
  }

}