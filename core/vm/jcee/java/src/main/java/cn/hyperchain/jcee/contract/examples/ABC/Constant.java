package cn.hyperchain.jcee.contract.examples.ABC;

import java.util.HashSet;

/**
 * Created by huhu on 2017/5/31.
 */
public class Constant {

    public  static final  HashSet<String> IDTypes = new HashSet<>();
    public  static final HashSet<String> accountTypes = new HashSet<>();
    public  static final HashSet<String> orderStates = new HashSet<>();
    public  static final HashSet<String> draftType = new HashSet<>();
    public  static final HashSet<String> transferFlags = new HashSet<>();
    public  static final HashSet<String> responseTypes = new HashSet<>();
    public  static final HashSet<String> resultTypes = new HashSet<>();
    public  static final HashSet<String> operationTypes = new HashSet<>();
    public  static final HashSet<String> orderUpdatable = new HashSet<>();

    {
        IDTypes.add("CT00");
        IDTypes.add("CT01");

        accountTypes.add("RC00");
        accountTypes.add("RC01");

        orderStates.add("VALID");
        orderStates.add("INVALID");

        draftType.add("AC01");
        draftType.add("AC02");

        transferFlags.add("EM00");
        transferFlags.add("EM01");

        responseTypes.add("SU00");
        responseTypes.add("SU01");

        resultTypes.add("SUCC");
        resultTypes.add("FAIL");

        orderUpdatable.add("YES");
        orderUpdatable.add("NO");

    }

}
