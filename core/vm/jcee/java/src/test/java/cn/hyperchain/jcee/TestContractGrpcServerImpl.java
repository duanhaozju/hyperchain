package cn.hyperchain.jcee;

import cn.hyperchain.jcee.server.executor.ContractHandler;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import org.junit.Test;

import java.io.IOException;

public class TestContractGrpcServerImpl {

    @Test
    public void TestContractGrpcServer() throws Exception {


        ContractHandler.init(50081);
        ContractGrpcServerImpl service = new ContractGrpcServerImpl();
        service.init();
        Server server = ServerBuilder.forPort(50051)
                .addService(service)
                .build().start();
        System.out.println("Server start ...");
        server.awaitTermination();

    }


}
