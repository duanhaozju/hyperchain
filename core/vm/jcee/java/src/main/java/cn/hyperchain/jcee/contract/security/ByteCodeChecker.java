/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.security;

import jdk.internal.org.objectweb.asm.ClassReader;
import jdk.internal.org.objectweb.asm.tree.ClassNode;
import jdk.internal.org.objectweb.asm.tree.FieldNode;
import jdk.internal.org.objectweb.asm.tree.LocalVariableNode;
import jdk.internal.org.objectweb.asm.tree.MethodNode;

import java.util.List;
import java.util.Map;

/**
 * Created by wangxiaoyi on 2017/4/25.
 */
public class ByteCodeChecker implements Checker {

    @Override
    public boolean pass(byte[] clazz) {
        ClassReader reader = new ClassReader(clazz);
        ClassNode cn = new ClassNode();
        reader.accept(cn, 0);

        if (interfaceIsNotAllowed(cn) || superClassIsNotAllowed(cn)) {
            return false;
        }

        if (!memberVariablesIsFinal(cn) || memberVariablesIsNotAllowed(cn)) {
            return false;
        }

        if (methodDeclareIsNotAllowed(cn) || localVariableIsNotAllowed(cn)) {
            return false;
        }
        return true;
    }

    public boolean memberVariablesIsFinal(ClassNode cn) {
        List<FieldNode> fields = cn.fields;
        for (FieldNode fieldNode : fields) {
            boolean finalFlag = containsOpcode(Rule.allowedRules, fieldNode.access);
            if (!finalFlag) {
                return false;
            }
        }
        return true;
    }

    public boolean memberVariablesIsNotAllowed(ClassNode cn) {
        List<FieldNode> fields = cn.fields;
        for (FieldNode fieldNode : fields) {
            String desc = fieldNode.desc;
            boolean notAllowedFlag = containsKeyWords(Rule.notAllowedRules, desc);
            if (notAllowedFlag) {
                return true;
            }
        }
        return false;
    }

    public boolean interfaceIsNotAllowed(ClassNode cn) {
        List<String> interfaces = cn.interfaces;
        for (String interfacee : interfaces) {
            boolean notAllowedFlag = containsKeyWords(Rule.notAllowedRules, interfacee);
            if (notAllowedFlag) {
                return true;
            }
        }
        return false;
    }

    public boolean superClassIsNotAllowed(ClassNode cn) {
        String superClass = cn.superName;
        return containsKeyWords(Rule.notAllowedRules, superClass);
    }

    public boolean methodDeclareIsNotAllowed(ClassNode cn) {
        List<MethodNode> methodNodes = cn.methods;
        for (MethodNode methodNode : methodNodes) {
            boolean syncFlag = containsOpcode(Rule.notAllowedRules, methodNode.access);
            if (syncFlag) {
                return true;
            }
        }
        return false;
    }

    public boolean localVariableIsNotAllowed(ClassNode cn) {
        List<MethodNode> methodNodes = cn.methods;
        for (MethodNode methodNode : methodNodes) {
            List<LocalVariableNode> LocalVariableNodes = methodNode.localVariables;
            for (LocalVariableNode localVariableNode : LocalVariableNodes) {
                String desc = localVariableNode.desc;
                boolean notAllowedFlag = containsKeyWords(Rule.notAllowedRules, desc);
                if (notAllowedFlag) {
                    return true;
                }
            }
        }
        return false;
    }

    public boolean containsOpcode(Map<String, String> rules, int access) {
        while (access > 0) {
            double temp = Math.pow(2, getExp(access));
            if (rules.containsKey(String.valueOf((int)temp))) {
                System.out.println(rules.get(String.valueOf((int)temp)));
                return true;
            }
            access -= temp;
        }
        System.out.println("not in " + (rules.equals(Rule.allowedRules)?"allowedRules":"notAllowedRules"));
        return false;
    }

    public boolean containsKeyWords(Map<String, String> rules, String desc) {
        desc = desc.replaceAll(";", "");
        String[] keyWords = desc.split("/");
        for (String keyWord : keyWords) {
            if (rules.containsKey(keyWord)) {
                System.out.println(rules.get(keyWord));
                return true;
            }
        }
        return false;
    }

    public int getExp(int num) {
        int exp = 0;
        while (num > 1) {
            num /= 2;
            exp++;
        }
        return exp;
    }
}