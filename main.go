package main

import (
	"fmt"

	"os"

	"strings"

	"io/ioutil"

	"flag"

	"log"

	"path"

	"gopkg.in/yaml.v2"
)

var (
	config     *Config
	configFile = flag.String("config", "", "Config file")
	redisAddr  = flag.String("redis", "", "Redis")
	namespace  = flag.String("namespace", "", "Namespace")
	artifact   = flag.String("artifact", "", "Create YAML with what was deployed")
	clearState = flag.Bool("clear-state", false, "Clear the state for this namespace")
	state      State
	err        error
)

func main() {
	flag.Parse()

	if *clearState && *namespace != "" && *redisAddr != "" {
		state, err = NewRedisState(*redisAddr)
		if err != nil {
			log.Fatal(err)
		}
		err = state.Clear(*namespace)
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	if *configFile == "" {
		log.Fatal("Missing required field: config")
	}

	config, err = parseConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Set Namespace
	if config.Namespace == "<no value>" || config.Namespace == "" {
		config.Namespace = "dev"
	}
	if *namespace != "" {
		config.Namespace = *namespace
	}

	// Set DefaultBranch
	if config.DefaultBranch == "<no value>" || config.DefaultBranch == "" {
		config.DefaultBranch = "master"
	}

	// Set KubeFolder
	if config.KubeFolder == "<no value>" || config.KubeFolder == "" {
		config.KubeFolder = "k8s"
	}

	// Set BaseDir
	if config.BaseDir == "<no value>" || config.BaseDir == "" {
		config.BaseDir = "/tmp/deployer/"
	}

	// Connect to Redis if address given
	if *redisAddr != "" {
		state, err = NewRedisState(*redisAddr)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Namespace:", config.Namespace)

	// Create namespace if it doesn't already exist
	if !namespaceExists(config.Namespace) {
		err = kubeCreateNamespace(config.Namespace)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Record environment variable for Namespace
	envMap := envToMap()
	envMap["NAMESPACE"] = config.Namespace

	// Start recording values that we can later write to the "artifact" file
	outConf := Config{
		KubeFolder: config.KubeFolder,
	}

	// If we are in a repo we should record the remote uri and current commit
	if _, err := os.Stat(".git"); err == nil {
		localRemote, err := getLocalRemote(".git")
		if err != nil {
			log.Fatal(err)
		}
		localRef, err := getLocalRef(".git")
		if err != nil {
			log.Fatal(err)
		}
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		localName := path.Base(wd)
		outConf.Repositories = append(outConf.Repositories, Repository{
			Name:   localName,
			URI:    localRemote,
			Commit: localRef,
		})
	}

	// Apply k8s files in local repo
	if config.KubeFolder != "" && config.KubeFolder != "<no value>" {
		config.KubeFolder = strings.TrimSuffix(config.KubeFolder, "/")
		files, err := ioutil.ReadDir(config.KubeFolder)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			log.Println("./" + config.KubeFolder + "/" + f.Name())
			err = kubeApply(config.KubeFolder+"/"+f.Name(), "", envMap)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Read environment variables that will signal what repo to update to some commit
	updateRepo := os.Getenv(config.UpdateRepoVar)
	updateRepoRef := os.Getenv(config.UpdateRefVar)

	// Loop over repositories
	for _, repo := range config.Repositories {
		statePath := fmt.Sprintf("k8s-deployer/%s/%s", config.Namespace, repo.URI)
		var refName string
		oldRef := repo.Commit

		// If this repository is the one signaled in updateRepo we should apply that ref,
		// otherwise apply ref either from state db or from config.
		// If neither exist apply from DefaultBranch.
		if repo.Name != "" && repo.Name == updateRepo && updateRepoRef != "" {
			refName = os.Getenv(config.UpdateRefVar)
		} else {
			if oldRef == "" {
				if state != nil {
					oldRef, _ = state.Get(statePath)
				}
			}
			if oldRef != "" {
				refName = oldRef
			} else {
				refName = "refs/remotes/origin/" + config.DefaultBranch
			}
		}

		// Clone files from repository and if it has prefix of KubeFolder we execute this function.
		// If ref has changed we should apply the k8s file.
		ref, err := cloneFiles(repo.URI, refName, func(filePath, ref string) error {
			log.Println(repo.URI, ref, path.Base(filePath))
			if oldRef != ref {
				err = kubeApply(filePath, ref, envMap)
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		// Record state
		if state != nil {
			state.Set(statePath, ref)
		}
		outConf.Repositories = append(outConf.Repositories, Repository{
			Name:   repo.Name,
			URI:    repo.URI,
			Commit: ref,
		})
	}

	// If the -artifact parameter was given, write outConf to file.
	if *artifact != "" {
		outConfBytes, err := yaml.Marshal(outConf)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(*artifact, outConfBytes, 0644)
	}
}
