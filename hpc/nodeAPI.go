package hpc

type PublicNodeAPI struct{

}

func NewPublicNodeAPI() *PublicNodeAPI{
	return &PublicNodeAPI{}
}

// TODO 得到节点状态
func (node *PublicNodeAPI) GetNodes() error{
	return nil
}