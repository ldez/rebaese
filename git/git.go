package git

import (
	"fmt"
	"log"
	"os/exec"
)

func CloneWithLimit(repositoryURL string, branchName string, depth string) (string, error) {
	log.Printf("git clone %s -b %s --depth %s --single-branch .", repositoryURL, branchName, depth)

	return command("git", "clone", repositoryURL, "-b", branchName, "--depth", depth, "--single-branch", ".")
}

func CloneBranch(repositoryURL string, branchName string) (string, error) {
	log.Printf("git clone %s -b %s .", repositoryURL, branchName)

	return command("git", "clone", repositoryURL, "-b", branchName, ".")
}

func Clone(repositoryURL string) (string, error) {
	log.Printf("git clone %s .", repositoryURL)

	return command("git", "clone", repositoryURL, ".")
}

func AddRemote(remoteName string, repositoryURL string) (string, error) {
	log.Printf("git remote add %s %s", remoteName, repositoryURL)

	return command("git", "remote", "add", remoteName, repositoryURL)
}

func FetchWithLimit(remoteName string, branchName string, depth string) (string, error) {
	log.Printf("git fetch --depth %s --no-tags %s %s", depth, remoteName, branchName)

	return command("git", "fetch", "--depth", depth, "--no-tags", remoteName, branchName)
}

func Fetch(remoteName string, branchName string) (string, error) {
	log.Printf("git fetch --no-tags %s %s", remoteName, branchName)

	return command("git", "fetch", "--no-tags", remoteName, branchName)
}

func Rebase(remoteName string, branchName string) (string, error) {
	log.Printf("git rebase --preserve-merges %s/%s", remoteName, branchName)

	return command("git", "rebase", "--preserve-merges", fmt.Sprintf("%s/%s", remoteName, branchName))
}

func Checkout(branchName string) (string, error) {
	log.Printf("git checkout %s", branchName)

	return command("git", "checkout", branchName)
}

func PushForce(remoteName string, branchName string) (string, error) {
	log.Printf("git push -f --force-with-lease %s %s", remoteName, branchName)

	return command("git", "push", "-f", "--force-with-lease", remoteName, branchName)
}

func Config(configKey string, configValue string) (string, error) {
	log.Printf("git config %s %s", configKey, configValue)

	return command("git", "config", configKey, configValue)
}

func command(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()

	return string(output), err
}
