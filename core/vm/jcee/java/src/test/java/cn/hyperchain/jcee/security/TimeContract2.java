package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.client.contract.ContractTemplate;
import cn.hyperchain.jcee.client.ledger.AbstractLedger;

import java.text.SimpleDateFormat;
import java.util.List;

/**
 * Created by Think on 4/27/17.
 */
public class TimeContract2 extends ContractTemplate {

    public void error(){
        new SimpleDateFormat("yyyy-mm-dd");
    }

    public TimeContract2() {
        super();
    }

    @Override
    public void setLedger(AbstractLedger ledger) {
        super.setLedger(ledger);
    }

    @Override
    public AbstractLedger getLedger() {
        return super.getLedger();
    }

    @Override
    public void setOwner(String owner) {
        super.setOwner(owner);
    }

    @Override
    public String getOwner() {
        return super.getOwner();
    }

    @Override
    public String getCid() {
        return super.getCid();
    }

    @Override
    public void setCid(String cid) {
        super.setCid(cid);
    }

    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        return null;
    }
}
