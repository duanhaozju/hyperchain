# HYPERCHAIN

>HyperChain Proto  Internal version: alpha 0.3.0

## CHANGE LOG:

- alpha 0.1 Finished the basic node connection and information sync　2016/08/17 
- alpha 0.2 Finished a single thread proto system.
- alpha 0.3 to finish a multi-thread proto system, which based on PBFT.

## LATEST BRANCH

`develop`

## SETTINGS

**go path**

clone project in to `$GOPATH/src` to ensure running rightly

## CLONE

`git@git.hyperchain.cn:hyperchain/hyperchain.git`

## DEPENDENCY

`go get -u github.com/kardianos/govendor`

`govendor sync`

## BUILD

`govendor build`

## QUICK START 
Ubuntu:

`bash ubuntu.sh`

Mac:

`bash mac.sh`

## LOGGER PACKAGE USAGE
- add those lines before your codes:

```
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
}
```

> Note: `MustGetLogger` 的参数需要本文件的包名,是相对于`hyperchain`文件夹的包名,例如 `p2p/peer` , `consensus/pbtf` etc.
> IMPORTANT: 在一个包中仅仅需要执行一次上述代码,一般将上述代码放在一个单独文件中或者放在包的第一个文件中.在Test文件中请不要再重新声明!!


- log level
```
   High CRITICAL
        ERROR
        WARNING
        NOTICE
        INFO
   Low  DEBUG
```