package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/gorilla/mux"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	polyglot = ""
)

func CallHandler(w http.ResponseWriter, r *http.Request) {
	service := mux.Vars(r)["service"]
	method := mux.Vars(r)["method"]
	methodFlag := "--full_method=" + service + "/" + method
	args := []string{}
	args = append(args, methodFlag)

	endpointFlag := "--endpoint=" + r.Host
	endpointHeader := r.Header.Get("x-polyman-endpoint")
	if endpointHeader != "" {
		endpointFlag = "--endpoint=" + endpointHeader
	}
	args = append(args, endpointFlag)

	root := r.Header.Get("x-polyman-root")
	if root != "" {
		root, _ = homedir.Expand(root)
		rootFlag := "--proto_discovery_root=" + root
		args = append(args, rootFlag)
	}

	config := r.Header.Get("x-polyman-config")
	if config != "" {
		config, _ = homedir.Expand(config)
		configFlag := "--config_set_path=" + config
		args = append(args, configFlag)
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
	c1 := exec.Command("java", args...)
	var b bytes.Buffer
	c1.Stdout = &b
	c1.Stderr = os.Stderr
	c1.Start()
	c1.Wait()
	return b.Bytes()
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	args := []string{}
	root := r.Header.Get("x-polyman-root")
	if root != "" {
		root, _ = homedir.Expand(root)
		rootFlag := "--proto_discovery_root=" + root
		args = append(args, rootFlag)
	}
	config := r.Header.Get("x-polyman-config")
	if config != "" {
		config, _ = homedir.Expand(config)
		configFlag := "--config_set_path=" + config
		args = append(args, configFlag)
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

	res := List(args)
	w.Write(res)
}

func main() {
	var polyglotOverride = flag.String("polygot", "", "Location of Polyglot.jar")
	var port = flag.String("port", "8000", "Port to run Polyman on")

	flag.Parse()
	polyglot = *polyglotOverride
	if polyglot == "" {
		currentUser, _ := user.Current()
		polyglotDir := filepath.Join(currentUser.HomeDir, ".polyglot")
		if err := os.MkdirAll(polyglotDir, 0775); err != nil {
			fmt.Println("failed creating dir")
		}
		polyglot = filepath.Join(currentUser.HomeDir, ".polyglot", "polyglot.jar")
		if _, err := os.Stat(polyglot); os.IsNotExist(err) {
			fmt.Println("Downloading polyglot 1.6.0")
			resp, err := http.Get("https://github.com/grpc-ecosystem/polyglot/releases/download/v1.6.0/polyglot.jar")
			if err != nil {
				fmt.Println("http GET failed " + err.Error())
				return
			}

			jarData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("ioutil.ReadAll failed " + err.Error())
				return
			}

			if err := ioutil.WriteFile(polyglot, jarData, 0644); err != nil {
				fmt.Println("writing output failed " + err.Error())
				return
			}

			resp.Body.Close()
			fmt.Println("Download finished")
		}
	}

	fmt.Println("Starting Polyman Proxy - localhost:" + *port)
	r := mux.NewRouter()
	r.HandleFunc("/{service}/{method}", CallHandler).Methods("POST")
	r.HandleFunc("/list_services", ListHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+*port, r))
}
