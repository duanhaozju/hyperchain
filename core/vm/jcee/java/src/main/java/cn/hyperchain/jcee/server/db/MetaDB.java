/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.db;

import cn.hyperchain.jcee.client.contract.ContractInfo;
import lombok.Setter;
import org.apache.log4j.Logger;
import org.yaml.snakeyaml.Yaml;

import java.io.FileWriter;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Paths;

/**
 * Created by wangxiaoyi on 2017/5/4.
 */
public class MetaDB {
    private static final Logger logger = Logger.getLogger(MetaDB.class);
    @Setter
    private String path = "./hyperjvm/meta/meta.yaml";
    private static ContractsMeta metaData;

    private static volatile  MetaDB db = new MetaDB();

    public static MetaDB getDb() {
        return db;
    }

    private MetaDB(){
        metaData = new ContractsMeta();
    }

    public ContractsMeta load() {
        ContractsMeta meta = null;
        Yaml yaml = new Yaml();
        try( InputStream in = Files.newInputStream( Paths.get(path))) {
            if (in.available() != 0) {
                meta = yaml.loadAs(in, ContractsMeta.class);
            }
            return meta;
        }catch (IOException ioe) {
            logger.error(ioe.toString());
        }
        this.metaData = meta;
        return meta;
    }

    public void store(ContractsMeta meta) {
        Yaml yaml = new Yaml();
        try {
            FileWriter writer = new FileWriter(path);
            yaml.dump(meta, writer);
        }catch (IOException ioe) {
               logger.error(ioe.getMessage());
        }
    }

    public void store(ContractInfo info) {
        if (metaData == null) {
            metaData = new ContractsMeta();
        }
        metaData.addContractInfo(info);
        store(metaData);
    }

    public void remove(ContractInfo info) {
        if (metaData != null) {
            metaData.removeContractInfo(info);
            store(metaData);
        }
    }
}
