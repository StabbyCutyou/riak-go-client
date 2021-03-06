// +build security

package riak

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"
)

func buildClusterAndRunTest(t *testing.T, nodeOptions *NodeOptions) {
	var err error
	var node *Node
	if node, err = NewNode(nodeOptions); err != nil {
		t.Error(err.Error())
	}
	if node == nil {
		t.FailNow()
	}

	nodes := []*Node{node}
	opts := &ClusterOptions{
		Nodes: nodes,
	}

	if expected, actual := 1, len(opts.Nodes); expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
	if expected, actual := node, opts.Nodes[0]; expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}

	cluster, err := NewCluster(opts)
	if err != nil {
		t.Error(err.Error())
	}

	defer func() {
		if err := cluster.Stop(); err != nil {
			t.Error(err.Error())
		}
	}()

	if expected, actual := node, cluster.nodes[0]; expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}

	if err := cluster.Start(); err != nil {
		t.Error(err.Error())
	}

	command := &PingCommand{}
	if err := cluster.Execute(command); err != nil {
		t.Error(err.Error())
	}

	if expected, actual := true, command.Successful(); expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestExecuteCommandOnClusterWithSecurity(t *testing.T) {
	var err error
	var pemData []byte
	if pemData, err = ioutil.ReadFile("./tools/test-ca/certs/cacert.pem"); err != nil {
		t.Fatal(err.Error())
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(pemData); !ok {
		t.Fatal("could not append PEM cert data")
	}
	tlsConfig := &tls.Config{
		ServerName:         "riak-test",
		InsecureSkipVerify: false, // set to 'true' to not require CA certs
		ClientCAs:          caCertPool,
	}
	authOptions := &AuthOptions{
		User:      "riakpass",
		Password:  "Test1234",
		TlsConfig: tlsConfig,
	}
	nodeOptions := &NodeOptions{
		RemoteAddress: remoteAddress,
		AuthOptions:   authOptions,
	}
	buildClusterAndRunTest(t, nodeOptions)
}

func TestExecuteCommandOnClusterWithSecurityAndClientCertificate(t *testing.T) {
	var err error
	var caCertPemData []byte
	if caCertPemData, err = ioutil.ReadFile("./tools/test-ca/certs/cacert.pem"); err != nil {
		t.Fatal(err.Error())
	}
	var cert tls.Certificate
	if cert, err = tls.LoadX509KeyPair(
		"./tools/test-ca/certs/riakuser-client-cert.pem",
		"./tools/test-ca/private/riakuser-client-cert-key.pem"); err != nil {
		t.Fatal(err.Error())
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertPemData); !ok {
		t.Fatal("could not append PEM cert data")
	}
	tlsConfig := &tls.Config{
		ServerName:         "riak-test",
		InsecureSkipVerify: false, // set to 'true' to not require CA certs
		ClientCAs:          caCertPool,
		Certificates:       []tls.Certificate{cert},
	}
	authOptions := &AuthOptions{
		User:      "riakuser",
		TlsConfig: tlsConfig,
	}
	nodeOptions := &NodeOptions{
		RemoteAddress: remoteAddress,
		AuthOptions:   authOptions,
	}
	buildClusterAndRunTest(t, nodeOptions)
}
