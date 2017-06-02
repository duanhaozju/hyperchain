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
    private Rule rule;
    private String systemRulePath = "./hyperjvm/config/systemSecurityRule.yaml";
    private String allowedRule = "allowedRule";
    private String notAllowedRule = "notAllowedRule";
    private String keywordType = "keyword";
    private String mustKeywordType = "mustKeyword";
    private String classType = "class";
    private String opCodeType = "opCode";
    private String classSuffix = ".class";
    private String contractTemplate = "cn/hyperchain/jcee/contract/ContractTemplate";

    public ByteCodeChecker (String userRulePath) {
        YAMLFactory yamlFactory = new YAMLFactory();
        ObjectMapper objectMapper = new ObjectMapper(yamlFactory);
        try {
            this.rule = objectMapper.readValue(new File(userRulePath), Rule.class);
        } catch (IOException e) {
            logger.error(e.getMessage());
        }
    }

    public ByteCodeChecker () {
        YAMLFactory yamlFactory = new YAMLFactory();
        ObjectMapper objectMapper = new ObjectMapper(yamlFactory);
        try {
            this.rule = objectMapper.readValue(new File(systemRulePath), Rule.class);
        } catch (IOException e) {
            logger.error(e.getMessage());
        }
    }

    private boolean judge(ClassNode cn) {
        if (isContract(cn)) {
            if (interfaceIsOk(rule, cn) && superClassIsOk(rule, cn) &&
                    memberVariableKeyWordIsOk(rule, cn) && memberVariableClassIsOk(rule, cn) &&
                    methodDeclareIsOk(rule, cn) && methodVariableIsOk(rule, cn) &&
                    methodInstructionOpCodeIsOk(rule, cn) && methodInstructionDescIsOk(rule, cn)) {
                return true;
            }
        } else {
            if (interfaceIsOk(rule, cn) && superClassIsOk(rule, cn) &&
                    memberVariableClassIsOk(rule, cn) &&
                    methodDeclareIsOk(rule, cn) && methodVariableIsOk(rule, cn) &&
                    methodInstructionOpCodeIsOk(rule, cn) && methodInstructionDescIsOk(rule, cn)) {
                return true;
            }
        }
        return false;
    }

    @Override
    public boolean pass(byte[] clazz) {
        ClassReader reader = new ClassReader(clazz);
        ClassNode cn = new ClassNode();
        reader.accept(cn, 0);
        return judge(cn);
    }

    @Override
    public boolean pass(String absoluteClassPath) {
        File file = new File(absoluteClassPath);
        try {
            ClassReader reader = new ClassReader(new FileInputStream(file));
            ClassNode cn = new ClassNode();
            reader.accept(cn, 0);
            return judge(cn);
        } catch (IOException e) {
            logger.info("read file fail!");
            logger.error(e.getMessage());
        }
        return false;
    }

    @Override
    public boolean passAll(String absoluteDirPath) {
        List<String> classFiles = getAllClassFile(absoluteDirPath);
        for (String classFile: classFiles) {
            if (!pass(classFile)) {
                logger.warn(classFile + " is not safe!");
                return false;
            }
            logger.info(classFile + " is safe!");
        }
        return true;
    }

    private boolean interfaceIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            logger.warn("no rule found!");
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getInterfaceRule().get(allowedRule).get(classType);
        Map<String, String> classOfNotAllowedRule = rule.getInterfaceRule().get(notAllowedRule).get(classType);
        List<String> interfaces = cn.interfaces;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.debug("check interface true!");
            return true;
        } else if (classOfAllowedRule != null) {
            for (String interfacee : interfaces) {
                boolean flag = containsClass(classOfAllowedRule, interfacee);
                if (!flag) {
                    logger.info("check interface false!" + " interface name:" + interfacee);
                    return false;
                }
            }
            logger.debug("check interface true!");
            return true;
        } else {
            for (String interfacee : interfaces) {
                boolean flag = containsClass(classOfNotAllowedRule, interfacee);
                if (flag) {
                    logger.info("check interface false!" + " interface name:" + interfacee);
                    return false;
                }
            }
            logger.debug("check interface true!");
            return true;
        }
    }

    private boolean superClassIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            logger.warn("no rule found!");
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getSuperClassRule().get(allowedRule).get(classType);
        Map<String, String> classOfNotAllowedRule = rule.getSuperClassRule().get(notAllowedRule).get(classType);
        String superClass = cn.superName;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.debug("check super class true!");
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

    private boolean isContract(ClassNode cn) {
        String superClass = cn.superName;
        if (contractTemplate.equals(superClass)) {
            return true;
        }
        return false;
    }

    private boolean memberVariableKeyWordIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            logger.warn("no rule found!");
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
            logger.debug("check class member variable keyword true!");
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
            logger.debug("check class member variable keyword true!");
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
            logger.debug("check class member variable keyword true!");
            return true;
        }
    }

    private boolean memberVariableClassIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            logger.warn("no rule found!");
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getMemberVariableRule().get(allowedRule).get(classType);
        Map<String, String> classOfNotAllowedRule = rule.getMemberVariableRule().get(notAllowedRule).get(classType);
        List<FieldNode> fields = cn.fields;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.debug("check class member variable class true!");
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
            logger.debug("check class member variable class true!");
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
            logger.debug("check class member variable class true!");
            return true;
        }
    }

    private boolean methodDeclareIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            logger.warn("no rule found!");
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
            logger.debug("check method declare true!");
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
            logger.debug("check method declare true!");
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
            logger.debug("check method declare true!");
            return true;
        }
    }

    private boolean methodVariableIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            logger.warn("no rule found!");
            return true;
        }
        Map<String, String> classOfAllowedRule = rule.getMethodVariable().get(allowedRule).get(classType);
        Map<String, String> classOfNotAllowedRule = rule.getMethodVariable().get(notAllowedRule).get(classType);
        List<MethodNode> methodNodes = cn.methods;
        if (classOfAllowedRule == null && classOfNotAllowedRule == null) {
            logger.debug("check method variable true!");
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
            logger.debug("check method variable true!");
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
            logger.debug("check method variable true!");
            return true;
        }
    }

    private boolean methodInstructionOpCodeIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            logger.warn("no rule found!");
            return true;
        }
        Map<String, String> opCodeOfAllowedRule = rule.getMethodInstruction().get(allowedRule).get(opCodeType);
        Map<String, String> opCodeOfNotAllowedRule = rule.getMethodInstruction().get(notAllowedRule).get(opCodeType);
        List<MethodNode> methodNodes = cn.methods;
        if (opCodeOfAllowedRule == null && opCodeOfNotAllowedRule == null) {
            logger.debug("check method instruction opcode true!");
            return true;
        } else if (opCodeOfAllowedRule != null) {
            for (MethodNode methodNode : methodNodes) {
                InsnList insnList = methodNode.instructions;
                for (int i = 0; i < insnList.size(); ++i) {
                    int opCodeNum = insnList.get(i).getOpcode();
                    boolean flag = containsClass(opCodeOfAllowedRule, getOpCode(opCodeNum));
                    if (!flag) {
                        logger.info("check method instruction opcode false!" + " method name: " + methodNode.name + " method instruction opCode name: " + getOpCode(opCodeNum));
                        return false;
                    }
                }
            }
            logger.debug("check method instruction true!");
            return true;
        } else {
            for (MethodNode methodNode : methodNodes) {
                InsnList insnList = methodNode.instructions;
                for (int i = 0; i < insnList.size(); ++i) {
                    int opCodeNum = insnList.get(i).getOpcode();
                    boolean flag = containsClass(opCodeOfNotAllowedRule, getOpCode(opCodeNum));
                    if (flag) {
                        logger.info("check method instruction opcode false!" + " method name: " + methodNode.name + " method instruction opCode name: " + getOpCode(opCodeNum));
                        return false;
                    }
                }
            }
            logger.debug("check method instruction opcode true!");
            return true;
        }
    }

    private boolean methodInstructionDescIsOk(Rule rule, ClassNode cn) {
        if (rule == null) {
            logger.warn("no rule found!");
            return true;
        }
        Map<String, String> descOfAllowedRule = rule.getMethodInstruction().get(allowedRule).get(classType);
        Map<String, String> descOfNotAllowedRule = rule.getMethodInstruction().get(notAllowedRule).get(classType);
        List<MethodNode> methodNodes = cn.methods;
        if (descOfAllowedRule == null && descOfNotAllowedRule == null) {
            logger.debug("check method instruction desc true!");
            return true;
        } else if (descOfAllowedRule != null) {
            for (MethodNode methodNode : methodNodes) {
                InsnList insnList = methodNode.instructions;
                for (int i = 0; i < insnList.size(); ++i) {
                    if (insnList.get(i) instanceof MethodInsnNode) {
                        MethodInsnNode methodInsnNode = (MethodInsnNode)insnList.get(i);
                        boolean flag = containsClass(descOfAllowedRule, methodInsnNode.desc);
                        if (!flag) {
                            logger.info("check method instruction desc false!" + " method name: " + methodNode.name + " method instruction desc name: " + methodInsnNode.desc);
                            return false;
                        }
                    }
                }
            }
            logger.debug("check method instruction desc true!");
            return true;
        } else {
            for (MethodNode methodNode : methodNodes) {
                InsnList insnList = methodNode.instructions;
                for (int i = 0; i < insnList.size(); ++i) {
                    if (insnList.get(i) instanceof MethodInsnNode) {
                        MethodInsnNode methodInsnNode = (MethodInsnNode)insnList.get(i);
                        boolean flag = containsClass(descOfNotAllowedRule, methodInsnNode.desc);
                        if (flag) {
                            logger.info("check method instruction desc false!" + " method name: " + methodNode.name + " method instruction desc name: " + methodInsnNode.desc);
                            return false;
                        }
                    }
                }
            }
            logger.debug("check method instruction desc true!");
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