package holochain

import (
	"fmt"
	"strconv"
	"testing"
	"time"
	"github.com/google/uuid"
	"os"
	b58 "github.com/jbenet/go-base58"
	"crypto/ecdsa"
)

func TestNew(t *testing.T) {
	var key ecdsa.PrivateKey
	h := New("Joe",&key,"some/path")
	nID := string(uuid.NodeID());
	if (nID != string(h.Id.NodeID()) ) {
		t.Error("expected holochain UUID NodeID to be "+nID+" got",h.Id.NodeID())
	}
	if (h.Types[0] != "myData") {
		t.Error("data got:",h.Types)
	}
	if (h.agent != "Joe") {
		t.Error("agent got:",h.agent)
	}
	if (h.privKey != &key) {
		t.Error("key got:",h.privKey)
	}
	if (h.path != "some/path") {
		t.Error("path got:",h.path)
	}

}

func TestGenChain(t *testing.T) {
	err := GenChain()
	ExpectNoErr(t,err)
}

func TestInit(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {panic(err)}
	defer func() {os.Chdir(pwd)}()
	d := mkTestDir();
	defer func() {os.RemoveAll(d)}()

	if IsInitialized(d) != false {
		t.Error("expected no directory")
	}
	agent := "Fred Flintstone <fred@flintstone.com>"
	err = Init(d, Agent(agent))
	ExpectNoErr(t,err)

	if IsInitialized(d) != true {
		t.Error("expected initialized")
	}
	p := d+"/"+DirectoryName
	privP,err := UnmarshalPrivateKey(p, PrivKeyFileName)
	ExpectNoErr(t,err)

	pub2,err := UnmarshalPublicKey(p,PubKeyFileName)
	ExpectNoErr(t,err)

	if (fmt.Sprintf("%v",*pub2) != fmt.Sprintf("%v",privP.PublicKey)) {t.Error("expected pubkey match!")}

	a,err := readFile(p,AgentFileName)
	ExpectNoErr(t,err)
	if string(a) != agent {t.Error("expected "+agent+" got ",a)}

}

func TestGenDev(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {panic(err)}
	defer func() {os.Chdir(pwd)}()
	d := mkTestDir()
	defer func() {os.RemoveAll(d)}()
	agent := Agent("Herbert <h@bert.com>")
	err = Init(d,agent)
	ExpectNoErr(t,err)
	root := d+"/"+DirectoryName+"/"+"test"


	if err = IsConfigured(root); err == nil {
		t.Error("expected no dna got:",err)
	}

	_, err = Load(root)
	ExpectErrString(t,err,"open "+root+"/"+DNAFileName+": no such file or directory")

	h,err := GenDev(root)
	if err != nil {
		t.Error("expected no error got",err)
	}

	if err = IsConfigured(root); err != nil {
		t.Error(err)
	}

	lh, err := Load(root)
	if  err != nil {
		t.Error("Error parsing loading",err)
	}

	if (lh.Id != h.Id) {
		t.Error("expected matching ids!")
	}

	_,err = GenDev(root)
	ExpectErrString(t,err,"holochain: "+root+" already exists")


}

func TestNewEntry(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {panic(err)}
	defer func() {os.Chdir(pwd)}()
	d := mkTestDir();
	defer func() {os.RemoveAll(d)}()
	agent := Agent("Herbert <h@bert.com>")
	err = Init(d,agent)
	ExpectNoErr(t,err)
	root := d+"/"+DirectoryName
	n := "test"
	path := root+"/"+n
	h,err := GenDev(path)
	ExpectNoErr(t,err)
	myData := `{
"from": "Art"
"msg": "Hi there!"
}
`
	hash := b58.Decode("3vemK25pc5ewYtztPGYAdX39uXuyV13xdouCnZUr8RMA") // dummy link hash
	var link Hash
	copy(link[:],hash)

	now := time.Unix(1,1) // pick a constant time so the test will always work

	headerHash,header,err := h.NewEntry(now,"myData",link,myData)
	ExpectNoErr(t,err)

	if header.Time != now {t.Error("expected time:"+fmt.Sprintf("%v",now))}

	if header.Type != "myData" {t.Error("expected type myData")}

	// check the header link
	l :=  b58.Encode(header.HeaderLink[:])
	if l != "3vemK25pc5ewYtztPGYAdX39uXuyV13xdouCnZUr8RMA" {t.Error("expected header link, got",l)}

	// check the content link
	l =  b58.Encode(header.EntryLink[:])
	if l != "G4hiF3uvJhzimE4Tbyc4UgdUaznbm3vqbbH6G99SaMTL" {t.Error("expected entry hash, got",l)}

	// check the hash
	hash = headerHash[:]
	a := b58.Encode(hash)
	if a != "EdkgsdwazMZc9vJJgGXgbGwZFvy2Wa1hLCjngmkw3PbF" {
		t.Error("expected EdkgsdwazMZc9vJJgGXgbGwZFvy2Wa1hLCjngmkw3PbF got:",a)
	}

	// check the my signature of the entry
	pub,err := UnmarshalPublicKey(root,PubKeyFileName)
	ExpectNoErr(t,err)
	sig := header.MySignature
	hash = header.EntryLink[:]
	if !ecdsa.Verify(pub,hash,&sig.R,&sig.S) {t.Error("expected verify!")}
}

//----------------------------------------------------------------------------------------

func ExpectErrString(t *testing.T,err error,text string) {
	if err.Error() != text {
		t.Error("expected '"+text+"' got",err)
	}
}

func ExpectNoErr(t *testing.T,err error) {
	if err != nil {
		t.Error("expected no err, got",err)
	}
}

func mkTestDir() string {
	t := time.Now()
	d := "/tmp/holochain_test"+strconv.FormatInt(t.Unix(),10)+"."+strconv.Itoa(t.Nanosecond())
	return d
}
