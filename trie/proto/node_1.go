package trie

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"strings"
)

func (n *Node) SetVal(node interface{}) error {
	switch value := node.(type) {
	case *ShortNode:
		n.Val = &Node_Shortnode{
			Shortnode: value,
		}
		return nil
	case *FullNode:
		n.Val = &Node_Fullnode{
			Fullnode: value,
		}
		return nil
	case *ValueNode:
		n.Val = &Node_Valuenode{
			Valuenode: value,
		}
		return nil
	case *BlankNode:
		n.Val = &Node_Blanknode{}
		return nil
	case *HashNode:
		n.Val = &Node_Hashnode{
			Hashnode: value,
		}
		return nil
	default:
		return &InvalidNodeTypeError{}
	}
}

func (n *ShortNode) SetVal(node interface{}) error {
	local := &Node{}
	err := local.SetVal(node)
	n.Val = local
	return err
}

func (n *FullNode) SetChild(node interface{}, index int) error {
	if n.Children == nil {
		n.Children = make([]*Node, 17)
		// set each child as a blanknode
		n.ResetChildren()
	}
	if index >= 0 && index <= 16 {
		if index == 16 {
			if _, ok := node.(ValueNode); !ok {
				if _, ok1 := node.(BlankNode); !ok1 {
					return &InvalidNodeTypeError{}
				}
			}
		}
		local := &Node{}
		err := local.SetVal(node)
		n.Children[index] = local
		return err
	} else {
		return &InvalidNodeIndexError{}
	}
	return nil
}

func (n *FullNode) ResetChildren() {
	for index := 0; index < 17; index++ {
		local := &Node{}
		local.Val = &Node_Blanknode{
			Blanknode: &BlankNode{},
		}
		n.Children[index] = local
	}
}
func (n Node) cache() ([]byte, bool) {
	if x, ok := n.GetVal().(*Node_Shortnode); ok {
		return (*x.Shortnode).cache()
	}
	if x, ok := n.GetVal().(*Node_Fullnode); ok {
		return (*x.Fullnode).cache()
	}
	if x, ok := n.GetVal().(*Node_Hashnode); ok {
		return (*x.Hashnode).cache()
	}
	if x, ok := n.GetVal().(*Node_Valuenode); ok {
		return (*x.Valuenode).cache()
	}
	return nil, true

}

// Cache accessors to retrieve precalculated values (avoid lengthy type switches).
func (n FullNode) cache() ([]byte, bool)  { return n.Hash, n.Dirty }
func (n ShortNode) cache() ([]byte, bool) { return n.Hash, n.Dirty }
func (n HashNode) cache() ([]byte, bool)  { return nil, true }
func (n ValueNode) cache() ([]byte, bool) { return nil, true }

func mustDecodeNode(buf []byte) interface{} {
	n, err := decodeNode(buf)
	if err != nil {
		panic(fmt.Sprintf("node  %v", err))
	}
	return n
}

// decodeNode parses the RLP encoding of a trie node.
func decodeNode(buf []byte) (interface{}, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	// TODO
	// not error return but mismatch happened
	tmp := &FullNode{}
	err := proto.Unmarshal(buf, tmp)
	if err == nil {
		return tmp, err
	}
	tmp1 := &ShortNode{}
	err = proto.Unmarshal(buf, tmp1)
	if err == nil {
		return tmp1, err
	}
	return Node{}, err
}

// wraps a decoding error with information about the path to the
// invalid child node (for debugging encoding issues).
type decodeError struct {
	what  error
	stack []string
}

func wrapError(err error, ctx string) error {
	if err == nil {
		return nil
	}
	if decErr, ok := err.(*decodeError); ok {
		decErr.stack = append(decErr.stack, ctx)
		return decErr
	}
	return &decodeError{err, []string{ctx}}
}

func (err *decodeError) Error() string {
	return fmt.Sprintf("%v (decode path: %s)", err.what, strings.Join(err.stack, "<-"))
}
