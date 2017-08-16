/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.common;

import org.apache.log4j.Logger;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.file.*;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.Map;
import java.util.TreeMap;

/**
 * Created by wangxiaoyi on 2017/4/27.
 */
public class IOHelper {

    private static final Logger logger = Logger.getLogger(IOHelper.class.getSimpleName());

    /**
     * @param srcDir contract source code dir
     * @return the combined code content
     */
    public static synchronized byte[] readCode(String srcDir) {
        final ByteArrayOutputStream codes = new ByteArrayOutputStream();
        final Path path = Paths.get(srcDir);

        TreeMap<String, byte[]> classData = new TreeMap<>();

        FileVisitor<Path> fv = new SimpleFileVisitor<Path>() {
            @Override
            public FileVisitResult visitFile(Path file, BasicFileAttributes attrs)
                    throws IOException {
                if (file.getFileName().toString().endsWith(".class")){
                    logger.info("reading data from " + file.getFileName());
                    byte[] data = Files.readAllBytes(file);
//                    codes.write(data);
                    classData.put(file.getFileName().toString(), data);
                }
                return FileVisitResult.CONTINUE;
            }
        };

        try {
            Files.walkFileTree(path, fv);
            for (Map.Entry<String, byte[]> data : classData.entrySet()) {
                logger.debug(data.getKey());
                codes.write(data.getValue());
            }

        } catch (IOException e) {
            e.printStackTrace();
        }
        return codes.toByteArray();
    }
}
