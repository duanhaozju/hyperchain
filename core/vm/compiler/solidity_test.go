package compiler

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

const solcVersion = "0.1.1"

var (
	source = `
contract test {
   /// @notice Will multiply ` + "`a`" + ` by 7.
   function multiply(uint a) returns(uint d) {
       return a * 7;
   }
}
`
	code = "0x6060604052602a8060106000396000f3606060405260e060020a6000350463c6888fa18114601a575b005b6007600435026060908152602090f3"
	info = `{"source":"\ncontract test {\n   /// @notice Will multiply ` + "`a`" + ` by 7.\n   function multiply(uint a) returns(uint d) {\n       return a * 7;\n   }\n}\n","language":"Solidity","languageVersion":"0.1.1","compilerVersion":"0.1.1","compilerOptions":"--binary file --json-abi file --natspec-user file --natspec-dev file --add-std 1","abiDefinition":[{"constant":false,"inputs":[{"name":"a","type":"uint256"}],"name":"multiply","outputs":[{"name":"d","type":"uint256"}],"type":"function"}],"userDoc":{"methods":{"multiply(uint256)":{"notice":"Will multiply ` + "`a`" + ` by 7."}}},"developerDoc":{"methods":{}}}`

	infohash = common.HexToHash("0x9f3803735e7f16120c5a140ab3f02121fd3533a9655c69b33a10e78752cc49b0")
)

func TestCompiler(t *testing.T) {
	sol, err := NewCompiler("")
	if err != nil {
		t.Skipf("solc not found: %v", err)
	}
	contracts, err := sol.Compile(source)
	if err != nil {
		t.Errorf("error compiling source. result %v: %v", contracts, err)
		return
	}

	if len(contracts) != 1 {
		t.Errorf("one contract expected, got %d", len(contracts))
	}

	if contracts["test"].Code != code {
		t.Errorf("wrong code, expected\n%s, got\n%s", code, contracts["test"].Code)
	}

}

func TestCompileError(t *testing.T) {
	sol, err := NewCompiler("")
	if err != nil || sol.version != solcVersion {
		t.Skip("solc not found: skip")
	} else if sol.Version() != solcVersion {
		t.Skip("WARNING: skipping due to a newer version of solc found (%v, expect %v)", sol.Version(), solcVersion)
	}
	contracts, err := sol.Compile(source[2:])
	if err == nil {
		t.Errorf("error expected compiling source. got none. result %v", contracts)
		return
	}
}

func TestNoCompiler(t *testing.T) {
	_, err := NewCompiler("/path/to/solc")
	if err != nil {
		t.Logf("solidity quits with error: %v", err)
	} else {
		t.Errorf("no solc installed, but got no error")
	}
}

func TestSaveInfo(t *testing.T) {
	var cinfo ContractInfo
	err := json.Unmarshal([]byte(info), &cinfo)
	if err != nil {
		t.Errorf("%v", err)
	}
	filename := path.Join(os.TempDir(), "solctest.info.json")
	os.Remove(filename)
	cinfohash, err := SaveInfo(&cinfo, filename)
	if err != nil {
		t.Errorf("error extracting info: %v", err)
	}
	got, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("error reading '%v': %v", filename, err)
	}
	if string(got) != info {
		t.Errorf("incorrect info.json extracted, expected:\n%s\ngot\n%s", info, string(got))
	}
	if cinfohash != infohash {
		t.Errorf("content hash for info is incorrect. expected %v, got %v", infohash.Hex(), cinfohash.Hex())
	}
}
