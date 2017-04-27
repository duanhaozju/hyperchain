package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

import java.util.List;

/**
 * Created by Think on 4/27/17.
 */
public class Contract3 extends ContractBase {
    private static final Logger logger = Logger.getLogger(Contract3.class.getSimpleName());

    public synchronized void send(){

    }

    @Override
    public ByteString Query(String funcName, List<String> args) {
        return null;
    }

    public Contract3() {
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
    public boolean Invoke(String funcName, List<String> args) {
        return false;
    }
}
