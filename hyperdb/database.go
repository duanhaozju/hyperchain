package hyperdb

import "hyperchain-alpha/core/node"

type NodeHandler interface{
	NodeSave(node node.Node) error
	NodeGet(key string)
}

type Saver interface{
	Save()
}

type DB interface {
	Save(p *interface{}) bool
	Get(p *interface{}) * interface{}
	Del(key string) bool
}

//type db struct {
//
//}
//func (ldb leveldb)Save(s Saver) bool{
//	ret := s.Save()
//	return ret
//}
//func (ldb leveldb)Get(key string) *interface{}{
//	return nil
//}
//
//type mem struct {
//
//}
//
//func (m mem)Save(s Saver) bool{
//	ret := s.Save()
//	return ret
//}
//func (m mem)Get(p * interface{}) *interface{}{
//	key := p.(int)
//
//	return nil
//}
//
//
//trans = ldb.DBGet("123").(transaction)
//
//
//type mysql struct {
//	DBSave(s Saver) bool
//	DBGet(key string) * interface{}
//	DBDel(key string) bool
//}
//
//
//type Mem interface {
//
//}
//func main() {
//	db := DB{}
//	db.DBSave(trans)
//	db.get()
//}

