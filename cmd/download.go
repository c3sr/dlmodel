package cmd

import (
	"fmt"
	"github.com/c3sr/dlframework"
	"github.com/c3sr/downloadmanager"
	"github.com/spf13/cobra"
	"github.com/unknwon/com"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var downloadCmd = &cobra.Command{
	Use:   "download [path to download]",
	Short: "Download models.",
	Long:  `Download model and copy model manifests into the tempdir.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, pattern := range args {
			if err := downloadPattern(pattern); err != nil {
				fmt.Println(err.Error())
				return nil
			}
		}
		return nil
	},
}

func DownloadPattern(pattern string) error {
	return downloadPattern(pattern)
}

func downloadPattern(pattern string) error {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, path := range matches {
		err = downloadPath(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadPath(path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return downloadFile(p)
	})
}

func downloadFile(path string) error {
	if ext := filepath.Ext(path); ext != ".yml" && ext != ".yaml" {
		abspath, _ := filepath.Abs(path)
		log.WithField("path", abspath).Info("the given path is not a yaml file")
		return nil
	}

	bts, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var model dlframework.ModelManifest
	if err := yaml.Unmarshal(bts, &model); err != nil {
		return err
	}

	if model.GetHidden() {
		log.WithField("name", model.GetName()).Info("skipping registration of hidden model")
		return nil
	}

	workDir, err := model.WorkDir()
	if err != nil {
		return err
	}

	if !com.IsDir(workDir) {
		if err := os.MkdirAll(workDir, 0700); err != nil {
			return fmt.Errorf("failed to create work directory %v", workDir)
		}
	}

	modelChecksum := model.GetModel().GetGraphChecksum()

	if model.Model.IsArchive {
		baseURL := model.Model.BaseUrl
		str, err := downloadmanager.DownloadInto(baseURL, workDir, downloadmanager.MD5Sum(modelChecksum))
		if err != nil {
			return fmt.Errorf("failed to download model archive from %v", model.Model.BaseUrl)
		}
		if str == "" {
			return fmt.Errorf("MD5Sum of %v is incorrect.", baseURL)
		}
	} else {
		baseURL := strings.TrimSuffix(model.GetModel().GetBaseUrl(), "/")
		baseURL = strings.TrimSpace(baseURL)
		if baseURL != "" {
			baseURL += "/"
		}

		if model.GetModel().GetGraphPath() != "" {
			graphPath := filepath.Join(workDir, filepath.Base(model.GetModel().GetGraphPath()))
			str, ok, err := downloadmanager.DownloadFile(baseURL+model.GetModel().GetGraphPath(), graphPath, downloadmanager.MD5Sum(modelChecksum))
			if err != nil {
				return fmt.Errorf("failed to download model from %v", baseURL+model.GetModel().GetGraphPath())
			}
			if !ok && str == "" {
				return fmt.Errorf("MD5Sum of %v is incorrect.", baseURL+model.GetModel().GetGraphPath())
			}
		}
		if model.GetModel().GetWeightsPath() != "" {
			weightsPath := filepath.Join(workDir, filepath.Base(model.GetModel().GetWeightsPath()))
			weightsChecksum := model.GetModel().GetWeightsChecksum()
			str, ok, err := downloadmanager.DownloadFile(baseURL+model.GetModel().GetWeightsPath(), weightsPath, downloadmanager.MD5Sum(weightsChecksum))
			if err != nil {
				return fmt.Errorf("failed to download weight from %v", baseURL+model.GetModel().GetWeightsPath())
			}
			if !ok && str == "" {
				return fmt.Errorf("MD5Sum of %v is incorrect.", baseURL+model.GetModel().GetWeightsPath())
			}
		}
	}

	if model.GetModel().GetFeaturesPath() != "" {
		featuresPath := filepath.Join(workDir, filepath.Base(model.GetModel().GetFeaturesPath()))
		str, ok, err := downloadmanager.DownloadFile(model.GetModel().GetFeaturesPath(), featuresPath, downloadmanager.MD5Sum(model.GetModel().GetFeaturesChecksum()))
		if err != nil {
			return fmt.Errorf("failed to download features from %v", model.GetModel().GetFeaturesPath())
		}
		if !ok && str == "" {
			return fmt.Errorf("MD5Sum of %v is incorrect.", model.GetModel().GetFeaturesPath())
		}
	}

	// Write manifests
	manifestPath := filepath.Join(workDir, dlframework.CleanString(model.GetName()+"_"+model.GetVersion())+".yml")
	if err := ioutil.WriteFile(manifestPath, bts, 0644); err != nil {
		return fmt.Errorf("failed to write manifests to %v", manifestPath)
	}

	log.WithField("name", model.GetName()).WithField("version", model.GetVersion()).Info("downloaded ", model.GetName()+"_"+model.GetVersion())

	return nil
}
