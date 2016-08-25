# hyperChain Aplha

HyperChain 原型实现 

内部迭代版本: alpha 0.1
change log:
- alpha 0.1 完成了基本的节点连接数据同步和节点信息同步　2016/08/17 
- alpha 0.2 TODO
Todo List:
- [x] 加入block/chain结构
- [x] 加入`getBalance`功能
- [x] 加入节点和账户地址
- [ ] 加入`merktree`
- [x] 加入`checktransaction` 交易校验

##项目设置
**go path**
请将项目克隆到`$GOPATH/src`目录下，这样能够确保项目能够正常运行

##克隆项目
`git clone git@git.hyperchain.cn:chenquan/hyperchain-alpha.git`

## 修复依赖
`godep restore`

Ｐ.S. 在修复依赖的时候会有一些无法下载的问题，如果有pkg/golang.org/x/sys/unix 而且有go文件就不需要管，直接就可以正常运行
