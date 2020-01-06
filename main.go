package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Dependency struct {
	Path    string
	Version string
}

type License struct {
	Module  string
	License string
}

func main() {
	deps, err := ParseGoMod()
	if err != nil {
		log.Fatalf("couldn't get list of dependencies: %s\n", err)
	}

	licenses := make(map[string][]string)
	for _, license := range GetLicenses(deps) {
		modules := licenses[license.License]
		modules = append(modules, license.Module)
		licenses[license.License] = modules
	}

	// Pretty print

	licensesList := make([]string, 0, len(licenses))
	for l := range licenses {
		licensesList = append(licensesList, l)
	}
	sort.Strings(licensesList)

	fmt.Println("List of licenses:")
	for _, license := range licensesList {
		fmt.Printf("%s:\n", license)
		for _, m := range licenses[license] {
			fmt.Printf("  - %s\n", m)
		}
	}
}

// ParseGoMod calls `go list -m all` command and returns list of dependencies
func ParseGoMod() (deps []Dependency, err error) {
	cmd := exec.Command("go", "list", "-m", "all")

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("couldn't exec 'go list' command: %s", err)
	}

	sc := bufio.NewScanner(bytes.NewBuffer(out))

	// Ignore the first line because it is the path of current project
	sc.Scan()
	for sc.Scan() {
		text := sc.Text()
		line := strings.Split(text, " ")
		if len(line) != 2 {
			if len(line) == 5 {
				// Special case for 'replace' statement
				deps = append(deps, Dependency{Path: line[3], Version: line[4]})
				continue
			}
			return nil, fmt.Errorf("invalid output of 'go list' command: '%s'", text)
		}
		deps = append(deps, Dependency{Path: line[0], Version: line[1]})
	}

	return deps, nil
}

// GetLicenses makes parallel requests to https://pkg.go.dev service and parses module license tab
func GetLicenses(deps []Dependency) []License {
	depsChan := make(chan Dependency)
	go func() {
		for _, dep := range deps {
			depsChan <- dep
		}
		close(depsChan)
	}()

	licenseChan := make(chan License)

	httpClient := &http.Client{
		Timeout: time.Second * 5,
	}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for dep := range depsChan {
				module := dep.Path + "@" + dep.Version
				license, err := getLicense(httpClient, dep)
				if err != nil {
					log.Printf("couldn't get license for '%s' module: %s\n", module, err)
					license = "! unknown"
				}

				licenseChan <- License{
					Module:  module,
					License: license,
				}
			}
		}()
	}

	licenses := make([]License, 0, len(deps))
	done := make(chan struct{})
	go func() {
		for l := range licenseChan {
			licenses = append(licenses, l)
		}
		close(done)
	}()

	wg.Wait()
	close(licenseChan)
	<-done

	return licenses
}

const baseUrl = "https://pkg.go.dev/%s@%s?tab=licenses"

// licenseMatchRegexp is used to match license name. The regexp is case insensetive
var licenseMatchRegexp = regexp.MustCompile(`(?i)<div id="#license(?:\.md|\.txt|)">(.*?)</div>`)

// getLicense makes request to https://pkg.go.dev service and matches license with regexp
func getLicense(client *http.Client, dep Dependency) (string, error) {
	url := fmt.Sprintf(baseUrl, dep.Path, dep.Version)
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("couldn't get license: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	indexes := licenseMatchRegexp.FindSubmatchIndex(body)
	if indexes == nil {
		return "", fmt.Errorf("couldn't find any license")
	}

	return string(body[indexes[2]:indexes[3]]), nil
}
