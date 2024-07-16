package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

type repo struct {
	Sync               bool                `yaml:"sync"`
	Name               string              `yaml:"name"`
	RepoLink           string              `yaml:"repo_link"`
	Branch             string              `yaml:"branch"`
	RemoteAPIDirectory string              `yaml:"remote_api_directory"`
	LocalAPIDirectory  string              `yaml:"local_api_directory"`
	ReplaceImports     []map[string]string `yaml:"replace_imports"`
}

func main() {
	// Enable glog output
	_ = flag.Lookup("logtostderr").Value.Set("true")
	_ = flag.Lookup("v").Value.Set("100")
	configFiles := flag.String("config-file", "internal/sync/configs", "path to config files")
	flag.Parse()

	glog.V(100).Infof("Loading config file")

	config := newConfig(*configFiles)

	glog.V(100).Infof("Initiating repository sync")

	for _, repo := range config {
		if repo.Sync {
			glog.V(100).Infof("#### Syncing repo %s ####", repo.Name)
			syncRemoteRepo(&repo)
		} else {
			glog.V(100).Infof("Sync disabled for repo %s. Skip", repo.Name)
		}
	}
}

func syncRemoteRepo(repo *repo) {
	glog.V(100).Infof("Syncing repo: %s, destination repo link: %s", repo.Name, repo.RemoteAPIDirectory)

	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	projectClonedDirectory := path.Join(basePath, repo.Name, repo.RemoteAPIDirectory)
	projectLocalDirectory := path.Join("./pkg", repo.LocalAPIDirectory)

	gitClone(basePath, repo)

	glog.V(100).Infof("Comparing local %s and cloned %s api directories for repo %s",
		projectLocalDirectory, projectClonedDirectory, repo.Name)

	if repoSynced(projectClonedDirectory, projectLocalDirectory, repo) {
		gitReset(repo.LocalAPIDirectory)
	} else {
		syncDirectories(projectClonedDirectory, projectLocalDirectory, repo)
	}

	glog.V(100).Infof("Remove cloned directory from filesystem: %s", path.Join(basePath, repo.Name))

	if _, err := os.Stat(path.Join(basePath, repo.Name)); !os.IsNotExist(err) {
		err := os.RemoveAll(path.Join(basePath, repo.Name))

		if err != nil {
			glog.V(100).Infof("Failed to remove cloned directory %s exit with error code 1",
				path.Join(basePath, repo.Name))
			os.Exit(1)
		}
	}
}

func repoSynced(clonedDir, localDir string, repo *repo) bool {
	glog.V(100).Infof("Verifying destination directory %s exists", localDir)

	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		glog.V(100).Infof("Destination api directory %s doesn't exist creating directory", localDir)

		if os.MkdirAll(localDir, 0777) != nil {
			glog.V(100).Infof("Failed to create api directory. Exit with error code 1")
			os.Exit(1)
		}
	}

	glog.V(100).Infof("Replace cloned package name: %s with the local package name: %s",
		path.Base(clonedDir), path.Base(localDir))

	err := refactor(
		fmt.Sprintf("package %s", path.Base(localDir)),
		fmt.Sprintf("package %s", path.Base(clonedDir)),
		fmt.Sprintf("./%s", localDir), "*.go")

	if err != nil {
		glog.V(100).Infof("Failed to refactor file before sync due to %w. Exit with error 1", err)
		os.Exit(1)
	}

	glog.V(100).Infof("Replace cloned package imports")

	for _, importMap := range repo.ReplaceImports {
		err = refactor(
			importMap["new"],
			importMap["old"],
			localDir,
			"*.go")

		if err != nil {
			glog.V(100).Infof("Failed to refactor file. Exit with error 1")
			os.Exit(1)
		}
	}

	err = execCmd("", "diff", []string{clonedDir, localDir})

	if err == nil {
		glog.V(100).Infof("Repo synced. Revert local files to original state")

		return true
	}

	return false
}

func syncDirectories(clonedDir, localDir string, repo *repo) {
	glog.V(100).Infof("Repos are not synced. Cleaning local directory: %s", localDir)

	if os.RemoveAll(localDir) != nil {
		glog.V(100).Infof("Failed to remove local api directory. Exit with error code 1")
		os.Exit(1)
	}

	glog.V(100).Infof("Create new local directory: %s", localDir)

	if os.MkdirAll(localDir, 0750) != nil {
		glog.V(100).Infof("Failed to recreate api directory. Exit with error code 1")
		os.Exit(1)
	}

	glog.V(100).Infof("Copy api filed from cloned directory to local api directory")

	err := execCmd(
		"",
		"cp",
		[]string{"-a", fmt.Sprintf("%s/.", clonedDir), fmt.Sprintf("%s/", localDir)})

	if err != nil {
		glog.Infof("Failed to sync directories. Exit with error code 1")
		os.Exit(1)
	}

	glog.V(100).Infof("Fix packages names")

	err = refactor(
		fmt.Sprintf("package %s", path.Base(clonedDir)),
		fmt.Sprintf("package %s", path.Base(localDir)),
		localDir,
		"*.go")

	if err != nil {
		glog.V(100).Infof("Failed to refactor file. Exit with error 1")
		os.Exit(1)
	}

	for _, importMap := range repo.ReplaceImports {
		err = refactor(importMap["old"], importMap["new"], localDir, "*.go")

		if err != nil {
			glog.V(100).Infof("Failed to refactor file. Exit with error 1")
			os.Exit(1)
		}
	}
}

func gitReset(packageName string) {
	for _, cmdToRun := range [][]string{
		{"reset", "--", fmt.Sprintf("./pkg/%s", packageName)},
		{"checkout", "--", fmt.Sprintf("./pkg/%s", packageName)},
		{"clean", "-d", "-f", fmt.Sprintf("./pkg/%s", packageName)},
	} {
		err := execCmd("", "git", cmdToRun)

		if err != nil {
			glog.Infof("Failed to reset project to it's original state. Exit with error 1")
			os.Exit(1)
		}
	}
}

func gitClone(localPath string, repo *repo) {
	glog.V(100).Infof("Cloning repo %s from %s", repo.Name, repo.RepoLink)
	localDirectory := path.Join(localPath, repo.Name)

	if _, err := os.Stat(localDirectory); !os.IsNotExist(err) {
		glog.V(100).Infof(
			"Local directory already exists for the project: %s. Removing directory", repo.Name)

		err := os.RemoveAll(localDirectory)

		if err != nil {
			glog.V(100).Infof("Failed to remove repo directory %s due to %w exit with error code 1",
				localDirectory, err)
			os.Exit(1)
		}
	}

	err := execCmd(
		localPath,
		"git",
		[]string{"clone", "-n", "--depth=1", "--filter=tree:0", "-b", repo.Branch, repo.RepoLink, repo.Name})

	if err != nil {
		glog.V(100).Infof("Failed to clone repo due to cmd error. Exit with error code 1")
		os.Exit(1)
	}

	err = execCmd(localDirectory, "git", []string{"sparse-checkout", "set", "--no-cone", repo.RemoteAPIDirectory})

	if err != nil {
		glog.V(100).Infof("Failed to sparse-checkout repo due to cmd error. Exit with error code 1")
		os.Exit(1)
	}

	err = execCmd(localDirectory, "git", []string{"checkout"})

	if err != nil {
		glog.V(100).Infof("Failed to checkout repo due to cmd error. Exit with error code 1")
		os.Exit(1)
	}
}

func execCmd(dirName, binary string, args []string) error {
	glog.V(100).Infof("Executing cmd: %s, with args: %v, in directory: %s", binary, args, dirName)

	cmd := exec.Command(binary, args...)

	if dirName != "" {
		cmd.Dir = dirName
	}

	out, err := cmd.Output()

	if err != nil {
		glog.V(100).Infof("Failed to execute cmd due to %s. Output: %s", err, string(out))

		return err
	}

	return nil
}

func newConfig(pathToConfigFiles string) []repo {
	var repoConfigs []repo

	glog.V(100).Infof("Read files in configs directory")

	err := filepath.Walk(pathToConfigFiles, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			glog.V(100).Infof("Loading config file %s", info.Name())

			var config []repo
			err := readFile(&config, path)

			if err != nil {
				glog.V(100).Infof("Error to read config file: %w", err)

				return err
			}

			repoConfigs = append(repoConfigs, config...)
		}

		return nil
	})

	if err != nil {
		glog.V(100).Infof("Error to list files in directory %s due to %w", pathToConfigFiles, err)

		return nil
	}

	if len(repoConfigs) == 0 {
		glog.V(100).Infof("Config directory is empty")

		return nil
	}

	return repoConfigs
}

func readFile(cfg *[]repo, cfgFile string) error {
	openedCfgFile, err := os.Open(cfgFile)
	if err != nil {
		return err
	}

	defer openedCfgFile.Close()

	decoder := yaml.NewDecoder(openedCfgFile)
	err = decoder.Decode(&cfg)

	if err != nil {
		return err
	}

	return nil
}

func refactor(oldLine, newLine, root string, patterns ...string) error {
	return filepath.Walk(root, refactorFunc(oldLine, newLine, patterns))
}

func refactorFunc(oldLine, newLine string, filePatterns []string) filepath.WalkFunc {
	return func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		var matched bool

		for _, pattern := range filePatterns {
			var err error
			matched, err = filepath.Match(pattern, fileInfo.Name())

			if err != nil {
				return err
			}

			if matched {
				oldContents, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}

				glog.V(100).Infof("Refactoring: %s", filePath)

				newContents := strings.ReplaceAll(string(oldContents), oldLine, newLine)

				err = os.WriteFile(filePath, []byte(newContents), 0)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
}
