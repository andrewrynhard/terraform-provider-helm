package repo

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

type Repo struct{}

// Options are the options for locating and retrieving a chart.
type Options struct {
	Host      string
	Name      string
	Namespace string
	Chart     string
	Version   string
}

func (r *Repo) FindChart(o *Options) (string, error) {
	settings := environment.EnvSettings{
		Home:            helmpath.Home(environment.DefaultHelmHome),
		TillerHost:      o.Host,
		TillerNamespace: o.Namespace,
	}

	chartDownloader := downloader.ChartDownloader{
		HelmHome: settings.Home,
		Out:      os.Stdout,
		Getters:  getter.All(settings),
		Verify:   downloader.VerifyIfPossible,
	}

	var ref string
	if isValidURL(o.Name) {
		absoluteChartURL, err := repo.FindChartInRepoURL(o.Name, o.Chart, o.Version, "", "", "", getter.All(settings))
		if err != nil {
			return "", err
		}
		filename, _, err := chartDownloader.DownloadTo(absoluteChartURL, o.Version, settings.Home.Archive())
		if err != nil {
			return "", err
		}
		absolutePath, err := filepath.Abs(filename)
		if err != nil {
			return "", err
		}

		ref = absolutePath
	}

	if o.Name == "stable" {
		absoluteChartURL := o.Name + "/" + o.Chart
		filename, _, err := chartDownloader.DownloadTo(absoluteChartURL, o.Version, settings.Home.Archive())
		if err != nil {
			return "", err
		}

		absolutePath, err := filepath.Abs(filename)
		if err != nil {
			return "", err
		}

		ref = absolutePath
	}

	if ref == "" {
		fileInfo, err := os.Stat(o.Chart)
		if err != nil {
			return "", fmt.Errorf("Could not find local chart: %s", err.Error())
		}
		if fileInfo.IsDir() {
			absolutePath, err := filepath.Abs(o.Chart)
			if err != nil {
				return "", err
			}

			ref = absolutePath
		}
	}

	return ref, nil
}

func isValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	return true
}
