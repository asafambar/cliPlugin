package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	artutils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/version"
	"net/http"
	"strconv"
	"strings"
)

var baseUrl = "https://bintray.com/api/v1/packages/jfrog"
var productEndpointTemplate = map[string]string{
	"artifactory": "artifactory-pro/jfrog-artifactory-pro/versions/%s/release_notes/",
	"xray":        "jfrog-xray/jfrog-xray/versions/%s/release_notes/",
	"pipelines":   "pipelines/jfrog-pipelines/versions/%s/release_notes/",
}

type ReleaseNotesResponse struct {
	Version string `json:"Version"`
	Package string `json:"package"`
	Repo    string `json:"repo"`
	Owner   string `json:"owner"`
	Bintray struct {
		Content string `json:"content"`
		Syntax  string `json:"syntax"`
	} `json:"bintray"`
}

type XrayVersion struct {
	XrayVersion  string `json:"xray_version"`
	XrayRevision string `json:"xray_revision"`
}

func GetReleaseNotesCommands() components.Command {
	return components.Command{
		Name:        "release-notes",
		Description: "Get Release notes.",
		Aliases:     []string{"rn"},
		Arguments:   getReleasesArgument(),
		Flags:       getReleaseNotesFlags(),
		Action: func(c *components.Context) error {
			return releaseNotesCmd(c)
		},
	}
}

func getReleasesArgument() []components.Argument {
	return []components.Argument{
		{
			Name:        "Product",
			Description: "The name of Jfrog Product to get release notes",
		},
		{
			Name:        "Version",
			Description: "The Version of Jfrog Product to get release notes (only if -Current flag not exist)",
		},
	}
}

func getReleaseNotesFlags() []components.Flag {
	return []components.Flag{
		components.BoolFlag{
			Name:         "Current",
			Description:  "Get release notes for the default Current Product Version - only <Product> argument should be provided",
			DefaultValue: false,
		},
		components.BoolFlag{
			Name:         "Date",
			Description:  "Get only the Date of jfrog Product and Version release",
			DefaultValue: false,
		},
		components.StringFlag{
			Name:        "Version",
			Description: "Version of the Product",
			Mandatory:   false,
		},
	}
}

type ReleaseNotesConfiguration struct {
	Product string
	Version string
	Current bool
	Date    bool
}

func releaseNotesCmd(c *components.Context) error {
	conf, err := extractAllArgsAndFlags(c)
	if err != nil {
		return err
	}
	releaseNotesString, err := DoGetReleaseNotes(c, conf)
	if err != nil {
		return err
	}
	log.Output(releaseNotesString)
	return nil
}

func extractAllArgsAndFlags(c *components.Context) (*ReleaseNotesConfiguration, error) {
	var conf = &ReleaseNotesConfiguration{}
	// get all flags
	if c.GetBoolFlagValue("Current") {
		if len(c.Arguments) != 1 {
			return nil, errors.New("Wrong number of arguments. -Current flag Expected: 1 argument: 'Product' " + "Received: " + strconv.Itoa(len(c.Arguments)))
		}
		conf.Current = true
	}
	if c.GetBoolFlagValue("Date") {
		conf.Date = true
	}
	if len(c.GetStringFlagValue("Version")) > 0 {
		conf.Version = c.GetStringFlagValue("Version")
	}
	// get all arguments
	if len(c.Arguments) == 2 {
		conf.Product = c.Arguments[0]
		conf.Version = c.Arguments[1]
	} else if len(c.Arguments) == 1 && (len(c.GetStringFlagValue("Version")) > 0 || c.GetBoolFlagValue("Current")) {
		conf.Product = c.Arguments[0]
	} else {
		return nil, errors.New("Wrong number of arguments. Expected: 1 or 2, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}
	return conf, nil
}

func DoGetReleaseNotes(ctx *components.Context, c *ReleaseNotesConfiguration) (string, error) {
	productUrlTempl, ok := productEndpointTemplate[c.Product]
	if !ok {
		return "", errors.New(fmt.Sprintf("Product name %s is not valid", c.Product))
	}
	if c.Current {
		version, err := getCurrentProductVersion(ctx, c.Product)
		if err != nil {
			return "", err
		}
		c.Version = version
	}
	productEndpoint := fmt.Sprintf(productUrlTempl, c.Version)
	url := fmt.Sprintf("%s/%s", baseUrl, productEndpoint)
	rn, err := makeReleaseNotesRequest(url, c)
	if err != nil {
		return "", err
	}
	rnContent, err := filterTextFlags(c, rn)
	if err != nil {
		return "", err
	}
	return rnContent, nil
}

func filterTextFlags(c *ReleaseNotesConfiguration, rn *ReleaseNotesResponse) (string, error) {
	content := rn.Bintray.Content
	var err error
	if c.Date {
		content, err = extractReleasedDate(rn.Bintray.Content, c.Version)
	}
	return content, err
}

func getCurrentProductVersion(ctx *components.Context, product string) (string, error) {
	rtDefault, err := getRtDetails(ctx)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get artifactory configuration, err: %v", err))
	}
	if rtDefault == nil {
		return "", errors.New("Failed to get Current default artifactory details")
	}
	rtVersion, err := getArtiVersion(rtDefault)
	if err != nil {
		return "", err
	}
	if product == "xray" {
		if version.NewVersion(rtVersion).Compare("7.0.0") < 0 {
			return getXrayVersion(rtDefault)
		}
		return "", errors.New("Cant get release notes for Xray Version lower than 3.0.0")
	}
	return rtVersion, nil
}

func makeReleaseNotesRequest(url string, c *ReleaseNotesConfiguration) (*ReleaseNotesResponse, error) {
	client, err := httpclient.ClientBuilder().Build()
	resp, respBody, _, err := client.SendGet(url, false, httputils.HttpClientDetails{Headers: map[string]string{"Content-Type": "application/json"}})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get Release notes: %v", err))
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New(fmt.Sprintf("couldnt find release notes for %s Version %s", c.Product, c.Version))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Recieved unexpected status code from %s. status code: %v", url, resp.StatusCode))
	}
	rn := &ReleaseNotesResponse{}
	err = json.Unmarshal(respBody, rn)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Recieved unexpected response, "+
			"Failed to parse response from %s. err: %v", url, err))
	}
	return rn, nil
}

func getArtiVersion(rtDetails *config.ArtifactoryDetails) (string, error) {
	artService, err := artutils.CreateServiceManager(rtDetails, false)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get artifactory details, err: %v", err))
	}
	version, err := artService.GetConfig().GetServiceDetails().GetVersion()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get artifactory Version, err: %v", err))
	}
	return version, nil
}

//support 3.x only
func getXrayVersion(rtDetails *config.ArtifactoryDetails) (string, error) {
	xrayUrl := fmt.Sprintf("%s/%s",
		strings.TrimSuffix(rtDetails.Url[:strings.LastIndex(strings.TrimSuffix(rtDetails.Url, "/"), "/")], "/"), "xray/api/v1/system/Version")
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get Release notes for Xray: %v", err))
	}
	resp, respBody, _, err := client.SendGet(xrayUrl, false, httputils.HttpClientDetails{Headers: map[string]string{"Content-Type": "application/json"}})
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Recieved unexpected status code from %s. status code: %v", xrayUrl, resp.StatusCode))
	}
	xrayVersion := &XrayVersion{}
	json.Unmarshal(respBody, xrayVersion)
	return xrayVersion.XrayVersion, nil
}

func getRtDetails(c *components.Context) (*config.ArtifactoryDetails, error) {
	details, err := commands.GetConfig(c.GetStringFlagValue("server-id"), false)
	if err != nil {
		return nil, err
	}
	if details.Url == "" {
		return nil, errors.New("no server-id was found, or the server-id has no url")
	}
	details.Url = utils.AddTrailingSlashIfNeeded(details.Url)
	err = config.CreateInitialRefreshableTokensIfNeeded(details)
	if err != nil {
		return nil, err
	}
	return details, nil
}

func extractReleasedDate(fullReleaseNotes string, version string) (string, error) {
	indexOfReleaseDate := strings.Index(fullReleaseNotes, "Released:")
	if indexOfReleaseDate == -1 {
		return "", errors.New(fmt.Sprintf("Couldnt find Date for release Date for Version %s", version))
	}
	releaseDateStart := fullReleaseNotes[indexOfReleaseDate:]
	indexLastOfDate := strings.Index(releaseDateStart, "####")
	return strings.TrimSpace(releaseDateStart[:indexLastOfDate]), nil
}
