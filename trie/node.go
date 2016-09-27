package trie

import (
	"encoding/json"
	"fmt"
	"hyperchain/common"
	"io"
	"strings"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type node interface {
	cache() (hashNode, bool)
}

type (
	fullNode struct {
		Children [17]node `json:"Children"` // Actual trie node data to encode/decode (needs custom encoder)
		hash     hashNode `json:"hash"`     // Cached hash of the node to prevent rehashing (may be nil)
		dirty    bool     `json:"dirty"`    // Cached flag whether the node's new or already stored
	}
	shortNode struct {
		Key   []byte   `json:"Key"`
		Val   node     `json:"Val"`
		hash  hashNode `json:"hash"`  // Cached hash of the node to prevent rehashing (may be nil)
		dirty bool     `json:"dirty"` // Cached flag whether the node's new or already stored
	}
	hashNode  []byte
	valueNode []byte

	memFullNode struct {
		Content [17][]byte `json:"Content"`
		Type    string     `json:"Type"`
	}
	memShortNode struct {
		Content []byte `json:"Content"`
		Type    string `json:"Type"`
		Key     []byte `json:"Key"`
	}
)

// Cache accessors to retrieve precalculated values (avoid lengthy type switches).
func (n fullNode) cache() (hashNode, bool)  { return n.hash, n.dirty }
func (n shortNode) cache() (hashNode, bool) { return n.hash, n.dirty }
func (n hashNode) cache() (hashNode, bool)  { return nil, true }
func (n valueNode) cache() (hashNode, bool) { return nil, true }

func mustDecodeNode(hash, buf []byte) node {
	n, err := decodeNode(hash, buf)
	if err != nil {
		panic(fmt.Sprintf("node %x: %v", hash, err))
	}
	return n
}

// Deserialize the node
func decodeNode(hash, buf []byte) (node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	var n interface{}
	err := json.Unmarshal(buf, &n)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	if val, ok := n.(map[string]interface{}); ok {
		for k, v := range val {
			if strings.Compare(k, "Type") != 0 {
				continue
			}
			str := v.(string)
			switch str {
			case "full":
				return decodeFull(hash, buf)
			case "short":
				return decodeShort(hash, buf)
			default:
				return nil, fmt.Errorf("decode error")
			}
		}
	} else {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	return nil, nil
}

func decodeShort(hash, buf []byte) (node, error) {
	var mShortNode memShortNode
	err := json.Unmarshal(buf, &mShortNode)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	key := compactDecode(mShortNode.Key)
	if key[len(key)-1] == 16 {
		// value node
		return shortNode{key, valueNode(mShortNode.Content), hash, false}, nil
	}
	r, err := decodeRef(mShortNode.Content)
	if err != nil {
		return nil, wrapError(err, "val")
	}
	return shortNode{key, r, hash, false}, nil
}

func decodeFull(hash, buf []byte) (node, error) {
	var mFullNode memFullNode
	err := json.Unmarshal(buf, &mFullNode)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}

	n := fullNode{hash: hash}
	for i := 0; i < 16; i++ {
		cld, err := decodeRef(mFullNode.Content[i])
		if err != nil {
			return n, wrapError(err, fmt.Sprintf("[%d]", i))
		}
		n.Children[i] = cld
	}
	if len(mFullNode.Content[16]) > 0 {
		n.Children[16] = valueNode(mFullNode.Content[16])
	}
	return n, nil
}

const hashLen = len(common.Hash{})

func decodeRef(buf []byte) (node, error) {
	if len(buf) == 0 {
		return nil, nil
	} else if len(buf) == 32 {
		return hashNode(buf), nil
	} else {
		// TODO embedded node reference
		return nil, nil
	}
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
