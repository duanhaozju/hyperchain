package cn.hyperchain.jcee;

import cn.hyperchain.jcee.executor.Handler;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.RequestContext;
import com.google.protobuf.ByteString;
import org.junit.Assert;
import org.junit.Test;

import java.nio.charset.Charset;

/**
 * Created by wangxiaoyi on 2017/4/25.
 */
public class TestHandler {

    @Test
    public void testDeploy() {

        String contractDir = TestHandler.class.getResource("/contracts").getPath();
//        System.out.println(contractDir);
        ContractGrpcServerImpl cgsi = new ContractGrpcServerImpl(123);
        Handler handler = new Handler(123);
        Request request = Request.newBuilder()
                .addArgs(ByteString.copyFrom(contractDir, Charset.defaultCharset()))
                .addArgs(ByteString.copyFrom("String", Charset.defaultCharset()))
                .addArgs(ByteString.copyFrom("long", Charset.defaultCharset()))
                .addArgs(ByteString.copyFrom("boolean", Charset.defaultCharset()))
                .addArgs(ByteString.copyFrom("bank001", Charset.defaultCharset()))
                .addArgs(ByteString.copyFrom("1", Charset.defaultCharset()))
                .addArgs(ByteString.copyFrom("true", Charset.defaultCharset()))
                .setContext(RequestContext.newBuilder().setCid("cid0001")
                            .setNamespace("global")
                            .setTxid("tx001"))
                .build();

        handler.deploy(request, null);
        Assert.assertEquals(true, handler.getContractMgr().getContract("cid0001").getCid().equals("cid0001"));
    }

    @Test
    public void testDeployWithoutConstructorArgs() {

        String contractDir = TestHandler.class.getResource("/contracts").getPath();
//        System.out.println(contractDir);
        ContractGrpcServerImpl cgsi = new ContractGrpcServerImpl(123);
        Handler handler = new Handler(123);
        Request request = Request.newBuilder()
                .addArgs(ByteString.copyFrom(contractDir, Charset.defaultCharset()))
                .setContext(RequestContext.newBuilder().setCid("cid0001")
                        .setNamespace("global")
                        .setTxid("tx001"))
                .build();

        handler.deploy(request, null);
        Assert.assertEquals(true, handler.getContractMgr().getContract("cid0001").getCid().equals("cid0001"));
    }
}
