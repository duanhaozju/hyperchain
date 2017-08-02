package cn.hyperchain.jcee.mock.test;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.contract.examples.sb.src.SimulateBank;
import cn.hyperchain.jcee.mock.MockServer;
import org.apache.log4j.Logger;

import java.util.Arrays;

/**
 * Created by huhu on 2017/6/21.
 */
public class SimulateBankTest {

    private static MockServer server = new MockServer();
    protected static Logger logger = Logger.getLogger(SimulateBankTest.class);


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
        System.out.println("balance: " + result.getResult());
        logger.info(result.getResult());

    }

    public static void testDelete(){
        server.invoke("testDelete",Arrays.asList(new String[]{}));
    }


    public static void testRangeQuery(){

        server.invoke("testRangeQuery",Arrays.asList(new String[]{}));
    }

    public static void testNewAccountTable() {

        ExecuteResult result = server.invoke("newAccountTable", Arrays.asList(new String[]{}));
        System.out.println(result.getResult());
    }

    public static void testGetTableDesc() {
        String[] args = new String[]{"Account"};
        ExecuteResult result = server.invoke("getTableDesc", Arrays.asList(args));
        System.out.println(result.getResult());
    }

    public static void testIssueByTable() {
        String[] args = new String[]{"bk-1", "bk-1", "100"};
        String[] args2 = new String[]{"bk-2", "bk-2", "100"};
        server.invoke("issueByTable", Arrays.asList(args));
        server.invoke("issueByTable", Arrays.asList(args2));
    }

    public static void testTransferByTable() {
        String[] args = new String[]{"bk-1", "bk-2", "10"};
        server.invoke("transferByTable", Arrays.asList(args));
    }

    public static void testGetAccountByRange() {
        String[] args = new String[]{"bk-1", "bk-2"};
        server.invoke("getAccountByRange", Arrays.asList(args));
    }

    public static void testSysQuery() {
        ExecuteResult result =  server.invoke("testSysQuery", Arrays.asList(new String[]{}));
        System.out.println("SysQuery result is :" + result);
    }
    public static void main(String[] args) {
        SimulateBank sb = new SimulateBank();
        deploy(sb);
//        testIssueAndGetBalance();
//        testRangeQuery();
//        testDelete();

//        testNewAccountTable();
//        testGetTableDesc();
//        testIssueByTable();
//        testTransferByTable();
//        testGetAccountByRange();

        testSysQuery();
    }

}
