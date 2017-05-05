package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.contract.ContractBase;
import com.google.protobuf.ByteString;
import com.google.protobuf.DoubleValue;
import org.apache.log4j.Logger;

import java.util.List;

/**
 * Created by Think on 4/27/17.
 */
public class NormalContract extends ContractBase {

        private static final Logger logger = Logger.getLogger(NormalContract.class.getSimpleName());

        /**
         * Invoke smart contract method
         *
         * @param funcName function name user defined in contract
         * @param args     arguments of funcName
         */
        @Override
        public boolean Invoke(String funcName, List<String> args) {
            switch (funcName) {
                case "issue":
                    return issue(args);
                case "transfer":
                    return transfer(args);
                default:
                    logger.error("method " + funcName  + " not found!");

            }
            return false;
        }

        /**
         * Query data stored in the smart contract
         *
         * @param funcName function name
         * @param args     function related arguments
         * @return the query result
         */
        @Override
        public ByteString Query(String funcName, List<String> args) {
            switch (funcName) {
                case "getAccountBalance":
                    return getAccountBalance(args);
                default:
                    logger.error("method " + funcName  + " not found!");
            }
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
                byte[] balance = ledger.get(accountA.getBytes());

                if(balance != null) {
                    double balanceA = DoubleValue.parseFrom(balance).getValue();
                    double balanceB = DoubleValue.parseFrom(ledger.get(accountB.getBytes())).getValue();
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
                byte[] data = ledger.get(args.get(0).getBytes());
                logger.info(new String(data));
                if (data != null) {
                    return ByteString.copyFrom(data);
                }else {
                    logger.error("getAccountBalance error");
                }
            }catch (Exception e) {
                e.printStackTrace();
            }
            return null;
        }
}
