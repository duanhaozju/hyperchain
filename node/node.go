package node

type Node interface {
	Start()


}

type GrpcNode struct {
	Id int
}

func (self *GrpcNode)Start()  {

}