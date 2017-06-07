package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.ledger.Result;
import com.google.protobuf.ByteString;
import com.google.protobuf.DoubleValue;
import org.apache.log4j.Logger;

import java.util.List;

/**
 * Created by Think on 4/27/17.
 */
public class NormalContract extends ContractTemplate {

        private static final Logger logger = Logger.getLogger(NormalContract.class.getSimpleName());

    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        return null;
    }

    //String account, double num
        private boolean issue(List<String> args) {
            if(args.size() != 2) {
                logger.error("args num is invalid");
            }
            logger.info("account: " + args.get(0));
            logger.info("num: " + args.get(1));

            boolean rs = ledger.put(args.get(0).getBytes(), args.get(1).getBytes());
            if(rs == false) {
                logger.error("issue func error");
            }
            return true;
        }

        //String accountA, String accountB, double num
        private boolean transfer(List<String> args) {
            if(args.size() != 3) {
                logger.error("args num is invalid");
            }
            try {
                String accountA = args.get(0);
                String accountB = args.get(1);
                double num = Double.valueOf(args.get(2));

                Result result = ledger.get(accountA);

                if(!result.isEmpty()) {
                    double balanceA = result.toDouble();
                    double balanceB = ledger.get(accountB.getBytes()).toDouble();
                    if (balanceA >= num) {
                        ledger.put(accountA.getBytes(), String.valueOf(balanceA - num).getBytes());
                        ledger.put(accountB.getBytes(), String.valueOf(balanceB + num).getBytes());
                    }
                }else {
                    logger.error("get account " + accountA  + " balance error");
                    return false;
                }

            }catch (Exception e) {
                e.printStackTrace();
            }

            return true;
        }

        private ByteString getAccountBalance(List<String> args) {
            if(args.size() != 1) {
                logger.error("args num is invalid");
            }
            try {
                Result result = ledger.get(args.get(0));
                if (!result.isEmpty()) {
                    return result.getValue();
                }else {
                    logger.error("getAccountBalance error");
                }
            }catch (Exception e) {
                e.printStackTrace();
            }
            return null;
        }
}
