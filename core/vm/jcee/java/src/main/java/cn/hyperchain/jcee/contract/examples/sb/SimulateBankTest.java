package cn.hyperchain.jcee.contract.examples.sb;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.contract.examples.sb.src.SimulateBank;
import cn.hyperchain.jcee.mock.MockServer;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * Created by huhu on 2017/6/21.
 */
public class SimulateBankTest {

    private static MockServer server = new MockServer();

    public static String deploy(ContractTemplate ct){
        String cid = server.deploy(ct);
        server.setCid(cid);
        return cid;
    }

    public static void testIssueAndGetBalance(){
        String[] issueArgs = new String[]{"A","100"};
        String[] getArgs = new String[]{"A"};

        server.invoke("issue",Arrays.asList(issueArgs));

        ExecuteResult result = server.invoke("getAccountBalance",Arrays.asList(getArgs));
        System.out.println(result.getResult());

    }

    public static void testDelete(){
        server.invoke("testDelete",Arrays.asList(new String[]{}));
    }

    public static void testRangeQuery(){
        server.invoke("testRangeQuery",Arrays.asList(new String[]{}));
    }
    public static void main(String[] args) {
        SimulateBank sb = new SimulateBank();
        deploy(sb);
//        testIssueAndGetBalance();
//        testRangeQuery();
        testDelete();
    }

}
