package main

import (
	"errors"
	giturls "github.com/whilp/git-urls"
	"net/url"
	"os"
	"os/user"
	"regexp"
	"strings"
)

type ConfigInput struct {
	src     string
	dst     string
	target  string
	cdir    string
	key     string
	proto   string
	quiet   bool
	verbose bool
}

type Config struct {
	src     []string
	dst     []string
	url     url.URL
	cdir    string
	key     string
	proto   string
	quiet   bool
	verbose bool
}

func NewConfig(ci ConfigInput) (Config, error) {
	// validate src & dst must not be empty
	if ci.src == "" {
		return Config{}, errors.New("config: src must not be empty")
	}
	if ci.dst == "" {
		return Config{}, errors.New("config: dst must not be empty")
	}

	// parse src & dst strings into slices
	src := strings.Split(ci.src, ",")
	dst := strings.Split(ci.dst, ",")

	// validate src length must match dst length if dst length is greater than 1
	srcLen := len(src)
	dstLen := len(dst)
	if srcLen > 1 && srcLen != dstLen && dstLen != 1 {
		return Config{}, errors.New("config: dst count must be 1, or match src count")
	}
	for _, x := range dst {
		if _, err := os.Stat(x); os.IsNotExist(err) {
			return Config{}, errors.New("config: destinations must be valid paths")
		}
	}

	// validate clone directory
	if ci.cdir == "" {
		return Config{}, errors.New("config: cdir value must not be empty")
	}
	if _, err := os.Stat(ci.cdir); os.IsNotExist(err) {
		return Config{}, errors.New("config: cdir must be a valid path")
	}

	// validate proto is auto, https, or ssh
	if ci.proto == "" {
		return Config{}, errors.New("config: proto must not be empty")
	} else if ci.proto != "auto" && ci.proto != "https" && ci.proto != "ssh" {
		return Config{}, errors.New("config: invalid proto")
	}

	// validate url directory
	if ci.target == "" {
		return Config{}, errors.New("config: url must not be empty")
	}
	// convert short target into url
	gitUrl, err := giturls.Parse(ci.target)
	if err != nil {
		return Config{}, err
	}

	// convert short format to actual url
	filePath := regexp.MustCompile(`^file://@[a-z0-9-]{0,38}/`)
	if filePath.MatchString(gitUrl.String()) {
		configURL := ""
		if ci.proto == "ssh" || ci.proto == "auto" {
			// TODO: Only set auto to ssh if the repo is private
			ci.proto = "ssh"
			configURL = "git@github.com:" + ci.target
		} else {
			configURL = "https://github.com/" + ci.target + ".git"
		}
		gitUrl, _ = giturls.Parse(configURL)
	}

	// validate ssh key exists if proto is ssh and repo is private
	if ci.proto == "ssh" && ci.key == "" {
		usr, _ := user.Current()
		ci.key = usr.HomeDir + "/.ssh/id_rsa"
	}

	return Config{
		url:   *gitUrl,
		src:   src,
		dst:   dst,
		cdir:  ci.cdir,
		key:   ci.key,
		proto: ci.proto,
	}, nil
}
