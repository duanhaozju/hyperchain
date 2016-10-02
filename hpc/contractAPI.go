package hpc

import "hyperchain/common"

type PublicContractAPI struct {}

func NewPublicContractAPI() *PublicContractAPI {
	return &PublicContractAPI{}
}

//type CompileCode struct{
//	Abi []string
//	Bin []string
//}
//
//// ComplieContract complies contract to ABI
//func (contract *PublicContractAPI) ComplieContract(ct string) (*CompileCode,error){
//
//	abi, bin, err := compiler.CompileSourcefile(ct)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return &CompileCode{
//		Abi: abi,
//		Bin: bin,
//	}, nil
//}
func (contract *PublicContractAPI) DeployContract() common.Hash{
	return common.Hash{}
}
