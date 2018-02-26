package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	polyglot = ""
)

func CallHandler(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	method := mux.Vars(r)["method"]
	method_flag := "--full_method=" + service + "/" + method
	args := []string{}
	args = append(args, method_flag)

	endpoint_flag := "--endpoint=" + r.Host
	endpointHeader := r.Header.Get("x-polyman-endpoint")
	if endpointHeader != "" {
		endpoint_flag = "--endpoint=" + endpointHeader
	}
	args = append(args, endpoint_flag)

	root := r.Header.Get("x-polyman-root")
	if root != "" {
		root, _ = homedir.Expand(root)
		root_flag := "--proto_discovery_root=" + root
		args = append(args, root_flag)
	}

	config := r.Header.Get("x-polyman-config")
	if config != "" {
		config, _ = homedir.Expand(config)
		config_flag := "--config_set_path=" + config
		args = append(args, config_flag)
	}

	body, _ := ioutil.ReadAll(r.Body)
	input := string(body[:])

	res, err := Call(input, args)
	if err != nil {
		http.Error(w, string(res[:]), http.StatusInternalServerError)
	}

	w.Write(res)
}

func Call(body string, opts []string) ([]byte, error) {
	c1 := exec.Command("echo", body)
	args := []string{"-jar", polyglot, "--command=call"}
	args = append(args, opts...)
	c2 := exec.Command("java", args...)
	c2.Stdin, _ = c1.StdoutPipe()
	var b bytes.Buffer
	var e bytes.Buffer
	c2.Stdout = &b
	c2.Stderr = &e
	c1.Start()
	c2.Start()
	c1.Wait()
	err := c2.Wait()
	if err != nil {
		return e.Bytes(), err
	}
	return b.Bytes(), nil
}

func List(opts []string) []byte {
	args := []string{"-jar", polyglot, "--command=list_services", "--with_message=true"}
	args = append(args, opts...)
	fmt.Printf("%v\n", args)
	c1 := exec.Command("java", args...)
	var b bytes.Buffer
	c1.Stdout = &b
	c1.Stderr = os.Stderr
	c1.Start()
	c1.Wait()
	return b.Bytes()
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	args := []string{}
	root := r.Header.Get("x-polyman-root")
	if root != "" {
		root, _ = homedir.Expand(root)
		root_flag := "--proto_discovery_root=" + root
		args = append(args, root_flag)
	}
	config := r.Header.Get("x-polyman-config")
	if config != "" {
		config, _ = homedir.Expand(config)
		config_flag := "--config_set_path=" + config
		args = append(args, config_flag)
	}
	methodFilter := v.Get("method")
	if methodFilter != "" {
		methodFlag := "--method_filter=" + methodFilter
		args = append(args, methodFlag)
	}
	serviceFilter := v.Get("service")
	if serviceFilter != "" {
		serviceFlag := "--service_filter=" + serviceFilter
		args = append(args, serviceFlag)
	}

	fmt.Printf("%v\n", args)
	res := List(args)
	w.Write(res)
}

func main() {
	polyglot = "~/polyglot.jar"
	polyglot, _ = homedir.Expand(polyglot)

	r := mux.NewRouter()
	r.HandleFunc("/{service}/{method}", CallHandler).Methods("POST")
	r.HandleFunc("/list_services", ListHandler).Methods("GET")
	r.HandleFunc("/", HomeHandler)
	log.Fatal(http.ListenAndServe(":8000", r))
}
