/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.security;

import java.util.Map;

/**
 * Created by Think on 4/26/17.
 */
public class Rule {
    private Map<String, Map<String, Map<String, String>>> interfaceRule;
    private Map<String, Map<String, Map<String, String>>> superClassRule;
    private Map<String, Map<String, Map<String, String>>> memberVariableRule;
    private Map<String, Map<String, Map<String, String>>> methodDeclare;
    private Map<String, Map<String, Map<String, String>>> methodVariable;
    private Map<String, Map<String, Map<String, String>>> methodInstruction;

    public Map<String, Map<String, Map<String, String>>> getInterfaceRule() {
        return interfaceRule;
    }

    public void setInterfaceRule(Map<String, Map<String, Map<String, String>>> interfaceRule) {
        this.interfaceRule = interfaceRule;
    }

    public Map<String, Map<String, Map<String, String>>> getSuperClassRule() {
        return superClassRule;
    }

    public void setSuperClassRule(Map<String, Map<String, Map<String, String>>> superClassRule) {
        this.superClassRule = superClassRule;
    }

    public Map<String, Map<String, Map<String, String>>> getMemberVariableRule() {
        return memberVariableRule;
    }

    public void setMemberVariableRule(Map<String, Map<String, Map<String, String>>> memberVariableRule) {
        this.memberVariableRule = memberVariableRule;
    }

    public Map<String, Map<String, Map<String, String>>> getMethodDeclare() {
        return methodDeclare;
    }

    public void setMethodDeclare(Map<String, Map<String, Map<String, String>>> methodDeclare) {
        this.methodDeclare = methodDeclare;
    }

    public Map<String, Map<String, Map<String, String>>> getMethodVariable() {
        return methodVariable;
    }

    public void setMethodVariable(Map<String, Map<String, Map<String, String>>> methodVariable) {
        this.methodVariable = methodVariable;
    }

    public Map<String, Map<String, Map<String, String>>> getMethodInstruction() {
        return methodInstruction;
    }

    public void setMethodInstruction(Map<String, Map<String, Map<String, String>>> methodInstruction) {
        this.methodInstruction = methodInstruction;
    }
}