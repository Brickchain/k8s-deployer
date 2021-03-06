package main

import (
	"bytes"
	"fmt"
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
		log.Printf("Command 'kubectl create namespace %s' returned with non-zero code: %s\n", namespace, err.Error())
		log.Println(strings.Trim(errBuf.String(), "\r\n"))

		return err
	}

	return nil
}

func kubeApply(kubefile, tag string, env map[string]string) error {
	configBytes, err := ioutil.ReadFile(kubefile)
	if err != nil {
		return fmt.Errorf("Failed to read file %s: %s", kubefile, err)
	}

	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
		"Title":   strings.Title,
		"TrimPrefix": func(t, s string) string {
			return strings.TrimPrefix(s, t)
		},
		"TrimSuffix": func(t, s string) string {
			return strings.TrimSuffix(s, t)
		},
		"Replace": func(f, t, s string) string {
			return strings.Replace(s, f, t, 1000)
		},
	}

	tmpl, err := template.New("config").Funcs(funcMap).Parse(string(configBytes))
	if err != nil {
		return fmt.Errorf("Failed to create template: %s", err)
	}
	out, err := ioutil.TempFile(config.BaseDir, tag)
	if err != nil {
		return fmt.Errorf("Failed to create temp file: %s", err)
	}
	env["TAG"] = tag
	err = tmpl.Execute(out, env)
	if err != nil {
		return fmt.Errorf("Failed to render template: %s", err)
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
