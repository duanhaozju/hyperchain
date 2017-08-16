package cn.hyperchain.jcee.client.executor;

import cn.hyperchain.jcee.client.contract.IContractManager;
import cn.hyperchain.jcee.client.contract.ContractInfo;
import cn.hyperchain.protos.ContractProto;
import io.grpc.stub.StreamObserver;

public interface IHandler {

    void freeze(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver);

    void unfreeze(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver);

    void destroy(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver);

    void update(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver);

    void invoke(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver);

    void deploy(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver);

    boolean deploy(ContractInfo info) throws ClassNotFoundException;

    IContractManager getContractMgr();
}
