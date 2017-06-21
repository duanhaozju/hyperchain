package cn.hyperchain.jcee.common;

/**
 * Created by wangxiaoyi on 2017/5/9.
 */
public class TestExecuteResult {

    public void testConstruct() {
        ExecuteResult<Integer> result = new ExecuteResult<>();
        result.setResult(1);

        ExecuteResult<Double> result1 = new ExecuteResult<>();
    }
}
