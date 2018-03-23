package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

const changelogFileName = "CHANGELOG.md"

type Params struct {
	PathToSwaggerFile string
	SwaggerFileName   string
	RepoURL           string
}

type diff struct {
	hash    string
	author  string
	date    string
	changes string
}

func GenerateMDFile(p Params) error {
	result, err := exec.Command("git", "log", "--follow", pathToFile(p.PathToSwaggerFile, p.SwaggerFileName)).Output()
	if err != nil {
		return err
	}
	commitReg, err := regexp.Compile(`commit (\S*)\nAuthor: (.*) (<.*)\nDate:   (.*)`)
	if err != nil {
		return err
	}
	commitsArr := commitReg.FindAllStringSubmatch(string(result), -1)
	commits := make([]diff, len(commitsArr))
	t := 0
	for _, c := range commitsArr {
		commits[t].hash = c[1]
		commits[t].author = c[2]
		commits[t].date = c[4]
		t++
	}
	onlyDiffReg, err := regexp.Compile(`\+\+\+ (.*)\n([\s\S]*)`)
	if err != nil {
		return err
	}
	for i := 0; i < len(commits)-1; i++ {
		result, err = exec.Command("git", "diff",
			commits[i+1].hash, commits[i].hash,
			"--", pathToFile(p.PathToSwaggerFile, p.SwaggerFileName)).Output()
		if err != nil {
			return err
		}
		resultString := string(result)
		changes := onlyDiffReg.FindAllStringSubmatch(resultString, -1)
		if len(changes) > 0 {
			r := changes[0][2]
			if len(r) != 0 {
				commits[i].changes = r
			}
		}
	}

	file, err := os.Create(p.PathToSwaggerFile + changelogFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// For github url get hash of file path
	hasher := md5.New()
	hasher.Write([]byte(pathToFile(p.PathToSwaggerFile, p.SwaggerFileName)))
	githubPath := hex.EncodeToString(hasher.Sum(nil))

	// Write result information
	fmt.Fprint(file, "# History of changes in swagger file \n\n")
	for _, c := range commits {
		if len(c.changes) == 0 {
			continue
		}
		fmt.Fprintf(file, "## **Date:** %s \n", c.date)
		fmt.Fprintf(file, "**Author:** %s \n", c.author)
		fmt.Fprintf(file, "\n[Changes](%s):\n```\n%s```\n", urlToGithubChanges(p.RepoURL, c.hash, githubPath), c.changes)
	}
	return nil
}

func pathToFile(path, filename string) string {
	return path + filename
}

func urlToGithubChanges(repoURL, hash, githubPath string) string {
	return repoURL + "commit/" + hash + "#diff-" + githubPath
}
