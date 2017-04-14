package cn.hyperchain.jcee.contract.examples.c1;

import cn.hyperchain.jcee.contract.ContractBase;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/7.
 */
public class MySmartContract extends ContractBase {
    private final Logger logger = Logger.getLogger(MySmartContract.class.getSimpleName());

    public boolean Invoke(String funcName, List<String> args) {
        //logger.info("invoke function name: " + funcName);
        switch (funcName) {
            case "test": {
                this.test("invoke method:" + args.get(0));
                //String name = new String(ledger.get("name".getBytes()));
                //System.out.println("get name from ledger: " + name);
                return true;
            }
            default:
                logger.error("no such method found");
        }
        return false;
    }

    public ByteString Query(String funcName, List<String> args) {
        return null;
    }

    public void test(String name) {
        logger.info(getLedger().getContext().getId());
        //logger.info("invoke in test");
    }
}