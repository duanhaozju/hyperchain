/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.contract;

import org.apache.log4j.Logger;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

//ContractClassLoader used to load the contract class at runtime.
public class ContractClassLoader extends ClassLoader{
    private static final Logger logger = Logger.getLogger(ContractClassLoader.class.getSimpleName());
    private String contractDir;

    public ContractClassLoader(String contractDir) {
        this.contractDir = contractDir;
    }

    public Class<?> load(String name) throws  ClassNotFoundException {
        return super.loadClass(name, true);
    }

    @Override
    protected Class<?> findClass(String name) throws ClassNotFoundException {
        String fileName = contractDir + "/" + name.replaceAll("\\.", "/") + ".class";
        logger.info(fileName);
        try {
            byte[] bytes;
            bytes = loadClassContent(fileName);
            return defineClass(name, bytes, 0, bytes.length);
        } catch (Exception e) {
            e.printStackTrace();
            throw new ClassNotFoundException();
        }
    }

    private void loadClassFromFile(File file) throws Exception{
        if (file != null) {
            if(file.isDirectory()) {
                String [] subFileNames = file.list();
                for(String fileName: subFileNames) {
                    File subFile = new File(fileName);
                    loadClassFromFile(subFile);
                }
            }else {
                if(file.canRead()) {
                    Path p = file.toPath();
                    byte[] data = Files.readAllBytes(p);
                }else {
                    throw new RuntimeException("can not read file " + file.getName());
                }
            }
        }else {
            throw new NullPointerException("file is null");
        }
    }

    //read class data from disk
    private byte[] loadClassContent(String spath) throws Exception {
        byte[] classData;
        if(spath != null && spath.length() > 0){
            Path path = Paths.get(spath);
            classData = Files.readAllBytes(path);
        }else {
               throw new IOException("Invalid path");
        }
        return classData;
    }
}
