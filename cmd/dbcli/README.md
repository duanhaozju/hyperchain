DBCLI
---

dbcli is a tool to get data from leveldb, no hyperchian required. it can get block, transaction, chain and receipt struct data.

Version
---

dbcli version 1.0


Example
---

1. Get a block by block number

		./dbcli block getBlockByNumber -p ../../build/node1/namespaces/global/data/leveldb/blockchain --num 1

2. Get a block by block hash

		./dbcli block getBlockByHash -p ../../build/node1/namespaces/global/data/leveldb/blockchain --hash 6309a7b3ec79c7803bf66fa981e5664eac13b362543bd8ef2c08420e088e5126

3. Get a block hash by block number

		./dbcli block getBlockHashByNumber -p ../../build/node1/namespaces/global/data/leveldb/blockchain --num 1

4. Get the block chain and output it to a file

		./dbcli chain getChain -p ../../build/node1/namespaces/global/data/leveldb/blockchain -o chain.txt

5. Get a receipt by transaction hash

		./dbcli receipt getReceipt -p ../../build/node1/namespaces/global/data/leveldb/blockchain -hash 609282ad441db509e3dfb8a65fa8ecf115165f109c26575efff1aa4aaa379950

6. Get a transaction by transaction hash

		 ./dbcli transaction getTransactionByHash -p ../../build/node1/namespaces/global/data/leveldb/blockchain -hash 609282ad441db509e3dfb8a65fa8ecf115165f109c26575efff1aa4aaa379950

7. Get all transaction

		./dbcli transaction getAllTransaction -p ../../build/node1/namespaces/global/data/leveldb/blockchain

8. Get a transaction meta by transaction hash

		./dbcli transaction getTransactionMetaByHash -p ../../build/node1/namespaces/global/data/leveldb/blockchain -hash 609282ad441db509e3dfb8a65fa8ecf115165f109c26575efff1aa4aaa379950

9. Get all discard transaction

		./dbcli transaction getAllDiscardTransaction -p ../../build/node1/namespaces/global/data/leveldb/blockchain

10. Get a discard transaction by transaction hash

		./dbcli transaction getDiscardTransactionByHash -p ../../build/node1/namespaces/global/data/leveldb/blockchain -hash 609282ad441db509e3dfb8a65fa8ecf115165f109c26575efff1aa4aaa379950

Notice
---
if you have any question,

input `./dbcli --help` to get help about dbcli.

input `./dbcli block --help` to get help about block.

input `./dbcli block getBlockByNumber --help` to get help about getBlockByNumber.

others are similar.

Contact
---
shichaohao@hyperchain.cn