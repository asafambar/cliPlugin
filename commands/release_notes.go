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
}

type ReleaseNotesResponse struct {
	Version string `json:"version"`
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
			Name:        "product",
			Description: "The name of Jfrog product to get release notes",
		},
		{
			Name:        "version",
			Description: "The version of Jfrog product to get release notes (only if -current flag not exist)",
		},
	}
}

func getReleaseNotesFlags() []components.Flag {
	return []components.Flag{
		components.BoolFlag{
			Name:         "current",
			Description:  "Get release notes for the default current product version - only <product> argument should be provided",
			DefaultValue: false,
		},
		components.BoolFlag{
			Name:         "date",
			Description:  "Get only the date of jfrog product and version release",
			DefaultValue: false,
		},
		components.StringFlag{
			Name:        "version",
			Description: "version of the product",
			Mandatory:   false,
		},
	}
}

type releaseNotesConfiguration struct {
	product string
	version string
	current bool
	date    bool
}

func releaseNotesCmd(c *components.Context) error {
	conf, err := extractAllArgsAndFlags(c)
	if err != nil {
		return err
	}
	releaseNotesString, err := doGetReleaseNotes(c, conf)
	if err != nil {
		return err
	}
	log.Output(releaseNotesString)
	return nil
}

func extractAllArgsAndFlags(c *components.Context) (*releaseNotesConfiguration, error) {
	var conf = &releaseNotesConfiguration{}
	// get all flags
	if c.GetBoolFlagValue("current") {
		if len(c.Arguments) != 1 {
			return nil, errors.New("Wrong number of arguments. -current flag Expected: 1 argument: 'product' " + "Received: " + strconv.Itoa(len(c.Arguments)))
		}
		conf.current = true
	}
	if c.GetBoolFlagValue("date") {
		conf.date = true
	}
	if len(c.GetStringFlagValue("version")) > 0 {
		conf.version = c.GetStringFlagValue("version")
	}

	// get all arguments
	if len(c.Arguments) == 2 {
		conf.product = c.Arguments[0]
		conf.version = c.Arguments[1]
	} else if len(c.Arguments) == 1 && (len(c.GetStringFlagValue("version")) > 0 || c.GetBoolFlagValue("current")) {
		conf.product = c.Arguments[0]
	} else {
		return nil, errors.New("Wrong number of arguments. Expected: 1 or 2, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}
	return conf, nil
}

func doGetReleaseNotes(ctx *components.Context, c *releaseNotesConfiguration) (string, error) {
	productUrlTempl, ok := productEndpointTemplate[c.product]
	if !ok {
		return "", errors.New(fmt.Sprintf("Product name %s is not valid", c.product))
	}
	if c.current {
		version, err := getCurrentProductVersion(ctx, c.product)
		if err != nil {
			return "", err
		}
		c.version = version
	}
	productEndpoint := fmt.Sprintf(productUrlTempl, c.version)
	url := fmt.Sprintf("%s/%s", baseUrl, productEndpoint)
	rn, err := makeReleaseNotesRequest(url, c)
	if err != nil {
		return "", err
	}
	if c.date {
		return extractReleasedDate(rn.Bintray.Content, c.version)
	}
	return rn.Bintray.Content, nil
}

func getCurrentProductVersion(ctx *components.Context, product string) (string, error) {
	rtDefault, err := getRtDetails(ctx)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get artifactory configuration, err: %v", err))
	}
	if rtDefault == nil {
		return "", errors.New("Failed to get current default artifactory details")
	}
	rtVersion, err := getArtiVersion(rtDefault)
	if err != nil {
		return "", err
	}
	if product == "xray" {
		if version.NewVersion(rtVersion).Compare("7.0.0") < 0 {
			return getXrayVersion(rtDefault)
		}
		return "", errors.New("Cant get release notes for Xray version lower than 3.0.0")
	}
	return rtVersion, nil
}

func makeReleaseNotesRequest(url string, c *releaseNotesConfiguration) (*ReleaseNotesResponse, error) {
	client, err := httpclient.ClientBuilder().Build()
	resp, respBody, _, err := client.SendGet(url, false, httputils.HttpClientDetails{Headers: map[string]string{"Content-Type": "application/json"}})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get Release notes: %v", err))
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New(fmt.Sprintf("couldnt find release notes for %s version %s", c.product, c.version))
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
		return "", errors.New(fmt.Sprintf("Failed to get artifactory version, err: %v", err))
	}
	return version, nil
}

//support 3.x only
func getXrayVersion(rtDetails *config.ArtifactoryDetails) (string, error) {
	xrayUrl := fmt.Sprintf("%s/%s",
		strings.TrimSuffix(rtDetails.Url[:strings.LastIndex(strings.TrimSuffix(rtDetails.Url, "/"), "/")], "/"), "xray/api/v1/system/version")
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
		return "", errors.New(fmt.Sprintf("Couldnt find date for release date for version %s", version))
	}
	releaseDateStart := fullReleaseNotes[indexOfReleaseDate:]
	indexLastOfDate := strings.Index(releaseDateStart, "####")
	return releaseDateStart[:indexLastOfDate], nil
}
