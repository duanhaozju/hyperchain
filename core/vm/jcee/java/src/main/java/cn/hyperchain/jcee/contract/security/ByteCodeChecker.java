/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.security;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import jdk.internal.org.objectweb.asm.ClassReader;
import jdk.internal.org.objectweb.asm.Opcodes;
import jdk.internal.org.objectweb.asm.tree.*;
import org.apache.log4j.Logger;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Created by wangxiaoyi on 2017/4/25.
 */

public class ByteCodeChecker implements Checker {

    private static final Logger logger = Logger.getLogger(ByteCodeChecker.class.getSimpleName());
    private Rule systemRule;
    private Rule userRule;
    private static final String systemRulePath = System.getProperty("user.dir") + "/src/main/resources/securityRule.yaml";
    private static final String allowedRule = "allowedRule";
    private static final String notAllowedRule = "notAllowedRule";
    private static final String keywordType = "keyword";
    private static final String mustKeywordType = "mustKeyword";
    private static final String classType = "class";
    private static final String opCodeType = "opCode";
    private static final String classSuffix = ".class";

    public ByteCodeChecker (String userRulePath) {
        YAMLFactory yamlFactory = new YAMLFactory();
        ObjectMapper objectMapper = new ObjectMapper(yamlFactory);
        try {
            this.systemRule = objectMapper.readValue(new File(systemRulePath), Rule.class);
            this.userRule = objectMapper.readValue(new File(userRulePath), Rule.class);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public ByteCodeChecker () {
        YAMLFactory yamlFactory = new YAMLFactory();
        ObjectMapper objectMapper = new ObjectMapper(yamlFactory);
        try {
            this.systemRule = objectMapper.readValue(new File(systemRulePath), Rule.class);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    @Override
    public boolean pass(byte[] clazz) {
        ClassReader reader = new ClassReader(clazz);
        ClassNode cn = new ClassNode();
        reader.accept(cn, 0);

        if (interfaceIsOk(systemRule, cn) && superClassIsOk(systemRule, cn) &&
                memberVariableKeyWordIsOk(systemRule, cn) && memberVariableClassIsOk(systemRule, cn) &&
                methodDeclareIsOk(systemRule, cn) && methodVariableIsOk(systemRule, cn) &&
                methodInstructionIsOk(systemRule, cn) &&
                interfaceIsOk(userRule, cn) && superClassIsOk(userRule, cn) &&
                memberVariableKeyWordIsOk(userRule, cn) && memberVariableClassIsOk(userRule, cn) &&
                methodDeclareIsOk(userRule, cn) && methodVariableIsOk(userRule, cn) &&
                methodInstructionIsOk(userRule, cn)) {
            return true;
        }
        return false;
    }

    @Override
    public boolean pass(String absoluteClassPath) {
        File file = new File(absoluteClassPath);
        try {
            ClassReader reader = new ClassReader(new FileInputStream(file));
            ClassNode cn = new ClassNode();
            reader.accept(cn, 0);
            if (interfaceIsOk(systemRule, cn) && superClassIsOk(systemRule, cn) &&
                    memberVariableKeyWordIsOk(systemRule, cn) && memberVariableClassIsOk(systemRule, cn) &&
                    methodDeclareIsOk(systemRule, cn) && methodVariableIsOk(systemRule, cn) &&
                    methodInstructionIsOk(systemRule, cn) &&
                    interfaceIsOk(userRule, cn) && superClassIsOk(userRule, cn) &&
                    memberVariableKeyWordIsOk(userRule, cn) && memberVariableClassIsOk(userRule, cn) &&
                    methodDeclareIsOk(userRule, cn) && methodVariableIsOk(userRule, cn) &&
                    methodInstructionIsOk(userRule, cn)) {
                return true;
            }
        } catch (IOException e) {
            logger.info("read file fail!");
            e.printStackTrace();
        }
        return false;
    }

    @Override
    public boolean passAll(String absoluteDirPath) {
        List<String> classFiles = getAllClassFile(absoluteDirPath);
        for (String classFile: classFiles) {
            if (!pass(classFile)) {
                return false;
            }
        }
        return true;
    }

    private boolean interfaceIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getInterfaceRule().get(allowedRule).get(classType);
        Map<String, String> classOfNotAllowedRule = rule.getInterfaceRule().get(notAllowedRule).get(classType);
        List<String> interfaces = cn.interfaces;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.info("check interface true!");
            return true;
        } else if (classOfAllowedRule != null) {
            for (String interfacee : interfaces) {
                boolean flag = containsClass(classOfAllowedRule, interfacee);
                if (!flag) {
                    logger.info("check interface false!" + " interface name:" + interfacee);
                    return false;
                }
            }
            logger.info("check interface true!");
            return true;
        } else {
            for (String interfacee : interfaces) {
                boolean flag = containsClass(classOfNotAllowedRule, interfacee);
                if (flag) {
                    logger.info("check interface false!" + " interface name:" + interfacee);
                    return false;
                }
            }
            logger.info("check interface true!");
            return true;
        }
    }

    private boolean superClassIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getSuperClassRule().get(allowedRule).get(classType);
        Map<String, String> classOfNotAllowedRule = rule.getSuperClassRule().get(notAllowedRule).get(classType);
        String superClass = cn.superName;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.info("check super class true!");
            return true;
        } else if (classOfAllowedRule != null) {
            boolean flag = containsClass(classOfAllowedRule, superClass);
            logger.info("check super class " + String.valueOf(flag) + "!" + " super class name: " + superClass);
            return flag;
        } else {
            boolean flag = containsClass(classOfNotAllowedRule, superClass);
            logger.info("check super class " + String.valueOf(!flag) + "!" + " super class name: " + superClass);
            return !flag;
        }
    }

    private boolean memberVariableKeyWordIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            return true;
        }
        Map<String, String> classOfMustAllowedRule = rule.getMemberVariableRule().get(allowedRule).get(mustKeywordType);
        Map<String, String> classOfMustNotAllowedRule = rule.getMemberVariableRule().get(notAllowedRule).get(mustKeywordType);
        Map<String, String> classOfAllowedRule = rule.getMemberVariableRule().get(allowedRule).get(keywordType);
        Map<String, String> classOfNotAllowedRule = rule.getMemberVariableRule().get(notAllowedRule).get(keywordType);
        List<FieldNode> fields = cn.fields;

        if (classOfMustAllowedRule != null) {
            for (FieldNode fieldNode : fields) {
                boolean flag = containsAllOpcode(classOfMustAllowedRule, fieldNode.access);
                if (!flag) {
                    logger.info("check class member variable false!" + " class member variable name:" + fieldNode.name);
                    return false;
                }
            }
        }
        if (classOfMustNotAllowedRule != null) {
            for (FieldNode fieldNode : fields) {
                boolean flag = containsAllOpcode(classOfMustNotAllowedRule, fieldNode.access);
                if (flag) {
                    logger.info("check class member variable false!" + " class member variable name:" + fieldNode.name);
                    return false;
                }
            }
        }
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.info("check class member variable keyword true!");
            return true;
        }
        if (classOfAllowedRule != null) {
            for (FieldNode fieldNode : fields) {
                int access = fieldNode.access;
                while (access > 0) {
                    double temp = Math.pow(2, getExp(access));
                    boolean flag = containsOpcode(classOfAllowedRule, (int)temp);
                    if (!flag) {
                        logger.info("check class member variable false!" + " class member variable name:" + fieldNode.name);
                        return false;
                    }
                    access -= temp;
                }
            }
            logger.info("check class member variable keyword true!");
            return true;
        } else {
            for (FieldNode fieldNode : fields) {
                int access = fieldNode.access;
                while (access > 0) {
                    double temp = Math.pow(2, getExp(access));
                    boolean flag = containsOpcode(classOfNotAllowedRule, (int)temp);
                    if (flag) {
                        logger.info("check class member variable false!" + " class member variable name:" + fieldNode.name);
                        return false;
                    }
                    access -= temp;
                }
            }
            logger.info("check class member variable keyword true!");
            return true;
        }
    }

    private boolean memberVariableClassIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getMemberVariableRule().get(allowedRule).get(classType);
        Map<String, String> classOfNotAllowedRule = rule.getMemberVariableRule().get(notAllowedRule).get(classType);
        List<FieldNode> fields = cn.fields;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.info("check class member variable class true!");
            return true;
        } else if (classOfAllowedRule != null) {
            for (FieldNode fieldNode : fields) {
                String desc = fieldNode.desc;
                boolean flag = containsClass(classOfAllowedRule, desc);
                if (!flag) {
                    logger.info("check class member variable false!" + " class member variable name:" + fieldNode.name);
                    return false;
                }
            }
            logger.info("check class member variable class true!");
            return true;
        } else {
            for (FieldNode fieldNode : fields) {
                String desc = fieldNode.desc;
                boolean flag = containsClass(classOfNotAllowedRule, desc);
                if (flag) {
                    logger.info("check class member variable false!" + " class member variable name:" + fieldNode.name);
                    return false;
                }
            }
            logger.info("check class member variable class true!");
            return true;
        }
    }

    private boolean methodDeclareIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            return true;
        }
        Map<String, String> classOfMustAllowedRule = rule.getMethodDeclare().get(allowedRule).get(mustKeywordType);
        Map<String, String> classOfMustNotAllowedRule = rule.getMethodDeclare().get(notAllowedRule).get(mustKeywordType);
        Map<String, String> classOfAllowedRule = rule.getMethodDeclare().get(allowedRule).get(keywordType);
        Map<String, String> classOfNotAllowedRule = rule.getMethodDeclare().get(notAllowedRule).get(keywordType);
        List<MethodNode> methodNodes = cn.methods;

        if (classOfMustAllowedRule != null) {
            for (MethodNode methodNode : methodNodes) {
                boolean flag = containsAllOpcode(classOfMustAllowedRule, methodNode.access);
                if (!flag) {
                    logger.info("check method declare false!" + " method name: " + methodNode.name);
                    return false;
                }
            }
        }
        if (classOfMustNotAllowedRule != null) {
            for (MethodNode methodNode : methodNodes) {
                boolean flag = containsAllOpcode(classOfMustNotAllowedRule, methodNode.access);
                if (flag) {
                    logger.info("check method declare false!" + " method name: " + methodNode.name);
                    return false;
                }
            }
        }
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.info("check method declare true!");
            return true;
        }
        if (classOfAllowedRule != null) {
            for (MethodNode methodNode : methodNodes) {
                int access = methodNode.access;
                while (access > 0) {
                    double temp = Math.pow(2, getExp(access));
                    boolean flag = containsOpcode(classOfAllowedRule, (int)temp);
                    if (!flag) {
                        logger.info("check method declare false!" + " method name: " + methodNode.name);
                        return false;
                    }
                    access -= temp;
                }
            }
            logger.info("check method declare true!");
            return true;
        } else {
            for (MethodNode methodNode : methodNodes) {
                int access = methodNode.access;
                while (access > 0) {
                    double temp = Math.pow(2, getExp(access));
                    boolean flag = containsOpcode(classOfNotAllowedRule, (int)temp);
                    if (flag) {
                        logger.info("check method declare false!" + " method name: " + methodNode.name);
                        return false;
                    }
                    access -= temp;
                }
            }
            logger.info("check method declare true!");
            return true;
        }
    }

    private boolean methodVariableIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getMethodVariable().get(allowedRule).get(classType);
        Map<String, String> classOfNotAllowedRule = rule.getMethodVariable().get(notAllowedRule).get(classType);
        List<MethodNode> methodNodes = cn.methods;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.info("check method variable true!");
            return true;
        } else if (classOfAllowedRule != null) {
            for (MethodNode methodNode : methodNodes) {
                List<LocalVariableNode> LocalVariableNodes = methodNode.localVariables;
                for (LocalVariableNode localVariableNode : LocalVariableNodes) {
                    String desc = localVariableNode.desc;
                    boolean flag = containsClass(classOfAllowedRule, desc);
                    if (!flag) {
                        logger.info("check method variable false!" + " method variable name: " + localVariableNode.name);
                        return false;
                    }
                }
            }
            logger.info("check method variable true!");
            return true;
        } else {
            for (MethodNode methodNode : methodNodes) {
                List<LocalVariableNode> LocalVariableNodes = methodNode.localVariables;
                for (LocalVariableNode localVariableNode : LocalVariableNodes) {
                    String desc = localVariableNode.desc;
                    boolean flag = containsClass(classOfNotAllowedRule, desc);
                    if (flag) {
                        logger.info("check method variable false!" + " method variable name: " + localVariableNode.name);
                        return false;
                    }
                }
            }
            logger.info("check method variable true!");
            return true;
        }
    }

    private boolean methodInstructionIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getMethodInstruction().get(allowedRule).get(opCodeType);
        Map<String, String> classOfNotAllowedRule = rule.getMethodInstruction().get(notAllowedRule).get(opCodeType);
        List<MethodNode> methodNodes = cn.methods;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.info("check method instruction true!");
            return true;
        } else if (classOfAllowedRule != null) {
            for (MethodNode methodNode : methodNodes) {
                InsnList insnList = methodNode.instructions;
                for (int i = 0; i < insnList.size(); ++i) {
                    int opCodeNum = insnList.get(i).getOpcode();
                    boolean flag = containsClass(classOfAllowedRule, getOpCode(opCodeNum));
                    if (!flag) {
                        logger.info("check method instruction false!" + " method instruction name: " + getOpCode(opCodeNum));
                        return false;
                    }
                }
            }
            logger.info("check method instruction true!");
            return true;
        } else {
            for (MethodNode methodNode : methodNodes) {
                InsnList insnList = methodNode.instructions;
                for (int i = 0; i < insnList.size(); ++i) {
                    int opCodeNum = insnList.get(i).getOpcode();
                    boolean flag = containsClass(classOfNotAllowedRule, getOpCode(opCodeNum));
                    if (flag) {
                        logger.info("check method instruction false!" + " method name: " + methodNode.name + " method instruction name: " + getOpCode(opCodeNum));
                        return false;
                    }
                }
            }
            logger.info("check method instruction true!");
            return true;
        }
    }

    private boolean containsAllOpcode(Map<String, String> rule, int access) {
        Map<Integer, String> tempMap = new HashMap<>();
        for (Map.Entry<String, String> entry: rule.entrySet()) {
            tempMap.put(getOpCodeNum(entry.getKey()), entry.getValue());
        }
        while (access > 0) {
            double temp = Math.pow(2, getExp(access));
            tempMap.remove((int)temp);
            access -= temp;
        }
        if (tempMap.size() == 0) {
            return true;
        } else {
            return false;
        }
    }

    private boolean containsOpcode(Map<String, String> rule, int access) {
        if (rule.containsKey(getOpCode(access))) {
                return true;
        }
        return false;
    }

    private boolean containsClass(Map<String, String> rules, String desc) {
        for (Map.Entry<String, String> entry: rules.entrySet()) {
            if (desc.contains(entry.getKey())) {
                return true;
            }
        }
        return false;
    }

    private int getExp(int num) {
        int exp = 0;
        while (num > 1) {
            num /= 2;
            exp++;
        }
        return exp;
    }

    private List<String> getAllClassFile(String absoluteDirPath) {
        List<String> classFiles = new ArrayList<>();
        File dir = new File(absoluteDirPath);
        if (dir.exists() && dir.isDirectory()) {
            File[] files = dir.listFiles();
            for (int i = 0; i < files.length; ++i) {
                File file = files[i];
                if (file.isFile() && file.getName().endsWith(classSuffix)) {
                    classFiles.add(file.getAbsolutePath());
                } else if (file.isDirectory()) {
                    classFiles.addAll(getAllClassFile(file.getAbsolutePath()));
                }
            }
        }
        return classFiles;
    }

    private String getOpCode(int num) {
        switch (num) {
            case Opcodes.ACC_FINAL:
                return "final";
            case Opcodes.ACC_SYNCHRONIZED:
                return "synchronized";
            case Opcodes.MONITORENTER:
                return "monitorenter";
            case Opcodes.MONITOREXIT:
                return "monitorexit";
            case Opcodes.INVOKESTATIC:
                return "invokestatic";
            default:
                return "sorry";
        }
    }

    private int getOpCodeNum(String Opcode) {
        switch (Opcode) {
            case "public":
                return Opcodes.ACC_PUBLIC;
            case "private":
                return Opcodes.ACC_PRIVATE;
            case "protected":
                return Opcodes.ACC_PROTECTED;
            case "static":
                return Opcodes.ACC_STATIC;
            case "final":
                return Opcodes.ACC_FINAL;
            case "super":
                return Opcodes.ACC_SUPER;
            case "synchronized":
                return Opcodes.ACC_SYNCHRONIZED;
            case "volatile":
                return Opcodes.ACC_VOLATILE;
            default:
                return Integer.MIN_VALUE;
        }
    }
}