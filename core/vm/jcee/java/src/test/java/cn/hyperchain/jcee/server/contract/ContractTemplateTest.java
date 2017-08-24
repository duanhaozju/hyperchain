package cn.hyperchain.jcee.server.contract;

import cn.hyperchain.jcee.client.contract.ContractTemplate;
import cn.hyperchain.jcee.client.contract.examples.sb.src.SimulateBank;
import cn.hyperchain.jcee.client.mock.MockLedger;
import cn.hyperchain.jcee.common.ExecuteResult;
import com.carrotsearch.junitbenchmarks.BenchmarkOptions;
import com.carrotsearch.junitbenchmarks.BenchmarkRule;
import org.junit.Assert;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.TestRule;

import java.util.ArrayList;
import java.util.List;

public class ContractTemplateTest {

    @Rule
    public TestRule benchmarkRun = new BenchmarkRule();


    @Test
    @BenchmarkOptions(benchmarkRounds = 100000, warmupRounds = 0)
    public void invoke() throws Exception {
        ContractTemplate ct = new SimulateBank();
        ct.setLedger(new MockLedger());
        List<String> args = new ArrayList<>();
        args.add("bk-001");
        args.add("10000000");
        ExecuteResult result = ct.invoke("issue", args);
        Assert.assertEquals(true, result.isSuccess());
    }

    @Test
    @BenchmarkOptions(benchmarkRounds = 100000, warmupRounds = 0)
    public void invokeContract() throws Exception {
        SimulateBank sb = new SimulateBank();
        sb.setLedger(new MockLedger());
        List<String> args = new ArrayList<>();
        args.add("bk-001");
        args.add("10000000");
        ExecuteResult result = sb.issue(args);
        Assert.assertEquals(true, result.isSuccess());
    }
}