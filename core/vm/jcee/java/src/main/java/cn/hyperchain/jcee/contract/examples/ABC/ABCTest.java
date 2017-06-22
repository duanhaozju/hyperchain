package cn.hyperchain.jcee.contract.examples.ABC;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.contract.examples.ABC.src.ABC;
import cn.hyperchain.jcee.contract.examples.ABC.src.Account;
import cn.hyperchain.jcee.contract.examples.sb.src.SimulateBank;
import cn.hyperchain.jcee.mock.MockServer;

import java.util.Arrays;

/**
 * Created by huhu on 2017/6/21.
 */
public class ABCTest {
    private static MockServer server = new MockServer();

    public static String deploy(ContractTemplate ct){
        String cid = server.deploy(ct);
        server.setCid(cid);
        return cid;
    }

    public static void testNewAndGetAccount(){
        String[] newArgs = new String[]{"A11", "Acc1", "01", "CT00", "RC01", "abc01", "abcBank", "bejing", "11111111"};
        String[] getArgs = new String[]{"A11"};

        server.invoke("newAccount", Arrays.asList(newArgs));

        ExecuteResult result = server.invoke("getAccount",Arrays.asList(getArgs));
        Account account = Account.class.cast(result.getResult());
        System.out.println(account.getAccountNumber());

    }

    public static void testIssueDraft(){
        String[] newArgs1 = new String[]{"A1", "Acc1", "01", "CT00", "RC01", "abc01", "abcBank", "bejing", "11111111"};
        String[] newArgs2 = new String[]{"A2", "Acc1", "01", "CT00", "RC01", "abc01", "abcBank", "bejing", "11111111"};
        String[] newArgs3 = new String[]{"A3", "Acc1", "01", "CT00", "RC01", "abc01", "abcBank", "bejing", "11111111"};

        String[] orderArgs = new String[]{"order1", "100000000", "A1", "A1", "abc01", "A2", "A2", "abc01", "20170605", "20170610"};
        String[] draftArgs = new String[]{"draft1", "AC02", "1", "20170630", "20170630", "A1", "A2", "A3", "draftApply", "order1", "false"};

        String[] getDraftArgs = new String[]{"draft1"};
        server.invoke("newAccount", Arrays.asList(newArgs1));
        server.invoke("newAccount", Arrays.asList(newArgs2));
        server.invoke("newAccount", Arrays.asList(newArgs3));

        server.invoke("pubOrderInfo", Arrays.asList(orderArgs));
        server.invoke("issueDraftApply", Arrays.asList(draftArgs));

        server.invoke("getDraft", Arrays.asList(getDraftArgs));


    }

    public static void testBatch(){
        server.invoke("testBatch",Arrays.asList(new String[]{}));
    }
    public static void main(String[] args) {
        ABC abc = new ABC();
        deploy(abc);
//        testNewAndGetAccount();
//        testIssueDraft();
        testBatch();
    }
}
