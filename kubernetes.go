package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func namespaceExists(namespace string) bool {
	cmd := exec.Command("kubectl", "get", "namespace", namespace)
	err := cmd.Run()
	if err == nil {
		return true
	}

	return false
}

func kubeCreateNamespace(namespace string) error {
	cmd := exec.Command("kubectl", "create", "namespace", namespace)
	cmd.Env = os.Environ()
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	log.Println(strings.Trim(out.String(), "\r\n"))
	if err != nil {
		exitErr := err.(*exec.ExitError)
		log.Printf("Command 'kubectl create namespace %s' returned with non-zero code: %s\n", namespace, exitErr.String())
		log.Println(strings.Trim(exitErr.String(), "\r\n"))

		return err
	}

	return nil
}

func kubeApply(kubefile, tag string, env map[string]string) error {
	configBytes, err := ioutil.ReadFile(kubefile)
	if err != nil {
		return err
	}
	tmpl, err := template.New("config").Parse(string(configBytes))
	out, err := ioutil.TempFile(config.BaseDir, tag)
	env["TAG"] = tag
	err = tmpl.Execute(out, env)
	if err != nil {
		return err
	}
	out.Close()
	cmd := exec.Command("kubectl", "-n", config.Namespace, "apply", "-f", out.Name())
	cmd.Env = os.Environ()
	var cmdOut bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &cmdOut
	cmd.Stderr = &errBuf
	err = cmd.Run()
	log.Println(strings.Trim(cmdOut.String(), "\r\n"))
	if err != nil {
		exitErr := err.(*exec.ExitError)
		log.Printf("Command 'kubectl -n %s apply -f %s' returned with non-zero code: %s\n", namespace, out.Name(), exitErr.String())
		return err
	}

	return nil
}
